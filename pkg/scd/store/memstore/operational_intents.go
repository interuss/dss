package memstore

import (
	"context"
	"errors"
	"slices"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/memstore/utils"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

// toModel rebuilds the OperationalIntent model without its UssAvailability,
// which is attached separately (see buildOperationalIntents).
func (rec *operationalIntentRecord) toModel() *scdmodels.OperationalIntent {
	// If the managing USS has requested a specific OVN it is persisted, otherwise
	// a default DSS-generated OVN based on the last update time is used.
	var ovn scdmodels.OVN
	if rec.USSRequestedOVN != "" {
		ovn = scdmodels.OVN(rec.USSRequestedOVN)
	} else {
		ovn = scdmodels.NewOVNFromTime(rec.UpdatedAt, rec.ID.String())
	}
	return &scdmodels.OperationalIntent{
		ID:             rec.ID,
		Manager:        rec.Manager,
		Version:        rec.Version,
		State:          rec.State,
		OVN:            ovn,
		PastOVNs:       slices.Clone(rec.PastOVNs),
		USSBaseURL:     rec.USSBaseURL,
		SubscriptionID: utils.ClonePtr(rec.SubscriptionID),
		CellsVolume4D:  rec.CellsVolume4D.Clone(),
	}
}

// buildOperationalIntents converts records to models and attaches the
// UssAvailability of each managing USS
func (r *repo) buildOperationalIntents(ctx context.Context, recs []*operationalIntentRecord) ([]*scdmodels.OperationalIntent, error) {
	ussAvailabilities := map[dssmodels.Manager]scdmodels.UssAvailabilityState{}
	payload := make([]*scdmodels.OperationalIntent, 0, len(recs))
	for _, rec := range recs {
		o := rec.toModel()
		ussAvailabilities[o.Manager] = scdmodels.UssAvailabilityStateUnknown
		payload = append(payload, o)
	}

	for manager := range ussAvailabilities {
		ussAvailability, err := r.GetUssAvailability(ctx, manager)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, stacktrace.Propagate(err, "Error getting USS availability of %s", manager)
		}
		if ussAvailability != nil {
			ussAvailabilities[manager] = ussAvailability.Availability
		}
	}

	for _, op := range payload {
		op.UssAvailability = ussAvailabilities[op.Manager]
	}
	return payload, nil
}

func (r *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	rec, ok := r.state.OperationalIntents[id]
	if !ok {
		return nil, nil
	}
	built, err := r.buildOperationalIntents(ctx, []*operationalIntentRecord{rec})
	if err != nil {
		return nil, err
	}
	return built[0], nil
}

func (r *repo) DeleteOperationalIntent(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.OperationalIntents[id]; !ok {
		return stacktrace.NewError("Could not delete Operation that does not exist")
	}
	delete(r.state.OperationalIntents, id)
	return nil
}

func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	// An empty OVN means the DSS generates it; it is persisted as NULL in the
	// sqlstore (represented here by an empty USSRequestedOVN).
	var ussRequestedOVN string
	if operation.OVN != "" {
		ussRequestedOVN = operation.OVN.String()
	}

	now := timestamp.MustFromContext(ctx)

	rec := &operationalIntentRecord{
		ID:              operation.ID,
		Manager:         operation.Manager,
		Version:         operation.Version,
		State:           operation.State,
		USSBaseURL:      operation.USSBaseURL,
		SubscriptionID:  utils.ClonePtr(operation.SubscriptionID),
		USSRequestedOVN: ussRequestedOVN,
		PastOVNs:        slices.Clone(operation.PastOVNs),
		UpdatedAt:       now,
		CellsVolume4D:   operation.Clone(),
	}
	r.state.OperationalIntents[operation.ID] = rec

	built, err := r.buildOperationalIntents(ctx, []*operationalIntentRecord{rec})
	if err != nil {
		return nil, err
	}
	return built[0], nil
}

func (r *repo) SearchOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.OperationalIntent, error) {
	if len(cellsVolume.Cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
	}

	want := cellSet(cellsVolume.Cells)
	var matched []*operationalIntentRecord
	for _, rec := range r.state.OperationalIntents {
		if !overlaps(rec.CellsVolume4D.Cells, want) {
			continue
		}
		// COALESCE(altitude_upper >= $2, true) with $2 = cellsVolume.AltitudeLo
		if rec.CellsVolume4D.AltitudeHi != nil && cellsVolume.AltitudeLo != nil && *rec.CellsVolume4D.AltitudeHi < *cellsVolume.AltitudeLo {
			continue
		}
		// COALESCE(altitude_lower <= $3, true) with $3 = cellsVolume.AltitudeHi
		if rec.CellsVolume4D.AltitudeLo != nil && cellsVolume.AltitudeHi != nil && *rec.CellsVolume4D.AltitudeLo > *cellsVolume.AltitudeHi {
			continue
		}
		if !overlapsTime(rec.CellsVolume4D.StartTime, rec.CellsVolume4D.EndTime, cellsVolume) {
			continue
		}
		matched = append(matched, rec)

		if len(matched) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return r.buildOperationalIntents(ctx, matched)
}

func (r *repo) GetDependentOperationalIntents(_ context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	var dependentOps []dssmodels.ID
	for _, rec := range r.state.OperationalIntents {
		if rec.SubscriptionID != nil && *rec.SubscriptionID == subscriptionID {
			dependentOps = append(dependentOps, rec.ID)
		}
	}
	return dependentOps, nil
}

// TODO: use `limit` once evict is implemented for the raftstore.
func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time, _ int) ([]*scdmodels.OperationalIntent, error) {
	return r.buildOperationalIntents(ctx, listExpired(r.state.OperationalIntents, threshold, dssmodels.MaxResultLimit))
}

func (r *repo) DeleteExpiredOperationalIntents(_ context.Context, threshold time.Time, limit int) ([]*scdmodels.OperationalIntent, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredOperationalIntents not implemented for memstore")
}

func (r *repo) CountOperationalIntents(_ context.Context) (int64, error) {
	return int64(len(r.state.OperationalIntents)), nil
}
