package memstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for memory-based storage.
type repo struct {
	state state
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

func cloneCells(cells s2.CellUnion) s2.CellUnion {
	if cells == nil {
		return nil
	}
	return append(s2.CellUnion(nil), cells...)
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}

func cloneFloat32(f *float32) *float32 {
	if f == nil {
		return nil
	}
	v := *f
	return &v
}
