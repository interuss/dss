package raftstore

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore"
	raftstoreadmin "github.com/interuss/dss/pkg/raftstore/admin"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	scdmemstore "github.com/interuss/dss/pkg/scd/store/memstore"
	scdraftparams "github.com/interuss/dss/pkg/scd/store/raftstore/params"
	"github.com/interuss/stacktrace"

	"go.uber.org/zap"
)

const (
	getOperationalIntent           raftstore.RequestType = "getOperationalIntent"
	deleteOperationalIntent        raftstore.RequestType = "deleteOperationalIntent"
	upsertOperationalIntent        raftstore.RequestType = "upsertOperationalIntent"
	searchOperationalIntents       raftstore.RequestType = "searchOperationalIntents"
	getDependentOperationalIntents raftstore.RequestType = "getDependentOperationalIntents"
	listExpiredOperationalIntents  raftstore.RequestType = "listExpiredOperationalIntents"
	countOperationalIntents        raftstore.RequestType = "countOperationalIntents"

	DeleteOperationalIntentTransaction raftstore.RequestType = "deleteOperationalIntentTransaction"
	GetOperationalIntentTransaction    raftstore.RequestType = "getOperationalIntentTransaction"
	QueryOperationalIntentTransaction  raftstore.RequestType = "queryOperationalIntentTransaction"
	UpsertOperationalIntentTransaction raftstore.RequestType = "upsertOperationalIntentTransaction"

	searchSubscriptions                 raftstore.RequestType = "searchSubscriptions"
	getSubscription                     raftstore.RequestType = "getSubscription"
	upsertSubscription                  raftstore.RequestType = "upsertSubscription"
	deleteSubscription                  raftstore.RequestType = "deleteSubscription"
	incrementNotificationForOIs         raftstore.RequestType = "incrementNotificationForOIs"
	incrementNotificationForConstraints raftstore.RequestType = "incrementNotificationForConstraints"
	listExpiredSubscriptions            raftstore.RequestType = "listExpiredSubscriptions"
	countSubscriptions                  raftstore.RequestType = "countSubscriptions"

	DeleteSubscriptionTransaction raftstore.RequestType = "deleteSubscriptionTransaction"
	GetSubscriptionTransaction    raftstore.RequestType = "getSubscriptionTransaction"
	QuerySubscriptionTransaction  raftstore.RequestType = "querySubscriptionTransaction"
	UpsertSubscriptionTransaction raftstore.RequestType = "upsertSubscriptionTransaction"

	searchConstraints raftstore.RequestType = "searchConstraints"
	getConstraint     raftstore.RequestType = "getConstraint"
	upsertConstraint  raftstore.RequestType = "upsertConstraint"
	deleteConstraint  raftstore.RequestType = "deleteConstraint"
	countConstraints  raftstore.RequestType = "countConstraints"

	DeleteConstraintTransaction raftstore.RequestType = "deleteConstraintTransaction"
	GetConstraintTransaction    raftstore.RequestType = "getConstraintTransaction"
	QueryConstraintTransaction  raftstore.RequestType = "queryConstraintTransaction"
	UpsertConstraintTransaction raftstore.RequestType = "upsertConstraintTransaction"

	getUSSAvailability    raftstore.RequestType = "getUSSAvailability"
	upsertUSSAvailability raftstore.RequestType = "upsertUSSAvailability"

	GetUSSAvailabilityTransaction raftstore.RequestType = "getUSSAvailabilityTransaction"
	SetUSSAvailabilityTransaction raftstore.RequestType = "setUSSAvailabilityTransaction"
)

var readOnlyRequests = []raftstore.RequestType{
	getOperationalIntent,
	searchOperationalIntents,
	getDependentOperationalIntents,
	listExpiredOperationalIntents,
	countOperationalIntents,

	GetOperationalIntentTransaction,
	QueryOperationalIntentTransaction,

	searchSubscriptions,
	getSubscription,
	listExpiredSubscriptions,
	countSubscriptions,

	GetSubscriptionTransaction,
	QuerySubscriptionTransaction,

	searchConstraints,
	getConstraint,
	countConstraints,

	GetConstraintTransaction,
	QueryConstraintTransaction,

	getUSSAvailability,
	GetUSSAvailabilityTransaction,
}

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus

	memStore *memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := scdraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}

	memStore, err := scdmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize SCD memstore")
	}

	r := &repo{memStore: memStore}
	store, err := raftstore.Init(ctx, logger, params, r)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize SCDs raftstore")
	}

	r.consensus = store.Consensus

	raftstoreadmin.Register(ctx, logger, "scd", store)

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) IsReadOnly(requestType raftstore.RequestType) bool {
	return slices.Contains(readOnlyRequests, requestType)
}

