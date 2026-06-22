package raftstore

import (
	"context"
	"encoding/json"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

func (r *repo) deleteOperationalIntentTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.ChangeOperationalIntentReferenceResponse, error) {
	var req *restapi.DeleteOperationalIntentReferenceRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal delete operational intent request")
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	ovn := scdmodels.OVN(req.Ovn)
	if ovn == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing OVN for operational intent to modify")
	}

	old, err := mem.GetOperationalIntent(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get OperationIntent from repo")
	}
	if old == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
	}

	if old.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"OperationalIntent owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	}

	if old.OVN != ovn {
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, ovn)
	}

	// Get the Subscription supporting the OperationalIntent, if one is defined
	var previousSubscription *scdmodels.Subscription
	if old.SubscriptionID != nil {
		previousSubscription, err = mem.GetSubscription(ctx, *old.SubscriptionID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
		}
		if previousSubscription == nil {
			return nil, stacktrace.NewError("OperationalIntent's Subscription missing from repo")
		}
	}

	removeImplicitSubscription, err := repos.SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx, mem, id, previousSubscription)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not determine if Subscription can be removed")
	}

	// Gather the subscriptions that need to be notified
	notifyVolume := &dssmodels.Volume4D{
		StartTime: old.StartTime,
		EndTime:   old.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: old.AltitudeUpper,
			AltitudeLo: old.AltitudeLower,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return old.Cells, nil
			}),
		}}

	cp := r.memStore.Checkpoint()
	subsToNotify, err := repos.GetRelevantSubscriptionsAndIncrementIndices(ctx, mem, notifyVolume)
	if err != nil {
		restoreErr := r.memStore.Restore(cp)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}

		return nil, stacktrace.Propagate(err, "could not obtain relevant subscriptions")
	}

	if err := mem.DeleteOperationalIntent(ctx, id); err != nil {
		restoreErr := r.memStore.Restore(cp)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}

		return nil, stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
	}

	// removeImplicitSubscription is only true if the OIR had a subscription defined
	if removeImplicitSubscription {
		// Automatically remove a now-unused implicit Subscription
		err = mem.DeleteSubscription(ctx, previousSubscription.ID)
		if err != nil {
			restoreErr := r.memStore.Restore(cp)
			if restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}

			return nil, stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
		}
	}

	return &restapi.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: *old.ToRest(),
		Subscribers:                repos.MakeSubscribersToNotify(subsToNotify),
	}, nil
}

func (r *repo) getOperationalIntentTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.GetOperationalIntentReferenceResponse, error) {
	var req *restapi.GetOperationalIntentReferenceRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal get operational intent request")
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	op, err := mem.GetOperationalIntent(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get OperationalIntent from repo")
	}
	if op == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
	}

	if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		op.OVN = scdmodels.NoOvnPhrase
	}

	return &restapi.GetOperationalIntentReferenceResponse{
		OperationalIntentReference: *op.ToRest(),
	}, nil
}

func (r *repo) queryOperationalIntentTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*restapi.QueryOperationalIntentReferenceResponse, error) {
	var req *restapi.QueryOperationalIntentReferencesRequest
	if err := json.Unmarshal(proposal.Value, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal query operational intent request")
	}

	vol4, err := dssmodels.Volume4DFromSCDRest(req.Body.AreaOfInterest)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	ops, err := mem.SearchOperationalIntents(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
	}

	response := &restapi.QueryOperationalIntentReferenceResponse{
		OperationalIntentReferences: make([]restapi.OperationalIntentReference, 0, len(ops)),
	}
	for _, op := range ops {
		p := op.ToRest()
		if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
			p.Ovn = &noOvnPhrase
		}
		response.OperationalIntentReferences = append(response.OperationalIntentReferences, *p)
	}

	return response, nil
}

type UpsertOperationalIntentTransactionPayload struct {
	Manager     dssmodels.Manager
	ValidParams *repos.ValidOIRParams
	Key         []scdmodels.OVN
}

type UpsertOperationalIntentTransactionResult struct {
	ResponseOK       *restapi.ChangeOperationalIntentReferenceResponse
	ResponseConflict *restapi.AirspaceConflictResponse
}

