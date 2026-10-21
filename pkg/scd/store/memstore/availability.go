package memstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/dss/pkg/timestamp"
)

func (rec *availabilityRecord) toModel() *scdmodels.UssAvailabilityStatus {
	return &scdmodels.UssAvailabilityStatus{
		Uss:          rec.Uss,
		Availability: rec.Availability,
		Version:      scdmodels.NewOVNFromTime(rec.UpdatedAt, rec.Uss.String()),
	}
}

func (r *repo) GetUssAvailability(_ context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	rec, ok := r.state.Availabilities[id]
	if !ok {
		return nil, repos.ErrNotFound
	}
	return rec.toModel(), nil
}

func (r *repo) UpsertUssAvailability(ctx context.Context, s *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	now := timestamp.MustFromContext(ctx)

	rec := &availabilityRecord{
		Uss:          s.Uss,
		Availability: s.Availability,
		UpdatedAt:    now,
	}
	r.state.Availabilities[s.Uss] = rec
	return rec.toModel(), nil
}
