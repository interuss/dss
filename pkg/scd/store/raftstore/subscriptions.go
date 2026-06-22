package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(searchSubscriptions), v4d, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose searchSubscriptions")
	}

	if result == nil {
		return nil, nil
	}

	subs, ok := result.([]*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return subs, nil
}

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(getSubscription), id, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose getSubscription")
	}

	if result == nil {
		return nil, nil
	}

	sub, ok := result.(*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return sub, nil
}

func (r *repo) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(upsertSubscription), sub, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose upsertSubscription")
	}

	if result == nil {
		return nil, nil
	}

	upserted, ok := result.(*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return upserted, nil
}

func (r *repo) DeleteSubscription(ctx context.Context, id dssmodels.ID) error {
	_, err := r.consensus.ProposeValue(ctx, string(deleteSubscription), id, false)
	return err
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(incrementNotificationForOIs), v4d, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose incrementNotificationForOIs")
	}

	if result == nil {
		return nil, nil
	}

	subs, ok := result.([]*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return subs, nil
}

func (r *repo) IncrementNotificationIndicesForConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(incrementNotificationForConstraints), v4d, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose incrementNotificationForConstraints")
	}

	if result == nil {
		return nil, nil
	}

	subs, ok := result.([]*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return subs, nil
}

func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	// for the raftstore, LockSubscriptionsOnCells is a no-op
	return nil
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	result, err := r.consensus.ProposeValue(ctx, string(listExpiredSubscriptions), threshold, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose listExpiredSubscriptions")
	}

	if result == nil {
		return nil, nil
	}

	subs, ok := result.([]*scdmodels.Subscription)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return subs, nil
}

func (r *repo) CountSubscriptions(ctx context.Context) (int64, error) {
	result, err := r.consensus.ProposeValue(ctx, string(countSubscriptions), nil, true)
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to propose countSubscriptions")
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
