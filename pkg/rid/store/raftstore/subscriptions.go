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
	getSubscription                    consensus.RequestType = "getSubscription"
	deleteSubscription                 consensus.RequestType = "deleteSubscription"
	insertSubscription                 consensus.RequestType = "insertSubscription"
	updateSubscription                 consensus.RequestType = "updateSubscription"
	searchSubscriptions                consensus.RequestType = "searchSubscriptions"
	searchSubscriptionsByOwner         consensus.RequestType = "searchSubscriptionsByOwner"
	updateNotificationIdxsInCells      consensus.RequestType = "updateNotificationIdxsInCells"
	maxSubscriptionCountInCellsByOwner consensus.RequestType = "maxSubscriptionCountInCellsByOwner"
	listExpiredSubscriptions           consensus.RequestType = "listExpiredSubscriptions"
	countSubscriptions                 consensus.RequestType = "countSubscriptions"
)

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getSubscription, buf, true)
	if err != nil {
		return nil, err
	}
	if sub, ok := result.(*ridmodels.Subscription); ok {
		return sub, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, deleteSubscription, buf, false)
	if err != nil {
		return nil, err
	}
	if deleted, ok := result.(*ridmodels.Subscription); ok {
		return deleted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, insertSubscription, buf, false)
	if err != nil {
		return nil, err
	}
	if inserted, ok := result.(*ridmodels.Subscription); ok {
		return inserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, updateSubscription, buf, false)
	if err != nil {
		return nil, err
	}
	if updated, ok := result.(*ridmodels.Subscription); ok {
		return updated, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return r.searchSubscriptions(ctx, searchSubscriptions, cells)
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, searchSubscriptionsByOwner, buf, true)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*ridmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) searchSubscriptions(ctx context.Context, requestType consensus.RequestType, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, requestType, buf, true)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*ridmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, updateNotificationIdxsInCells, buf, false)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*ridmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	buf, err := json.Marshal(cellsByOwnerPayload{Cells: cells, Owner: owner})
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, maxSubscriptionCountInCellsByOwner, buf, true)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int); ok {
		return count, nil
	}
	return 0, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, listExpiredSubscriptions, buf, true)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*ridmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	result, err := r.consensus.HandleClientRequest(ctx, countSubscriptions, nil, true)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int64); ok {
		return count, nil
	}
	return 0, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) applySubscription(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case getSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getSubscription)
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case deleteSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
		}
		return r.Store.GetRepo().DeleteSubscription(ctx, &sub)

	case insertSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertSubscription)
		}
		return r.Store.GetRepo().InsertSubscription(ctx, &sub)

	case updateSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateSubscription)
		}
		return r.Store.GetRepo().UpdateSubscription(ctx, &sub)

	case searchSubscriptions:
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptions)
		}
		return r.Store.GetRepo().SearchSubscriptions(ctx, cells)

	case searchSubscriptionsByOwner:
		var payload cellsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptionsByOwner)
		}
		return r.Store.GetRepo().SearchSubscriptionsByOwner(ctx, payload.Cells, payload.Owner)

	case updateNotificationIdxsInCells:
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateNotificationIdxsInCells)
		}
		return r.Store.GetRepo().UpdateNotificationIdxsInCells(ctx, cells)

	case maxSubscriptionCountInCellsByOwner:
		var payload cellsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", maxSubscriptionCountInCellsByOwner)
		}
		return r.Store.GetRepo().MaxSubscriptionCountInCellsByOwner(ctx, payload.Cells, payload.Owner)

	case listExpiredSubscriptions:
		var payload expiredPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, payload.Writer, payload.Threshold)

	case countSubscriptions:
		return r.Store.GetRepo().CountSubscriptions(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized subscription request type: %s", proposal.RequestType)
	}
}
