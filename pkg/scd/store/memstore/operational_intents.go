package memstore

import (
	"context"
	"errors"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
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
		PastOVNs:       clonePastOVNs(rec.PastOVNs),
		StartTime:      cloneTime(rec.StartTime),
		EndTime:        cloneTime(rec.EndTime),
		USSBaseURL:     rec.USSBaseURL,
		SubscriptionID: cloneID(rec.SubscriptionID),
		AltitudeLower:  cloneFloat32(rec.AltitudeLower),
		AltitudeUpper:  cloneFloat32(rec.AltitudeUpper),
		Cells:          cloneCells(rec.Cells),
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

// GetOperationalIntent implements scd.repos.OperationalIntent.GetOperationalIntent.
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

// DeleteOperationalIntent implements scd.repos.OperationalIntent.DeleteOperationalIntent.
func (r *repo) DeleteOperationalIntent(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.OperationalIntents[id]; !ok {
		return stacktrace.NewError("Could not delete Operation that does not exist")
	}
	delete(r.state.OperationalIntents, id)
	return nil
}

// UpsertOperationalIntent implements scd.repos.OperationalIntent.UpsertOperationalIntent.
func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	// An empty OVN means the DSS generates it; it is persisted as NULL in the
	// sqlstore (represented here by an empty USSRequestedOVN).
	var ussRequestedOVN string
	if operation.OVN != "" {
		ussRequestedOVN = operation.OVN.String()
	}

	rec := &operationalIntentRecord{
		ID:              operation.ID,
		Manager:         operation.Manager,
		Version:         operation.Version,
		State:           operation.State,
		StartTime:       cloneTime(operation.StartTime),
		EndTime:         cloneTime(operation.EndTime),
		USSBaseURL:      operation.USSBaseURL,
		SubscriptionID:  cloneID(operation.SubscriptionID),
		AltitudeLower:   cloneFloat32(operation.AltitudeLower),
		AltitudeUpper:   cloneFloat32(operation.AltitudeUpper),
		Cells:           cloneCells(operation.Cells),
		USSRequestedOVN: ussRequestedOVN,
		PastOVNs:        clonePastOVNs(operation.PastOVNs),
		UpdatedAt:       timestamp.NowFromContext(ctx),
	}
	r.state.OperationalIntents[operation.ID] = rec

	built, err := r.buildOperationalIntents(ctx, []*operationalIntentRecord{rec})
	if err != nil {
		return nil, err
	}
	return built[0], nil
}

// SearchOperationalIntents implements scd.repos.OperationalIntent.SearchOperationalIntents.
func (r *repo) SearchOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	if v4d.SpatialVolume == nil || v4d.SpatialVolume.Footprint == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing geospatial footprint for query")
	}
	cells, err := v4d.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to calculate footprint covering")
	}
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
	}

	want := cellSet(cells)
	var matched []*operationalIntentRecord
	for _, rec := range r.state.OperationalIntents {
		if !overlaps(rec.Cells, want) {
			continue
		}
		// COALESCE(altitude_upper >= $2, true) with $2 = SpatialVolume.AltitudeLo
		if rec.AltitudeUpper != nil && v4d.SpatialVolume.AltitudeLo != nil && *rec.AltitudeUpper < *v4d.SpatialVolume.AltitudeLo {
			continue
		}
		// COALESCE(altitude_lower <= $3, true) with $3 = SpatialVolume.AltitudeHi
		if rec.AltitudeLower != nil && v4d.SpatialVolume.AltitudeHi != nil && *rec.AltitudeLower > *v4d.SpatialVolume.AltitudeHi {
			continue
		}
		// COALESCE(ends_at >= $4, true) with $4 = v4d.StartTime
		if rec.EndTime != nil && v4d.StartTime != nil && rec.EndTime.Before(*v4d.StartTime) {
			continue
		}
		// COALESCE(starts_at <= $5, true) with $5 = v4d.EndTime
		if rec.StartTime != nil && v4d.EndTime != nil && rec.StartTime.After(*v4d.EndTime) {
			continue
		}
		matched = append(matched, rec)
		if len(matched) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return r.buildOperationalIntents(ctx, matched)
}

// GetDependentOperationalIntents implements scd.repos.OperationalIntent.GetDependentOperationalIntents.
func (r *repo) GetDependentOperationalIntents(_ context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	var dependentOps []dssmodels.ID
	for _, rec := range r.state.OperationalIntents {
		if rec.SubscriptionID != nil && *rec.SubscriptionID == subscriptionID {
			dependentOps = append(dependentOps, rec.ID)
		}
	}
	return dependentOps, nil
}

// ListExpiredOperationalIntents implements scd.repos.OperationalIntent.ListExpiredOperationalIntents.
func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	var matched []*operationalIntentRecord
	for _, rec := range r.state.OperationalIntents {
		// (ends_at IS NOT NULL AND ends_at <= threshold) OR (ends_at IS NULL AND updated_at <= threshold)
		var expired bool
		if rec.EndTime != nil {
			expired = !rec.EndTime.After(threshold)
		} else {
			expired = !rec.UpdatedAt.After(threshold)
		}
		if !expired {
			continue
		}
		matched = append(matched, rec)
		if len(matched) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return r.buildOperationalIntents(ctx, matched)
}

func (r *repo) CountOperationalIntents(_ context.Context) (int64, error) {
	return int64(len(r.state.OperationalIntents)), nil
}
