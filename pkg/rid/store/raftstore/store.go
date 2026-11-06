package raftstore

import (
	"context"
	"encoding/json"

	"github.com/golang/geo/s2"
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

	default:
		return r.applyLegacy(ctx, proposal)
	}
}

// applyLegacy applies the proposals carrying a JSON encoded request.
// TODO: remove once all the requests are carried by a payload.
func (r *repo) applyLegacy(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
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
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, payload.Threshold)

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
