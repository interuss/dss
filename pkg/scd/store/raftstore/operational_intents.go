package raftstore

import (
	"context"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) GetOperationalIntent(_ context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) DeleteOperationalIntent(_ context.Context, id dssmodels.ID) error {
	// TODO: implement
	return nil
}

func (r *repo) UpsertOperationalIntent(_ context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) SearchOperationalIntents(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) GetDependentOperationalIntents(_ context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) ListExpiredOperationalIntents(_ context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	// TODO: implement
	return nil, nil
}
