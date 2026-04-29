package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) GetUssAvailability(_ context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpsertUssAvailability(_ context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	// TODO: implement
	return nil, nil
}
