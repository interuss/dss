package raftstore

import (
	"context"
	"encoding/json"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/operations"
	"github.com/interuss/dss/pkg/rid/repos"
	ridmemstore "github.com/interuss/dss/pkg/rid/store/memstore"
	ridraftparams "github.com/interuss/dss/pkg/rid/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger, locality string) (*raftstore.Store[repos.Repository], error) {
	params, err := ridraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get rid raft parameters")
	}

	memStore, err := ridmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize rid memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "rid")), locality, params, r, operations.Registry)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize rid raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	// ISA
	case string(getISA):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getISA)
		}
		return r.Store.GetRepo().GetISA(ctx, id, false)

	case string(deleteISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteISA)
		}
		return r.Store.GetRepo().DeleteISA(ctx, &isa)

	case string(insertISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertISA)
		}
		return r.Store.GetRepo().InsertISA(ctx, &isa)

	case string(updateISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateISA)
		}
		return r.Store.GetRepo().UpdateISA(ctx, &isa)

	case string(searchISAs):
		var payload searchISAsPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchISAs)
		}
		return r.Store.GetRepo().SearchISAs(ctx, payload.Cells, payload.Earliest, payload.Latest)

	case string(listExpiredISAs):
		var payload expiredPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredISAs)
		}
		return r.Store.GetRepo().ListExpiredISAs(ctx, payload.Writer, payload.Threshold)

	case string(countISAs):
		return r.Store.GetRepo().CountISAs(ctx)

	// Subscriptions

	case string(getSubscription):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getSubscription)
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case string(deleteSubscription):
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
		}
		return r.Store.GetRepo().DeleteSubscription(ctx, &sub)

	case string(insertSubscription):
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertSubscription)
		}
		return r.Store.GetRepo().InsertSubscription(ctx, &sub)

	case string(updateSubscription):
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateSubscription)
		}
		return r.Store.GetRepo().UpdateSubscription(ctx, &sub)

	case string(searchSubscriptions):
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptions)
		}
		return r.Store.GetRepo().SearchSubscriptions(ctx, cells)

	case string(searchSubscriptionsByOwner):
		var payload cellsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptionsByOwner)
		}
		return r.Store.GetRepo().SearchSubscriptionsByOwner(ctx, payload.Cells, payload.Owner)

	case string(updateNotificationIdxsInCells):
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateNotificationIdxsInCells)
		}
		return r.Store.GetRepo().UpdateNotificationIdxsInCells(ctx, cells)

	case string(maxSubscriptionCountInCellsByOwner):
		var payload cellsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", maxSubscriptionCountInCellsByOwner)
		}
		return r.Store.GetRepo().MaxSubscriptionCountInCellsByOwner(ctx, payload.Cells, payload.Owner)

	case string(listExpiredSubscriptions):
		var payload expiredPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, payload.Writer, payload.Threshold)

	case string(countSubscriptions):
		return r.Store.GetRepo().CountSubscriptions(ctx)

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
