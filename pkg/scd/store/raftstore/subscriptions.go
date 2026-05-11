package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchSubscriptions(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchSubscriptions not implemented for raftstore")
}

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetSubscription not implemented for raftstore")
}

func (r *repo) UpsertSubscription(_ context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpsertSubscription not implemented for raftstore")
}

func (r *repo) DeleteSubscription(_ context.Context, id dssmodels.ID) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteSubscription not implemented for raftstore")
}

func (r *repo) IncrementNotificationIndices(_ context.Context, subscriptionIds []dssmodels.ID) ([]int, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "IncrementNotificationIndices not implemented for raftstore")
}

func (r *repo) LockSubscriptionsOnCells(_ context.Context, cells s2.CellUnion, subscriptionIds []dssmodels.ID, startTime *time.Time, endTime *time.Time) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "LockSubscriptionsOnCells not implemented for raftstore")
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredSubscriptions not implemented for raftstore")
}
