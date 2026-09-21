package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

const (
	searchSubscriptions                               consensus.RequestType[[]*scdmodels.Subscription] = "searchSubscriptions"
	getSubscription                                   consensus.RequestType[*scdmodels.Subscription]   = "getSubscription"
	upsertSubscription                                consensus.RequestType[*scdmodels.Subscription]   = "upsertSubscription"
	deleteSubscription                                consensus.RequestType[any]                       = "deleteSubscription"
	incrementNotificationIndicesForOperationalIntents consensus.RequestType[[]*scdmodels.Subscription] = "incrementNotificationIndicesForOperationalIntents"
	incrementNotificationIndicesForConstraints        consensus.RequestType[[]*scdmodels.Subscription] = "incrementNotificationIndicesForConstraints"
	listExpiredSubscriptions                          consensus.RequestType[[]*scdmodels.Subscription] = "listExpiredSubscriptions"
	countSubscriptions                                consensus.RequestType[int64]                     = "countSubscriptions"
)

func (r *repo) SearchSubscriptions(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchSubscriptions, buf, true)
}

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getSubscription, buf, true)
}

func (r *repo) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, upsertSubscription, buf, false)
}

func (r *repo) DeleteSubscription(ctx context.Context, id dssmodels.ID) error {
	buf, err := json.Marshal(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleClientRequest(ctx, deleteSubscription, buf, false)
	return err
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, incrementNotificationIndicesForOperationalIntents, buf, false)
}

func (r *repo) IncrementNotificationIndicesForConstraints(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, incrementNotificationIndicesForConstraints, buf, false)
}

// LockSubscriptionsOnCells is a no-op in the raftstore implementation since raft requests are executed sequentially
func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	return nil
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, listExpiredSubscriptions, buf, true)
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.consensus.HandleClientRequest(ctx, countSubscriptions, nil, true)
}