func (r *repo) upsertOperationalIntentTransactionApplier(ctx context.Context, proposal consensus.Proposal, mem repos.Repository) (*UpsertOperationalIntentTransactionResult, error) {
	var payload *UpsertOperationalIntentTransactionPayload
	if err := json.Unmarshal(proposal.Value, &payload); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal upsert operational intent request")
	}

	upsertResult := &UpsertOperationalIntentTransactionResult{}

	key := make(map[scdmodels.OVN]bool, len(payload.Key))
	for _, ovn := range payload.Key {
		key[ovn] = true
	}

	old, err := mem.GetOperationalIntent(ctx, payload.ValidParams.ID)
	if err != nil {
		return upsertResult, stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
	}

	if err := repos.ValidateUpsertRequestAgainstPreviousOIR(payload.Manager, payload.ValidParams.OVN, old); err != nil {
		return upsertResult, stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Request validation failed")
	}

	var (
		version     = scdmodels.VersionNumber(1)
		pastOVNs    = make([]scdmodels.OVN, 0)
		previousSub *scdmodels.Subscription
	)
	if old != nil {
		version = old.Version + 1
		pastOVNs = append(old.PastOVNs, payload.ValidParams.OVN)

		if old.SubscriptionID != nil {
			previousSub, err = mem.GetSubscription(ctx, *old.SubscriptionID)
			if err != nil {
				return upsertResult, stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
			}
		}
	}

	previousSubIsBeingReplaced := previousSub != nil && payload.ValidParams.SubscriptionID != previousSub.ID
	removePreviousImplicitSubscription := false
	if previousSubIsBeingReplaced {
		removePreviousImplicitSubscription, err = repos.SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx, mem, payload.ValidParams.ID, previousSub)
		if err != nil {
			return upsertResult, stacktrace.Propagate(err, "Could not determine if previous Subscription can be removed")
		}
	}

	// Every error path after this point must restore the store to the checkpoint, since we may have already written to the store.
	cp := r.memStore.Checkpoint()

	attachedSub := previousSub
	if payload.ValidParams.SubscriptionID.Empty() {
		if payload.ValidParams.ImplicitSubscription.Requested {
			if attachedSub, err = repos.CreateAndStoreNewImplicitSubscription(ctx, mem, payload.Manager, payload.ValidParams); err != nil {
				restoreErr := r.memStore.Restore(cp)
				if restoreErr != nil {
					return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
				}
				return upsertResult, stacktrace.Propagate(err, "Failed to create implicit subscription")
			}
		} else {
			attachedSub = nil
		}
	} else {
		if attachedSub == nil || previousSubIsBeingReplaced {
			attachedSub, err = mem.GetSubscription(ctx, payload.ValidParams.SubscriptionID)
			if err != nil {
				restoreErr := r.memStore.Restore(cp)
				if restoreErr != nil {
					return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
				}
				return upsertResult, stacktrace.Propagate(err, "Failed to ensure subscription covers OIR")
			}

			if attachedSub == nil {
				restoreErr := r.memStore.Restore(cp)
				if restoreErr != nil {
					return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
				}
				return upsertResult, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", payload.ValidParams.SubscriptionID)
			}
		}

		if attachedSub.Manager != payload.Manager {
			restoreErr := r.memStore.Restore(cp)
			if restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}

			return upsertResult, stacktrace.Propagate(
				stacktrace.NewErrorWithCode(
					dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
				"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent",
				payload.ValidParams.SubscriptionID,
				attachedSub.Manager,
				payload.Manager,
			)
		}

		attachedSub, err = repos.EnsureSubscriptionCoversOIR(ctx, mem, attachedSub, payload.ValidParams)
		if err != nil {
			restoreErr := r.memStore.Restore(cp)
			if restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}

			return upsertResult, stacktrace.Propagate(err, "Failed to ensure subscription covers OIR")
		}
	}

	if payload.ValidParams.State.RequiresKey() {
		upsertResult.ResponseConflict, err = repos.ValidateKeyAndProvideConflictResponse(ctx, mem, payload.Manager, payload.ValidParams, attachedSub)
		if err != nil {
			restoreErr := r.memStore.Restore(cp)
			if restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}

			return upsertResult, stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Failed to validate key")
		}
	}

	op := payload.ValidParams.ToOIR(payload.Manager, attachedSub, version, pastOVNs)

	op, err = mem.UpsertOperationalIntent(ctx, op)
	if err != nil {
		restoreErr := r.memStore.Restore(cp)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return upsertResult, stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
	}

	if removePreviousImplicitSubscription {
		if err = mem.DeleteSubscription(ctx, previousSub.ID); err != nil {
			restoreErr := r.memStore.Restore(cp)
			if restoreErr != nil {
				return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
			}
			return upsertResult, stacktrace.Propagate(err, "Unable to delete previous implicit Subscription")
		}
	}

	notifyVolume, err := repos.ComputeNotificationVolume(old, payload.ValidParams.UExtent)
	if err != nil {
		restoreErr := r.memStore.Restore(cp)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return upsertResult, stacktrace.Propagate(err, "Failed to compute notification volume")
	}

	subsToNotify, err := repos.GetRelevantSubscriptionsAndIncrementIndices(ctx, mem, notifyVolume)
	if err != nil {
		restoreErr := r.memStore.Restore(cp)
		if restoreErr != nil {
			return nil, stacktrace.Propagate(restoreErr, "Failed to restore store")
		}
		return upsertResult, stacktrace.Propagate(err, "Failed to notify relevant Subscriptions")
	}

	upsertResult.ResponseOK = &restapi.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: *op.ToRest(),
		Subscribers:                repos.MakeSubscribersToNotify(subsToNotify),
	}

	return upsertResult, nil
}
