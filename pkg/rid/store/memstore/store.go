package memstore

import (
	"context"
	"slices"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/memstore"
	"github.com/interuss/dss/pkg/memstore/utils"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for memory-based storage.
type repo struct {
	state      state
	checkpoint state
}

// state is the serializable in-memory state.
type state struct {
	// ISAs holds the stored ISAs keyed by ID.
	ISAs map[dssmodels.ID]*isaRecord
	// Subscriptions holds the stored subscriptions keyed by ID.
	Subscriptions map[dssmodels.ID]*subscriptionRecord
}

// isaRecord is the gob-serializable representation of an ISA. It intentionally
// stores only primitive fields: the model's Version is never persisted, it is
// derived from UpdatedAt on read.
type isaRecord struct {
	ID         dssmodels.ID
	URL        string
	Owner      dssmodels.Owner
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	AltitudeHi *float32
	AltitudeLo *float32
	Writer     string
	UpdatedAt  time.Time
}

// subscriptionRecord is the gob-serializable representation of a Subscription.
type subscriptionRecord struct {
	ID                dssmodels.ID
	URL               string
	NotificationIndex int
	Owner             dssmodels.Owner
	Cells             s2.CellUnion
	StartTime         *time.Time
	EndTime           *time.Time
	AltitudeHi        *float32
	AltitudeLo        *float32
	Writer            string
	UpdatedAt         time.Time
}

func newRepo() *repo {
	r := &repo{}
	r.resetState()
	return r
}

func (r *repo) resetState() {

	r.state = state{
		ISAs:          map[dssmodels.ID]*isaRecord{},
		Subscriptions: map[dssmodels.ID]*subscriptionRecord{},
	}

	r.Checkpoint()
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "rid", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }

// validateWriteData validate constraints on an ISA
func validateWriteData(cells s2.CellUnion, start, end *time.Time) error {
	if len(cells) == 0 {
		return stacktrace.NewError("At least one cell must be provided")
	}
	for _, c := range cells {
		if err := geo.ValidateCell(c); err != nil {
			return stacktrace.Propagate(err, "Error validating cell")
		}
	}
	if start != nil && end != nil && !start.Before(*end) {
		return stacktrace.NewError("Start time must be strictly before end time")
	}
	return nil
}

// cellSet builds a lookup set from a cell union.
func cellSet(cells s2.CellUnion) map[s2.CellID]struct{} {
	set := make(map[s2.CellID]struct{}, len(cells))
	for _, c := range cells {
		set[c] = struct{}{}
	}
	return set
}

// overlaps reports whether any cell is present in set (equivalent to the SQL
// "cells && $x" array-overlap operator).
func overlaps(cells s2.CellUnion, set map[s2.CellID]struct{}) bool {
	for _, c := range cells {
		if _, ok := set[c]; ok {
			return true
		}
	}
	return false
}

type versionedRecord interface {
	version() *dssmodels.Version
}

func (rec *isaRecord) version() *dssmodels.Version {
	return dssmodels.VersionFromTime(rec.UpdatedAt)
}

func (rec *subscriptionRecord) version() *dssmodels.Version {
	return dssmodels.VersionFromTime(rec.UpdatedAt)
}

func findForWrite[R versionedRecord](store map[dssmodels.ID]R, id dssmodels.ID, want *dssmodels.Version) (R, bool) {
	rec, ok := store[id]
	if !ok || !rec.version().Matches(want) {
		var zero R
		return zero, false
	}
	return rec, true
}

type expiringRecord[M any] interface {
	endTime() *time.Time
	writerName() string
	toModel() *M
}

func (rec *isaRecord) endTime() *time.Time { return rec.EndTime } //nolint:unused

func (rec *isaRecord) writerName() string { return rec.Writer } //nolint:unused

func (rec *subscriptionRecord) endTime() *time.Time { return rec.EndTime } //nolint:unused

func (rec *subscriptionRecord) writerName() string { return rec.Writer } //nolint:unused

// listExpired returns the records whose end time is at or before threshold and
// whose writer matches. A limit of 0 means unlimited.
func listExpired[M any, R expiringRecord[M]](store map[dssmodels.ID]R, writer string, threshold time.Time, limit int) []*M {
	var out []*M
	for _, rec := range store {
		if t := rec.endTime(); t == nil || t.After(threshold) { // TODO: Don't allow endtime to be null, see #1492
			continue
		}
		if rec.writerName() != writer {
			continue
		}
		out = append(out, rec.toModel())

		if limit > 0 && len(out) > limit { // This mimics sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out
}

func (rec *isaRecord) clone() *isaRecord {
	cp := *rec
	cp.Cells = slices.Clone(rec.Cells)
	cp.StartTime = utils.ClonePtr(rec.StartTime)
	cp.EndTime = utils.ClonePtr(rec.EndTime)
	cp.AltitudeHi = utils.ClonePtr(rec.AltitudeHi)
	cp.AltitudeLo = utils.ClonePtr(rec.AltitudeLo)
	return &cp
}

func (rec *subscriptionRecord) clone() *subscriptionRecord {
	cp := *rec
	cp.Cells = slices.Clone(rec.Cells)
	cp.StartTime = utils.ClonePtr(rec.StartTime)
	cp.EndTime = utils.ClonePtr(rec.EndTime)
	cp.AltitudeHi = utils.ClonePtr(rec.AltitudeHi)
	cp.AltitudeLo = utils.ClonePtr(rec.AltitudeLo)
	return &cp
}

// clone returns a deep copy of s. May be optimzed in speed by not cloning everything, as long
// rest of the package don't mutate fields, iff speed of this function is important.
func (s state) clone() state {
	isas := make(map[dssmodels.ID]*isaRecord, len(s.ISAs))
	for id, rec := range s.ISAs {
		isas[id] = rec.clone()
	}
	subs := make(map[dssmodels.ID]*subscriptionRecord, len(s.Subscriptions))
	for id, rec := range s.Subscriptions {
		subs[id] = rec.clone()
	}
	return state{ISAs: isas, Subscriptions: subs}
}

// Checkpoint ask the repo to store a quick, internal checkpoint with its current state.
// There is at most one check point, any existing checkpoint is overwritten
func (r *repo) Checkpoint() {
	r.checkpoint = r.state.clone()
}

// Restore replaces the current state with the latest checkpoint. May be called multiple time
// to restore the same checkpoint.
func (r *repo) Restore() {
	r.state = r.checkpoint.clone()
}
