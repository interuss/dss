package raftstore

import (
	"context"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) GetOperationalIntent(_ context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	panic("GetOperationalIntent not yet implemented in raft store")
}

func (r *repo) DeleteOperationalIntent(_ context.Context, id dssmodels.ID) error {
	panic("DeleteOperationalIntent not yet implemented in raft store")
}

func (r *repo) UpsertOperationalIntent(_ context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	panic("UpsertOperationalIntent not yet implemented in raft store")
}

func (r *repo) SearchOperationalIntents(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	panic("SearchOperationalIntents not yet implemented in raft store")
}

func (r *repo) GetDependentOperationalIntents(_ context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	panic("GetDependentOperationalIntents not yet implemented in raft store")
}

func (r *repo) ListExpiredOperationalIntents(_ context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	panic("ListExpiredOperationalIntents not yet implemented in raft store")
}
