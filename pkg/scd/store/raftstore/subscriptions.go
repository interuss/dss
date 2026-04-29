package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) SearchSubscriptions(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpsertSubscription(_ context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) DeleteSubscription(_ context.Context, id dssmodels.ID) error {
	// TODO: implement
	return nil
}

func (r *repo) IncrementNotificationIndices(_ context.Context, subscriptionIds []dssmodels.ID) ([]int, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) LockSubscriptionsOnCells(_ context.Context, cells s2.CellUnion, subscriptionIds []dssmodels.ID, startTime *time.Time, endTime *time.Time) error {
	// TODO: implement
	return nil
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}
