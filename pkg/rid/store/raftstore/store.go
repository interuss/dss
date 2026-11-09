package raftstore

import (
	"context"

	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/operations"
	"github.com/interuss/dss/pkg/rid/repos"
	ridmemstore "github.com/interuss/dss/pkg/rid/store/memstore"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := raftparams.GetConnectParameters("rid")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get rid raft parameters")
	}

	memStore, err := ridmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize rid memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "rid")), params, r, operations.Registry)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize rid raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch payload := proposal.Payload.GetRidPayload().GetPayload().(type) {
	// ISA operations
	case *ridpb.Payload_CreateIdentificationServiceAreaOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.CreateIdentificationServiceAreaOperationID, r.Store.GetRepo(), proposal.Payload)

	case *ridpb.Payload_UpdateIdentificationServiceAreaOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.UpdateIdentificationServiceAreaOperationID, r.Store.GetRepo(), proposal.Payload)

	case *ridpb.Payload_DeleteIdentificationServiceAreaOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.DeleteIdentificationServiceAreaOperationID, r.Store.GetRepo(), proposal.Payload)

	// ISA repository requests
	case *ridpb.Payload_GetIsaRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.GetIsaRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid ISA ID")
		}
		return r.Store.GetRepo().GetISA(ctx, id, false)

	case *ridpb.Payload_DeleteIsaRepoCall:
		req := payload.DeleteIsaRepoCall
		isa := &ridmodels.IdentificationServiceArea{Version: new(dssmodels.Version)}
		if err := isa.ID.FromProto(req.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid ISA ID")
		}
		if err := isa.Version.FromProto(req.GetVersion()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid ISA version")
		}
		return r.Store.GetRepo().DeleteISA(ctx, isa)

	case *ridpb.Payload_InsertIsaRepoCall:
		isa := new(ridmodels.IdentificationServiceArea)
		if err := isa.FromProto(payload.InsertIsaRepoCall.GetIsa()); err != nil {
			return nil, err
		}
		return r.Store.GetRepo().InsertISA(ctx, isa)

	case *ridpb.Payload_UpdateIsaRepoCall:
		req := payload.UpdateIsaRepoCall
		isa := new(ridmodels.IdentificationServiceArea)
		if err := isa.FromProto(req.GetIsa()); err != nil {
			return nil, err
		}
		isa.Version = new(dssmodels.Version)
		if err := isa.Version.FromProto(req.GetVersion()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid ISA version")
		}
		return r.Store.GetRepo().UpdateISA(ctx, isa)

	case *ridpb.Payload_SearchIsasRepoCall:
		req := payload.SearchIsasRepoCall
		return r.Store.GetRepo().SearchISAs(ctx, dssmodels.CellUnionFromProto(req.GetCells()), dssmodels.TimeFromProto(req.GetEarliest()), dssmodels.TimeFromProto(req.GetLatest()))

	case *ridpb.Payload_CountIsasRepoCall:
		return r.Store.GetRepo().CountISAs(ctx)

	// Subscription operations
	case *ridpb.Payload_CreateSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.CreateSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *ridpb.Payload_UpdateSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.UpdateSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *ridpb.Payload_DeleteSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, ridv2.DeleteSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	// Subscription repository requests
	case *ridpb.Payload_GetSubscriptionRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.GetSubscriptionRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID")
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case *ridpb.Payload_DeleteSubscriptionRepoCall:
		req := payload.DeleteSubscriptionRepoCall
		sub := &ridmodels.Subscription{Version: new(dssmodels.Version)}
		if err := sub.ID.FromProto(req.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID")
		}
		if err := sub.Version.FromProto(req.GetVersion()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription version")
		}
		return r.Store.GetRepo().DeleteSubscription(ctx, sub)

	case *ridpb.Payload_InsertSubscriptionRepoCall:
		sub := new(ridmodels.Subscription)
		if err := sub.FromProto(payload.InsertSubscriptionRepoCall.GetSubscription()); err != nil {
			return nil, err
		}
		return r.Store.GetRepo().InsertSubscription(ctx, sub)

	case *ridpb.Payload_UpdateSubscriptionRepoCall:
		req := payload.UpdateSubscriptionRepoCall
		sub := new(ridmodels.Subscription)
		if err := sub.FromProto(req.GetSubscription()); err != nil {
			return nil, err
		}
		sub.Version = new(dssmodels.Version)
		if err := sub.Version.FromProto(req.GetVersion()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription version")
		}
		return r.Store.GetRepo().UpdateSubscription(ctx, sub)

	case *ridpb.Payload_SearchSubscriptionsRepoCall:
		return r.Store.GetRepo().SearchSubscriptions(ctx, dssmodels.CellUnionFromProto(payload.SearchSubscriptionsRepoCall.GetCells()))

	case *ridpb.Payload_SearchSubscriptionsByOwnerRepoCall:
		req := payload.SearchSubscriptionsByOwnerRepoCall
		return r.Store.GetRepo().SearchSubscriptionsByOwner(ctx, dssmodels.CellUnionFromProto(req.GetCells()), dssmodels.Owner(req.GetOwner()))

	case *ridpb.Payload_UpdateNotificationIdxsInCellsRepoCall:
		return r.Store.GetRepo().UpdateNotificationIdxsInCells(ctx, dssmodels.CellUnionFromProto(payload.UpdateNotificationIdxsInCellsRepoCall.GetCells()))

	case *ridpb.Payload_MaxSubscriptionCountInCellsByOwnerRepoCall:
		req := payload.MaxSubscriptionCountInCellsByOwnerRepoCall
		return r.Store.GetRepo().MaxSubscriptionCountInCellsByOwner(ctx, dssmodels.CellUnionFromProto(req.GetCells()), dssmodels.Owner(req.GetOwner()))

	case *ridpb.Payload_CountSubscriptionsRepoCall:
		return r.Store.GetRepo().CountSubscriptions(ctx)

	default:
		return nil, stacktrace.NewError("unknown request: %q", proposal.Name())
	}
}
