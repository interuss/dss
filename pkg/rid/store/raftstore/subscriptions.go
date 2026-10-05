package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
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
	listExpiredSubscriptions           consensus.RequestType[[]dssmodels.ID]            = "listExpiredSubscriptions"
	countSubscriptions                 consensus.RequestType[int64]                     = "countSubscriptions"
)

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, getSubscription, buf)
}

func (r *repo) DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, deleteSubscription, buf)
}

func (r *repo) InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, insertSubscription, buf)
}

func (r *repo) UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, updateSubscription, buf)
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, searchSubscriptions, buf)
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, searchSubscriptionsByOwner, buf)
}

func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, updateNotificationIdxsInCells, buf)
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, maxSubscriptionCountInCellsByOwner, buf)
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, writer string, threshold time.Time) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, listExpiredSubscriptions, buf)
}

func (r *repo) DeleteExpiredSubscriptions(_ context.Context, writer string, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredSubscriptions not implemented for raftstore")
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadRequest(ctx, countSubscriptions, nil)
}
