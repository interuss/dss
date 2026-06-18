package memstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/jackc/pgx/v5"
)

func (rec *availabilityRecord) toModel() *scdmodels.UssAvailabilityStatus {
	return &scdmodels.UssAvailabilityStatus{
		Uss:          rec.Uss,
		Availability: rec.Availability,
		Version:      scdmodels.NewOVNFromTime(rec.UpdatedAt, rec.Uss.String()),
	}
}

// GetUssAvailability implements scd.repos.UssAvailability.GetUssAvailability.
func (r *repo) GetUssAvailability(_ context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	rec, ok := r.state.Availabilities[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return rec.toModel(), nil
}

// UpsertUssAvailability implements scd.repos.UssAvailability.UpsertUssAvailability.
func (r *repo) UpsertUssAvailability(ctx context.Context, s *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	rec := &availabilityRecord{
		Uss:          s.Uss,
		Availability: s.Availability,
		UpdatedAt:    timestamp.NowFromContext(ctx),
	}
	r.state.Availabilities[s.Uss] = rec
	return rec.toModel(), nil
}
