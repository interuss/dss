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
	searchSubscriptions                               consensus.RequestType = "searchSubscriptions"
	getSubscription                                   consensus.RequestType = "getSubscription"
	upsertSubscription                                consensus.RequestType = "upsertSubscription"
	deleteSubscription                                consensus.RequestType = "deleteSubscription"
	incrementNotificationIndicesForOperationalIntents consensus.RequestType = "incrementNotificationIndicesForOperationalIntents"
	incrementNotificationIndicesForConstraints        consensus.RequestType = "incrementNotificationIndicesForConstraints"
	listExpiredSubscriptions                          consensus.RequestType = "listExpiredSubscriptions"
	countSubscriptions                                consensus.RequestType = "countSubscriptions"
)

func (r *repo) SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(v4d)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, searchSubscriptions, buf, true)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*scdmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getSubscription, buf, true)
	if err != nil {
		return nil, err
	}
	if sub, ok := result.(*scdmodels.Subscription); ok {
		return sub, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	buf, err := json.Marshal(sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, upsertSubscription, buf, false)
	if err != nil {
		return nil, err
	}
	if upserted, ok := result.(*scdmodels.Subscription); ok {
		return upserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) DeleteSubscription(ctx context.Context, id dssmodels.ID) error {
	buf, err := json.Marshal(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleClientRequest(ctx, deleteSubscription, buf, false)
	return err
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(ctx, incrementNotificationIndicesForOperationalIntents, v4d)
}

func (r *repo) IncrementNotificationIndicesForConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(ctx, incrementNotificationIndicesForConstraints, v4d)
}

func (r *repo) incrementNotificationIndices(ctx context.Context, requestType consensus.RequestType, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(v4d)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, requestType, buf, false)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*scdmodels.Subscription); ok {
		return subscriptions, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

// LockSubscriptionsOnCells is a no-op in the raftstore implementation
func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	return nil
}

func (r *repo) ListExpiredSubscriptions(ctx context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, listExpiredSubscriptions, buf, true)
	if err != nil {
		return nil, err
	}
	if subscriptions, ok := result.([]*scdmodels.Subscription); ok {
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
	case searchSubscriptions:
		var v4d dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptions)
		}
		return r.Store.GetRepo().SearchSubscriptions(ctx, &v4d)

	case getSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getSubscription)
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case upsertSubscription:
		var sub scdmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertSubscription)
		}
		return r.Store.GetRepo().UpsertSubscription(ctx, &sub)

	case deleteSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
		}
		return nil, r.Store.GetRepo().DeleteSubscription(ctx, id)

	case incrementNotificationIndicesForOperationalIntents:
		var v4d dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", incrementNotificationIndicesForOperationalIntents)
		}
		return r.Store.GetRepo().IncrementNotificationIndicesForOperationalIntents(ctx, &v4d)

	case incrementNotificationIndicesForConstraints:
		var v4d dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", incrementNotificationIndicesForConstraints)
		}
		return r.Store.GetRepo().IncrementNotificationIndicesForConstraints(ctx, &v4d)

	case listExpiredSubscriptions:
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, threshold)

	case countSubscriptions:
		return r.Store.GetRepo().CountSubscriptions(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized subscription request type: %s", proposal.RequestType)
	}
}