func (r *repo) GetSnapshot() ([]byte, error) {
	return r.memStore.GetSnapshot()
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	return r.memStore.RestoreFromSnapshot(data)
}

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	mem, err := r.memStore.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain scd memstore repository")
	}

	switch raftstore.RequestType(proposal.RequestType) {
	case getOperationalIntent:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", getOperationalIntent)
		}

		return mem.GetOperationalIntent(ctx, id)

	case deleteOperationalIntent:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", deleteOperationalIntent)
		}

		return nil, mem.DeleteOperationalIntent(ctx, id)

	case upsertOperationalIntent:
		var operation *scdmodels.OperationalIntent
		if err := json.Unmarshal(proposal.Value, &operation); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", upsertOperationalIntent)
		}

		return mem.UpsertOperationalIntent(ctx, operation)

	case searchOperationalIntents:
		var v4d *dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", searchOperationalIntents)
		}

		return mem.SearchOperationalIntents(ctx, v4d)

	case getDependentOperationalIntents:
		var subscriptionID dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &subscriptionID); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", getDependentOperationalIntents)
		}

		return mem.GetDependentOperationalIntents(ctx, subscriptionID)

	case listExpiredOperationalIntents:
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", listExpiredOperationalIntents)
		}

		return mem.ListExpiredOperationalIntents(ctx, threshold)

	case countOperationalIntents:
		return mem.CountOperationalIntents(ctx)

	case DeleteOperationalIntentTransaction:
		return r.deleteOperationalIntentTransactionApplier(ctx, proposal, mem)

	case GetOperationalIntentTransaction:
		return r.getOperationalIntentTransactionApplier(ctx, proposal, mem)

	case QueryOperationalIntentTransaction:
		return r.queryOperationalIntentTransactionApplier(ctx, proposal, mem)

	case UpsertOperationalIntentTransaction:
		return r.upsertOperationalIntentTransactionApplier(ctx, proposal, mem)

	case searchSubscriptions:
		var v4d *dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", searchSubscriptions)
		}

		return mem.SearchSubscriptions(ctx, v4d)

	case getSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", getSubscription)
		}

		return mem.GetSubscription(ctx, id)

	case upsertSubscription:
		var sub *scdmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", upsertSubscription)
		}

		return mem.UpsertSubscription(ctx, sub)

	case deleteSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", deleteSubscription)
		}

		return nil, mem.DeleteSubscription(ctx, id)

	case incrementNotificationForOIs:
		var v4d *dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", incrementNotificationForOIs)
		}

		return mem.IncrementNotificationIndicesForOperationalIntents(ctx, v4d)

	case incrementNotificationForConstraints:
		var v4d *dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", incrementNotificationForConstraints)
		}

		return mem.IncrementNotificationIndicesForConstraints(ctx, v4d)

	case listExpiredSubscriptions:
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", listExpiredSubscriptions)
		}

		return mem.ListExpiredSubscriptions(ctx, threshold)

	case countSubscriptions:
		return mem.CountSubscriptions(ctx)

	case UpsertSubscriptionTransaction:
		return r.upsertSubscriptionTransactionApplier(ctx, proposal, mem)

	case DeleteSubscriptionTransaction:
		return r.deleteSubscriptionTransactionApplier(ctx, proposal, mem)

	case GetSubscriptionTransaction:
		return r.getSubscriptionTransactionApplier(ctx, proposal, mem)

	case QuerySubscriptionTransaction:
		return r.querySubscriptionTransactionApplier(ctx, proposal, mem)

	case searchConstraints:
		var v4d *dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", searchConstraints)
		}

		return mem.SearchConstraints(ctx, v4d)

	case getConstraint:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", getConstraint)
		}

		return mem.GetConstraint(ctx, id)

	case upsertConstraint:
		var constraint *scdmodels.Constraint
		if err := json.Unmarshal(proposal.Value, &constraint); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", upsertConstraint)
		}

		return mem.UpsertConstraint(ctx, constraint)

	case deleteConstraint:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", deleteConstraint)
		}

		return nil, mem.DeleteConstraint(ctx, id)

	case countConstraints:
		return mem.CountConstraints(ctx)

	case DeleteConstraintTransaction:
		return r.deleteConstraintTransactionApplier(ctx, proposal, mem)

	case GetConstraintTransaction:
		return r.getConstraintTransactionApplier(ctx, proposal, mem)

	case QueryConstraintTransaction:
		return r.queryConstraintTransactionApplier(ctx, proposal, mem)

	case UpsertConstraintTransaction:
		return r.upsertConstraintTransactionApplier(ctx, proposal, mem)

	case getUSSAvailability:
		var id dssmodels.Manager
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", getUSSAvailability)
		}

		return mem.GetUssAvailability(ctx, id)

	case upsertUSSAvailability:
		var ussa scdmodels.UssAvailabilityStatus
		if err := json.Unmarshal(proposal.Value, &ussa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s proposal value", upsertUSSAvailability)
		}

		return mem.UpsertUssAvailability(ctx, &ussa)

	case GetUSSAvailabilityTransaction:
		return r.getUSSAvailabilityTransactionApplier(ctx, proposal, mem)

	case SetUSSAvailabilityTransaction:
		return r.setUSSAvailabilityTransactionApplier(ctx, proposal, mem)

	default:
		return nil, stacktrace.NewError("unknown request type: %q", proposal.RequestType)
	}
}
