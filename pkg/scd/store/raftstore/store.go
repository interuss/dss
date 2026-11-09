package raftstore

import (
	"context"
	"encoding/json"
	"time"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/dss/pkg/scd/repos"
	scdmemstore "github.com/interuss/dss/pkg/scd/store/memstore"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := raftparams.GetConnectParameters("scd")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}

	memStore, err := scdmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "scd")), params, r, operations.Registry)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch payload := proposal.Payload.GetScdPayload().GetPayload().(type) {
	// USS availability operations
	case *scdpb.Payload_GetUssAvailabilityOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.GetUssAvailabilityOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_SetUssAvailabilityOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.SetUssAvailabilityOperationID, r.Store.GetRepo(), proposal.Payload)

	// USS availability repository requests
	case *scdpb.Payload_GetUssAvailabilityRepoCall:
		return r.Store.GetRepo().GetUssAvailability(ctx, dssmodels.Manager(payload.GetUssAvailabilityRepoCall.GetUss()))

	case *scdpb.Payload_UpsertUssAvailabilityRepoCall:
		req := payload.UpsertUssAvailabilityRepoCall
		ussa := &scdmodels.UssAvailabilityStatus{Uss: dssmodels.Manager(req.GetUss())}
		if err := ussa.Availability.FromProto(req.GetAvailability()); err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode USS availability state for operation %q", restapi.SetUssAvailabilityOperationID)
		}
		if err := ussa.Version.FromProto(req.GetVersion()); err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode USS version for operation %q", restapi.SetUssAvailabilityOperationID)
		}
		return r.Store.GetRepo().UpsertUssAvailability(ctx, ussa)

	// Constraint operations
	case *scdpb.Payload_GetConstraintReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.GetConstraintReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_CreateConstraintReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.CreateConstraintReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_UpdateConstraintReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.UpdateConstraintReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_DeleteConstraintReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.DeleteConstraintReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_QueryConstraintReferencesOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.QueryConstraintReferencesOperationID, r.Store.GetRepo(), proposal.Payload)

	// Constraint repository requests
	case *scdpb.Payload_SearchConstraintsRepoCall:
		volume := payload.SearchConstraintsRepoCall.GetVolume()
		if volume == nil {
			return nil, stacktrace.NewError("missing volume")
		}
		cellsVolume := new(dssmodels.CellsVolume4D)
		cellsVolume.FromProto(volume)
		return r.Store.GetRepo().SearchConstraints(ctx, cellsVolume)

	case *scdpb.Payload_GetConstraintRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.GetConstraintRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Constraint ID")
		}
		return r.Store.GetRepo().GetConstraint(ctx, id)

	case *scdpb.Payload_UpsertConstraintRepoCall:
		req := payload.UpsertConstraintRepoCall
		constraint := new(scdmodels.Constraint)
		if err := constraint.FromProto(req.GetConstraint()); err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode Constraint")
		}
		constraint.Version = scdmodels.VersionNumber(req.GetVersion())
		return r.Store.GetRepo().UpsertConstraint(ctx, constraint)

	case *scdpb.Payload_DeleteConstraintRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.DeleteConstraintRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Constraint ID")
		}
		return nil, r.Store.GetRepo().DeleteConstraint(ctx, id)

	case *scdpb.Payload_CountConstraintsRepoCall:
		return r.Store.GetRepo().CountConstraints(ctx)

	// Operational intent operations
	case *scdpb.Payload_GetOperationalIntentReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.GetOperationalIntentReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_CreateOperationalIntentReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.CreateOperationalIntentReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_UpdateOperationalIntentReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.UpdateOperationalIntentReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_DeleteOperationalIntentReferenceOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.DeleteOperationalIntentReferenceOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_QueryOperationalIntentReferencesOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.QueryOperationalIntentReferencesOperationID, r.Store.GetRepo(), proposal.Payload)

	// Operational intent repository requests
	case *scdpb.Payload_GetOperationalIntentRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.GetOperationalIntentRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid OperationalIntent ID")
		}
		return r.Store.GetRepo().GetOperationalIntent(ctx, id)

	case *scdpb.Payload_DeleteOperationalIntentRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.DeleteOperationalIntentRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid OperationalIntent ID")
		}
		return nil, r.Store.GetRepo().DeleteOperationalIntent(ctx, id)

	case *scdpb.Payload_UpsertOperationalIntentRepoCall:
		operation := new(scdmodels.OperationalIntent)
		if err := operation.FromProto(payload.UpsertOperationalIntentRepoCall.GetOperationalIntent()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid OperationalIntent")
		}
		return r.Store.GetRepo().UpsertOperationalIntent(ctx, operation)

	case *scdpb.Payload_SearchOperationalIntentsRepoCall:
		volume := payload.SearchOperationalIntentsRepoCall.GetVolume()
		if volume == nil {
			return nil, stacktrace.NewError("missing volume")
		}
		cellsVolume := new(dssmodels.CellsVolume4D)
		cellsVolume.FromProto(volume)
		return r.Store.GetRepo().SearchOperationalIntents(ctx, cellsVolume)

	case *scdpb.Payload_GetDependentOperationalIntentsRepoCall:
		var subscriptionID dssmodels.ID
		if err := subscriptionID.FromProto(payload.GetDependentOperationalIntentsRepoCall.GetSubscriptionId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID")
		}
		return r.Store.GetRepo().GetDependentOperationalIntents(ctx, subscriptionID)

	case *scdpb.Payload_CountOperationalIntentsRepoCall:
		return r.Store.GetRepo().CountOperationalIntents(ctx)

	// Subscription operations
	case *scdpb.Payload_GetSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.GetSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_CreateSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.CreateSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_UpdateSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.UpdateSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_DeleteSubscriptionOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.DeleteSubscriptionOperationID, r.Store.GetRepo(), proposal.Payload)

	case *scdpb.Payload_QuerySubscriptionsOperation:
		return dssstore.ExecutePayload(ctx, operations.Registry, restapi.QuerySubscriptionsOperationID, r.Store.GetRepo(), proposal.Payload)

	// Subscription repository requests
	case *scdpb.Payload_SearchSubscriptionsRepoCall:
		volume := payload.SearchSubscriptionsRepoCall.GetVolume()
		if volume == nil {
			return nil, stacktrace.NewError("missing volume")
		}
		cellsVolume := new(dssmodels.CellsVolume4D)
		cellsVolume.FromProto(volume)
		return r.Store.GetRepo().SearchSubscriptions(ctx, cellsVolume)

	case *scdpb.Payload_GetSubscriptionRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.GetSubscriptionRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID")
		}
		return r.Store.GetRepo().GetSubscription(ctx, id)

	case *scdpb.Payload_UpsertSubscriptionRepoCall:
		req := payload.UpsertSubscriptionRepoCall
		sub := new(scdmodels.Subscription)
		if err := sub.FromProto(req.GetSubscription()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription")
		}
		sub.NotificationIndex = int(req.GetNotificationIndex())
		return r.Store.GetRepo().UpsertSubscription(ctx, sub)

	case *scdpb.Payload_DeleteSubscriptionRepoCall:
		var id dssmodels.ID
		if err := id.FromProto(payload.DeleteSubscriptionRepoCall.GetId()); err != nil {
			return nil, stacktrace.Propagate(err, "invalid Subscription ID")
		}
		return nil, r.Store.GetRepo().DeleteSubscription(ctx, id)

	case *scdpb.Payload_IncrementNotificationIndicesForOperationalIntentsRepoCall:
		volume := payload.IncrementNotificationIndicesForOperationalIntentsRepoCall.GetVolume()
		if volume == nil {
			return nil, stacktrace.NewError("missing volume")
		}
		cellsVolume := new(dssmodels.CellsVolume4D)
		cellsVolume.FromProto(volume)
		return r.Store.GetRepo().IncrementNotificationIndicesForOperationalIntents(ctx, cellsVolume)

	case *scdpb.Payload_IncrementNotificationIndicesForConstraintsRepoCall:
		volume := payload.IncrementNotificationIndicesForConstraintsRepoCall.GetVolume()
		if volume == nil {
			return nil, stacktrace.NewError("missing volume")
		}
		cellsVolume := new(dssmodels.CellsVolume4D)
		cellsVolume.FromProto(volume)
		return r.Store.GetRepo().IncrementNotificationIndicesForConstraints(ctx, cellsVolume)

	case *scdpb.Payload_CountSubscriptionsRepoCall:
		return r.Store.GetRepo().CountSubscriptions(ctx)

	default:
		return r.applyLegacy(ctx, proposal)
	}
}

// applyLegacy applies the proposals carrying a JSON encoded request.
// TODO: remove once all the requests are carried by a payload.
func (r *repo) applyLegacy(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {

	// Subscriptions

	case string(listExpiredSubscriptions):
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return r.Store.GetRepo().ListExpiredSubscriptions(ctx, threshold)

	// Operational Intents

	case string(listExpiredOperationalIntents):
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredOperationalIntents)
		}
		return r.Store.GetRepo().ListExpiredOperationalIntents(ctx, threshold)

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
