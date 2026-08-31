package memstore

import (
	"context"
	"slices"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/memstore"
	"github.com/interuss/dss/pkg/memstore/utils"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for memory-based storage.
type repo struct {
	state      state
	checkpoint state
}

// state is the serializable in-memory state.
type state struct {
	// Constraints holds the stored constraints keyed by ID.
	Constraints map[dssmodels.ID]*constraintRecord
	// Subscriptions holds the stored subscriptions keyed by ID.
	Subscriptions map[dssmodels.ID]*subscriptionRecord
	// OperationalIntents holds the stored operational intents keyed by ID.
	OperationalIntents map[dssmodels.ID]*operationalIntentRecord
	// Availabilities holds the stored USS availabilities keyed by USS Manager.
	Availabilities map[dssmodels.Manager]*availabilityRecord
}

// constraintRecord is the gob-serializable representation of a Constraint. The
// model's OVN is never persisted: it is derived from UpdatedAt on read
type constraintRecord struct {
	ID            dssmodels.ID
	Manager       dssmodels.Manager
	Version       scdmodels.VersionNumber
	StartTime     *time.Time
	EndTime       *time.Time
	USSBaseURL    string
	AltitudeLower *float32
	AltitudeUpper *float32
	Cells         s2.CellUnion
	UpdatedAt     time.Time
}

// subscriptionRecord is the gob-serializable representation of a Subscription.
// The sqlstore stores the version column but always writes 0 and discards it on
// read (the model Version is derived from UpdatedAt), so it is not kept here.
type subscriptionRecord struct {
	ID                          dssmodels.ID
	Manager                     dssmodels.Manager
	NotificationIndex           int
	USSBaseURL                  string
	NotifyForOperationalIntents bool
	NotifyForConstraints        bool
	ImplicitSubscription        bool
	StartTime                   *time.Time
	EndTime                     *time.Time
	Cells                       s2.CellUnion
	UpdatedAt                   time.Time
}

// operationalIntentRecord is the gob-serializable representation of an
// OperationalIntent. USSRequestedOVN is empty when the OVN is DSS-generated.
type operationalIntentRecord struct {
	ID              dssmodels.ID
	Manager         dssmodels.Manager
	Version         scdmodels.VersionNumber
	State           scdmodels.OperationalIntentState
	StartTime       *time.Time
	EndTime         *time.Time
	USSBaseURL      string
	SubscriptionID  *dssmodels.ID
	AltitudeLower   *float32
	AltitudeUpper   *float32
	Cells           s2.CellUnion
	USSRequestedOVN string
	PastOVNs        []scdmodels.OVN
	UpdatedAt       time.Time
}

// availabilityRecord is the gob-serializable representation of a
// UssAvailabilityStatus. The model's Version is derived from UpdatedAt on read.
type availabilityRecord struct {
	Uss          dssmodels.Manager
	Availability scdmodels.UssAvailabilityState
	UpdatedAt    time.Time
}

func newRepo() *repo {
	r := &repo{}
	r.resetState()
	return r
}

func (r *repo) resetState() {
	r.state = state{
		Constraints:        map[dssmodels.ID]*constraintRecord{},
		Subscriptions:      map[dssmodels.ID]*subscriptionRecord{},
		OperationalIntents: map[dssmodels.ID]*operationalIntentRecord{},
		Availabilities:     map[dssmodels.Manager]*availabilityRecord{},
	}
	r.Checkpoint()
}

func Init(ctx context.Context, logger *zap.Logger) (*memstore.Store[repos.Repository], error) {
	return memstore.Init(ctx, logger, "scd", newRepo())
}

func (r *repo) GetRepo() repos.Repository { return r }

// cellSet builds a lookup set from a cell union.
func cellSet(cells s2.CellUnion) map[s2.CellID]struct{} {
	set := make(map[s2.CellID]struct{}, len(cells))
	for _, c := range cells {
		set[c] = struct{}{}
	}
	return set
}

// coveringSet builds a lookup set from the spatial covering of a volume.
func coveringSet(v4d *dssmodels.Volume4D) (map[s2.CellID]struct{}, error) {
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}
	return cellSet(cells), nil
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

// overlapsTime reports whether the [start, end] interval of a record intersects the one of v4d
// (equivalent to the SQL "COALESCE(starts_at <= $end, true) AND COALESCE(ends_at >= $start, true)").
func overlapsTime(start, end *time.Time, v4d *dssmodels.Volume4D) bool {
	if start != nil && v4d.EndTime != nil && start.After(*v4d.EndTime) { // TODO: Don't allow startup to be null, see #1492
		return false
	}
	if end != nil && v4d.StartTime != nil && end.Before(*v4d.StartTime) { // TODO: Don't allow endtime to be null, see #1492
		return false
	}
	return true
}

type expiringRecord interface {
	endTime() *time.Time
	lastUpdate() time.Time
}

func (rec *subscriptionRecord) endTime() *time.Time { return rec.EndTime }

func (rec *subscriptionRecord) lastUpdate() time.Time { return rec.UpdatedAt }

func (rec *operationalIntentRecord) endTime() *time.Time { return rec.EndTime }

func (rec *operationalIntentRecord) lastUpdate() time.Time { return rec.UpdatedAt }

// listExpired returns the records whose end time is at or before threshold, falling back on the
// last update time when the end time is unknown. A limit of 0 means unlimited.
func listExpired[R expiringRecord](store map[dssmodels.ID]R, threshold time.Time, limit int) []R {
	var out []R
	for _, rec := range store {
		// (ends_at IS NOT NULL AND ends_at <= threshold) OR (ends_at IS NULL AND updated_at <= threshold)
		if t := rec.endTime(); t != nil { // TODO: Don't allow endtime to be null, see #1492
			if t.After(threshold) {
				continue
			}
		} else if rec.lastUpdate().After(threshold) {
			continue
		}
		out = append(out, rec)

		if limit > 0 && len(out) >= limit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return out
}

func (rec *constraintRecord) clone() *constraintRecord {
	cp := *rec
	cp.Cells = slices.Clone(rec.Cells)
	cp.StartTime = utils.ClonePtr(rec.StartTime)
	cp.EndTime = utils.ClonePtr(rec.EndTime)
	cp.AltitudeLower = utils.ClonePtr(rec.AltitudeLower)
	cp.AltitudeUpper = utils.ClonePtr(rec.AltitudeUpper)
	return &cp
}

func (rec *subscriptionRecord) clone() *subscriptionRecord {
	cp := *rec
	cp.Cells = slices.Clone(rec.Cells)
	cp.StartTime = utils.ClonePtr(rec.StartTime)
	cp.EndTime = utils.ClonePtr(rec.EndTime)
	return &cp
}

func (rec *operationalIntentRecord) clone() *operationalIntentRecord {
	cp := *rec
	cp.Cells = slices.Clone(rec.Cells)
	cp.PastOVNs = slices.Clone(rec.PastOVNs)
	cp.StartTime = utils.ClonePtr(rec.StartTime)
	cp.EndTime = utils.ClonePtr(rec.EndTime)
	cp.SubscriptionID = utils.ClonePtr(rec.SubscriptionID)
	cp.AltitudeLower = utils.ClonePtr(rec.AltitudeLower)
	cp.AltitudeUpper = utils.ClonePtr(rec.AltitudeUpper)
	return &cp
}

func (rec *availabilityRecord) clone() *availabilityRecord {
	cp := *rec
	return &cp
}

// clone returns a deep copy of s. May be optimzed in speed by not cloning everything, as long
// rest of the package don't mutate fields, iff speed of this function is important.
func (s state) clone() state {
	constraints := make(map[dssmodels.ID]*constraintRecord, len(s.Constraints))
	for id, rec := range s.Constraints {
		constraints[id] = rec.clone()
	}
	subs := make(map[dssmodels.ID]*subscriptionRecord, len(s.Subscriptions))
	for id, rec := range s.Subscriptions {
		subs[id] = rec.clone()
	}
	ois := make(map[dssmodels.ID]*operationalIntentRecord, len(s.OperationalIntents))
	for id, rec := range s.OperationalIntents {
		ois[id] = rec.clone()
	}
	avails := make(map[dssmodels.Manager]*availabilityRecord, len(s.Availabilities))
	for uss, rec := range s.Availabilities {
		avails[uss] = rec.clone()
	}
	return state{
		Constraints:        constraints,
		Subscriptions:      subs,
		OperationalIntents: ois,
		Availabilities:     avails,
	}
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
