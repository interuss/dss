package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

const (
	getSubscription                    consensus.RequestType[*ridmodels.Subscription]   = "getSubscription"
	deleteSubscription                 consensus.RequestType[*ridmodels.Subscription]   = "deleteSubscription"
	insertSubscription                 consensus.RequestType[*ridmodels.Subscription]   = "insertSubscription"
	updateSubscription                 consensus.RequestType[*ridmodels.Subscription]   = "updateSubscription"
	searchSubscriptions                consensus.RequestType[[]*ridmodels.Subscription] = "searchSubscriptions"
	searchSubscriptionsByOwner         consensus.RequestType[[]*ridmodels.Subscription] = "searchSubscriptionsByOwner"
	updateNotificationIdxsInCells      consensus.RequestType[[]*ridmodels.Subscription] = "updateNotificationIdxsInCells"
	maxSubscriptionCountInCellsByOwner consensus.RequestType[int]                       = "maxSubscriptionCountInCellsByOwner"
	listExpiredSubscriptions           consensus.RequestType[[]*ridmodels.Subscription] = "listExpiredSubscriptions"
	countSubscriptions                 consensus.RequestType[int64]                     = "countSubscriptions"
)

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getSubscription, buf, true)
}

func (r *repo) DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, deleteSubscription, buf, false)
}

func (r *repo) InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, insertSubscription, buf, false)
}

func (r *repo) UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, updateSubscription, buf, false)
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchSubscriptions, buf, true)
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchSubscriptionsByOwner, buf, true)
}

func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, updateNotificationIdxsInCells, buf, false)
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, maxSubscriptionCountInCellsByOwner, buf, true)
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, listExpiredSubscriptions, buf, true)
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.consensus.HandleClientRequest(ctx, countSubscriptions, nil, true)
}
