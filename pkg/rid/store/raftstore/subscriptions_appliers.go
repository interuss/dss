package raftstore

import (
	"context"
	"encoding/json"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
)

func (r *repo) insertSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ridmodels.Subscription, error) {
	var payload *ridmodels.Subscription
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertSubscription)
	}

	old, err := mem.GetSubscription(ctx, payload.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "error getting Subscription from repo")
	}
	if old != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", payload.ID)
	}

	count, err := mem.MaxSubscriptionCountInCellsByOwner(ctx, payload.Cells, payload.Owner)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to fetch subscription count, rejecting request")
	}
	if count >= ridmodels.MaxSubscriptionsPerArea {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.Exhausted, "too many existing subscriptions in this area already"),
			"%s had %d subscriptions in the area", payload.Owner, count)
	}

	checkpoint := r.memStore.Checkpoint()
	ret, err := mem.InsertSubscription(ctx, payload)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error inserting Subscription")
	}
	return ret, nil
}

func (r *repo) updateSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ridmodels.Subscription, error) {
	var payload *ridmodels.Subscription
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateSubscription)
	}

	old, err := mem.GetSubscription(ctx, payload.ID)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", payload.ID.String())
	case !payload.Version.Matches(old.Version):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", payload.Version),
			"Subscription currently at version %s but client specified %s", old.Version, payload.Version)
	case old.Owner != payload.Owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to update", old.Owner, payload.Owner)
	}
	if err := payload.AdjustTimeRange(proposal.Timestamp, old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	count, err := mem.MaxSubscriptionCountInCellsByOwner(ctx, payload.Cells, payload.Owner)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to fetch subscription count, rejecting request")
	}
	if count >= ridmodels.MaxSubscriptionsPerArea {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.Exhausted, "Too many existing subscriptions in this area already"),
			"%s had %d subscriptions in the area", payload.Owner, count)
	}

	checkpoint := r.memStore.Checkpoint()
	ret, err := mem.UpdateSubscription(ctx, payload)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error updating Subscription")
	}
	return ret, nil
}

type DeleteSubscriptionPayload struct {
	ID      dssmodels.ID       `json:"id"`
	Owner   dssmodels.Owner    `json:"owner"`
	Version *dssmodels.Version `json:"version"`
}

func (r *repo) deleteSubscriptionTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*ridmodels.Subscription, error) {
	var payload *DeleteSubscriptionPayload
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
	}

	old, err := mem.GetSubscription(ctx, payload.ID)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting Subscription from repo")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", payload.ID.String())
	case !payload.Version.Matches(old.Version):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", payload.Version),
			"Subscription currently at version %s but client specified %s", old.Version, payload.Version)
	case old.Owner != payload.Owner:
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to delete", old.Owner, payload.Owner)
	}

	checkpoint := r.memStore.Checkpoint()
	ret, err := mem.DeleteSubscription(ctx, old)
	if err != nil {
		restoreErr := r.memStore.Restore(checkpoint)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Error restoring store")
		}

		return nil, stacktrace.Propagate(err, "Error deleting Subscription")
	}
	return ret, nil
}
