package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/dss/pkg/scd/repos"
	scdmemstore "github.com/interuss/dss/pkg/scd/store/memstore"
	scdraftparams "github.com/interuss/dss/pkg/scd/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger, locality string) (*raftstore.Store[repos.Repository], error) {
	params, err := scdraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}

	memStore, err := scdmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "scd")), locality, params, r, operations.Registry)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {

	// Constraints

	case string(searchConstraints):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchConstraints)
		}
		return r.Store.GetRepo().SearchConstraints(ctx, &cellsVolume)

	case string(getConstraint):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getConstraint)
		}
		return r.Store.GetRepo().GetConstraint(ctx, id)

	case string(upsertConstraint):
		var constraint scdmodels.Constraint
		if err := json.Unmarshal(proposal.Value, &constraint); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertConstraint)
		}
		return r.Store.GetRepo().UpsertConstraint(ctx, &constraint)

	case string(deleteConstraint):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteConstraint)
		}
		return nil, r.Store.GetRepo().DeleteConstraint(ctx, id)

	case string(countConstraints):
		return r.Store.GetRepo().CountConstraints(ctx)

	// Subscriptions

	case string(searchSubscriptions):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptions)
		}
		return r.Store.GetRepo().SearchSubscriptions(ctx, &cellsVolume)

	case string(getSubscription):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getSubscription)
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case string(upsertSubscription):
		var sub scdmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertSubscription)
		}
		return r.Store.GetRepo().UpsertSubscription(ctx, &sub)

	case string(deleteSubscription):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
		}
		return nil, r.Store.GetRepo().DeleteSubscription(ctx, id)

	case string(incrementNotificationIndicesForOperationalIntents):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", incrementNotificationIndicesForOperationalIntents)
		}
		return r.Store.GetRepo().IncrementNotificationIndicesForOperationalIntents(ctx, &cellsVolume)

	case string(incrementNotificationIndicesForConstraints):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", incrementNotificationIndicesForConstraints)
		}
		return r.Store.GetRepo().IncrementNotificationIndicesForConstraints(ctx, &cellsVolume)

	case string(listExpiredSubscriptions):
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, threshold)

	case string(countSubscriptions):
		return r.Store.GetRepo().CountSubscriptions(ctx)

	// Operational Intents

	case string(getOperationalIntent):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getOperationalIntent)
		}
		return r.Store.GetRepo().GetOperationalIntent(ctx, id)

	case string(deleteOperationalIntent):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteOperationalIntent)
		}
		return nil, r.Store.GetRepo().DeleteOperationalIntent(ctx, id)

	case string(upsertOperationalIntent):
		var operation scdmodels.OperationalIntent
		if err := json.Unmarshal(proposal.Value, &operation); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertOperationalIntent)
		}
		return r.Store.GetRepo().UpsertOperationalIntent(ctx, &operation)

	case string(searchOperationalIntents):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchOperationalIntents)
		}
		return r.Store.GetRepo().SearchOperationalIntents(ctx, &cellsVolume)

	case string(getDependentOperationalIntents):
		var subscriptionID dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &subscriptionID); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getDependentOperationalIntents)
		}
		return r.Store.GetRepo().GetDependentOperationalIntents(ctx, subscriptionID)

	case string(listExpiredOperationalIntents):
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredOperationalIntents)
		}
		return r.Store.GetRepo().ListExpiredOperationalIntents(ctx, threshold)

	case string(countOperationalIntents):
		return r.Store.GetRepo().CountOperationalIntents(ctx)

	// USS Availability

	case string(getUssAvailability):
		var manager dssmodels.Manager
		if err := json.Unmarshal(proposal.Value, &manager); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getUssAvailability)
		}
		return r.Store.GetRepo().GetUssAvailability(ctx, manager)

	case string(upsertUssAvailability):
		var ussa scdmodels.UssAvailabilityStatus
		if err := json.Unmarshal(proposal.Value, &ussa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertUssAvailability)
		}
		return r.Store.GetRepo().UpsertUssAvailability(ctx, &ussa)

	// Operations registry (transactions)
	default:
		handler, ok := operations.Registry[proposal.RequestType]
		if !ok {
			return nil, stacktrace.NewError("unrecognized request type: %s", proposal.RequestType)
		}

		request, err := handler.Decode(proposal.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode %s payload", proposal.RequestType)
		}

		return handler.Execute(ctx, r.Store.GetRepo(), request)
	}
}
