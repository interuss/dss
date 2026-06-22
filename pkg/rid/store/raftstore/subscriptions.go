package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, getSubscription, id, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	sub, ok := result.(*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return sub, nil
}

func (r *repo) DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, deleteSubscription, sub, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.(*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

func (r *repo) InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, insertSubscription, sub, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.(*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

func (r *repo) UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, updateSubscription, sub, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.(*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, searchSubscriptions, cells, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.([]*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

type searchSubscriptionsByOwnerPayload struct {
	Cells s2.CellUnion    `json:"cells"`
	Owner dssmodels.Owner `json:"owner"`
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, searchSubscriptionsByOwner, &searchSubscriptionsByOwnerPayload{Cells: cells, Owner: owner}, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.([]*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, updateNotificationIdxsInCells, cells, false)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.([]*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

type maxSubscriptionCountInCellsByOwnerPayload struct {
	Cells s2.CellUnion    `json:"cells"`
	Owner dssmodels.Owner `json:"owner"`
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	result, err := r.consensus.ProposeValue(ctx, maxSubscriptionCountInCellsByOwner, &maxSubscriptionCountInCellsByOwnerPayload{Cells: cells, Owner: owner}, true)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil
	}

	count, ok := result.(int)
	if !ok {
		return 0, stacktrace.NewError("invalid result type: %T", result)
	}

	return count, nil
}

type listExpiredSubscriptionsPayload struct {
	Writer    string    `json:"writer"`
	Threshold time.Time `json:"threshold"`
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, listExpiredSubscriptions, &listExpiredSubscriptionsPayload{Writer: writer, Threshold: threshold}, true)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	out, ok := result.([]*ridmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return out, nil
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	result, err := r.consensus.ProposeValue(ctx, countSubscriptions, nil, true)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil
	}

	count, ok := result.(int64)
	if !ok {
		return 0, stacktrace.NewError("invalid result type: %T", result)
	}

	return count, nil
}
