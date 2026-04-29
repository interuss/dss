package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) SearchSubscriptions(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	panic("SearchSubscriptions not yet implemented in raft store")
}

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	panic("GetSubscription not yet implemented in raft store")
}

func (r *repo) UpsertSubscription(_ context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	panic("UpsertSubscription not yet implemented in raft store")
}

func (r *repo) DeleteSubscription(_ context.Context, id dssmodels.ID) error {
	panic("DeleteSubscription not yet implemented in raft store")
}

func (r *repo) IncrementNotificationIndices(_ context.Context, subscriptionIds []dssmodels.ID) ([]int, error) {
	panic("IncrementNotificationIndices not yet implemented in raft store")
}

func (r *repo) LockSubscriptionsOnCells(_ context.Context, cells s2.CellUnion, subscriptionIds []dssmodels.ID, startTime *time.Time, endTime *time.Time) error {
	panic("LockSubscriptionsOnCells not yet implemented in raft store")
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	panic("ListExpiredSubscriptions not yet implemented in raft store")
}
