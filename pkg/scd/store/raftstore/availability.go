package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) GetUssAvailability(_ context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	panic("GetUssAvailability not yet implemented in raft store")
}

func (r *repo) UpsertUssAvailability(_ context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	panic("UpsertUssAvailability not yet implemented in raft store")
}
