package memstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for memory-based storage.
type repo struct {
	state state
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

func cloneID(id *dssmodels.ID) *dssmodels.ID {
	if id == nil {
		return nil
	}
	v := *id
	return &v
}

func clonePastOVNs(ovns []scdmodels.OVN) []scdmodels.OVN {
	if ovns == nil {
		return nil
	}
	return append([]scdmodels.OVN(nil), ovns...)
}

// clone returns a copy of s with independent maps and records. Cell slices,
// time pointers and OVN slices are shared, as they are never mutated in place.
func (s state) clone() state {
	constraints := make(map[dssmodels.ID]*constraintRecord, len(s.Constraints))
	for id, rec := range s.Constraints {
		cp := *rec
		constraints[id] = &cp
	}
	subs := make(map[dssmodels.ID]*subscriptionRecord, len(s.Subscriptions))
	for id, rec := range s.Subscriptions {
		cp := *rec
		subs[id] = &cp
	}
	ois := make(map[dssmodels.ID]*operationalIntentRecord, len(s.OperationalIntents))
	for id, rec := range s.OperationalIntents {
		cp := *rec
		ois[id] = &cp
	}
	avails := make(map[dssmodels.Manager]*availabilityRecord, len(s.Availabilities))
	for id, rec := range s.Availabilities {
		cp := *rec
		avails[id] = &cp
	}
	return state{
		Constraints:        constraints,
		Subscriptions:      subs,
		OperationalIntents: ois,
		Availabilities:     avails,
	}
}

// Checkpoint returns a fast, restorable in-memory copy of the current state.
// Unlike GetSnapshot it does not serialize, so it is cheap but only valid
// in-process.
func (r *repo) Checkpoint() any {
	return r.state.clone()
}

// Restore replaces the current state with a checkpoint previously returned by
// Checkpoint. The checkpoint is copied, so it stays reusable.
func (r *repo) Restore(cp any) error {
	s, ok := cp.(state)
	if !ok {
		return stacktrace.NewError("Invalid checkpoint type %T", cp)
	}
	r.state = s.clone()
	return nil
}
