package scd

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/scd/actions"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

// DeleteOperationalIntentReference deletes a single operational intent ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationalIntentReference(ctx context.Context, req *restapi.DeleteOperationalIntentReferenceRequest,
) restapi.DeleteOperationalIntentReferenceResponseSet {

	// Retrieve OperationalIntent ID
	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	// Retrieve OVN
	ovn := scdmodels.OVN(req.Ovn)
	if ovn == "" {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing OVN for operational intent to modify"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.ChangeOperationalIntentReferenceResponse](ctx, a.Store, req)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete operational intent")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.NotFound:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response404: errResp}
		case dsserr.VersionMismatch:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response409: errResp}
		default:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.DeleteOperationalIntentReferenceResponseSet{Response200: response}
}

// GetOperationalIntentReference returns a single operation intent ref for the given ID.
func (a *Server) GetOperationalIntentReference(ctx context.Context, req *restapi.GetOperationalIntentReferenceRequest,
) restapi.GetOperationalIntentReferenceResponseSet {

	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.GetOperationalIntentReferenceResponse](ctx, a.Store, req)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not get operational intent")
		if stacktrace.GetCode(err) == dsserr.NotFound {
			return restapi.GetOperationalIntentReferenceResponseSet{Response404: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.GetOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.GetOperationalIntentReferenceResponseSet{Response200: response}
}

// QueryOperationalIntentReferences queries existing operational intent refs in the given
// bounds.
func (a *Server) QueryOperationalIntentReferences(ctx context.Context, req *restapi.QueryOperationalIntentReferencesRequest,
) restapi.QueryOperationalIntentReferencesResponseSet {

	if req.BodyParseError != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest"))}}
	}

	// Parse area of interest to common Volume4D
	_, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.QueryOperationalIntentReferenceResponse](ctx, a.Store, req)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not query operational intent")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.QueryOperationalIntentReferencesResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.QueryOperationalIntentReferencesResponseSet{Response200: response}
}

func (a *Server) CreateOperationalIntentReference(ctx context.Context, req *restapi.CreateOperationalIntentReferenceRequest,
) restapi.CreateOperationalIntentReferenceResponseSet {

	if req.BodyParseError != nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	respOK, respConflict, err := a.upsertOperationalIntentReference(ctx, time.Now(), &req.Auth, req.Entityid, "", req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put Operational Intent Reference")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response400: errResp}
		case dsserr.VersionMismatch:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response409: &restapi.AirspaceConflictResponse{
				Message: dsserr.Handle(ctx, err)}}
		case dsserr.MissingOVNs:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.CreateOperationalIntentReferenceResponseSet{Response201: respOK}
}

func (a *Server) UpdateOperationalIntentReference(ctx context.Context, req *restapi.UpdateOperationalIntentReferenceRequest,
) restapi.UpdateOperationalIntentReferenceResponseSet {

	if req.BodyParseError != nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	respOK, respConflict, err := a.upsertOperationalIntentReference(ctx, time.Now(), &req.Auth, req.Entityid, req.Ovn, req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: errResp}
		case dsserr.VersionMismatch:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response409: &restapi.AirspaceConflictResponse{
				Message: dsserr.Handle(ctx, err)}}
		case dsserr.MissingOVNs:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.UpdateOperationalIntentReferenceResponseSet{Response200: respOK}
}

// upsertOperationalIntentReference inserts or updates an Operational Intent.
// If the ovn argument is empty (""), it will attempt to create a new Operational Intent.
func (a *Server) upsertOperationalIntentReference(ctx context.Context, now time.Time, authorizedManager *api.AuthorizationResult, entityid restapi.EntityID, ovn restapi.EntityOVN, params *restapi.PutOperationalIntentReferenceParameters,
) (*restapi.ChangeOperationalIntentReferenceResponse, *restapi.AirspaceConflictResponse, error) {
	// Note: validateAndReturnOIRUpsertParams and checkUpsertPermissionsAndReturnManager could be moved out of this method and only the valid params passed,
	// but this requires some changes in the caller that go beyond the immediate scope of #1088 and can be done later.
	validParams, err := actions.ValidateAndReturnOIRUpsertParams(now, entityid, ovn, params, a.AllowHTTPBaseUrls)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Operational Intent Reference upsert parameters")
	}
	manager, err := actions.CheckUpsertPermissionsAndReturnManager(authorizedManager, validParams.State)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.PermissionDenied, "Caller is not allowed to upsert with the requested state")
	}

	var responseOK *restapi.ChangeOperationalIntentReferenceResponse
	var responseConflict *restapi.AirspaceConflictResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {

		// Get existing OperationalIntent, if any
		old, err := r.GetOperationalIntent(ctx, validParams.ID)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
		}

		// Lock subscriptions based on the cell and subscriptions we're going to use
		// to reduce the number of retries under concurrent load.
		// See issue #1002 for details.
		var subscriptionIds = make([]dssmodels.ID, 0)

		if old != nil && old.SubscriptionID != nil {
			subscriptionIds = append(subscriptionIds, *old.SubscriptionID)
		}

		if !validParams.SubscriptionID.Empty() {
			subscriptionIds = append(subscriptionIds, validParams.SubscriptionID)
		}

		err = r.LockSubscriptionsOnCells(ctx, validParams.Cells, subscriptionIds, validParams.UExtent.StartTime, validParams.UExtent.EndTime)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to acquire lock")
		}

		// Validate the request against the previous OIR
		if err := actions.ValidateUpsertRequestAgainstPreviousOIR(manager, validParams.OVN, old); err != nil {
			return stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Request validation failed")
		}

		var (
			version     = scdmodels.VersionNumber(1)
			pastOVNs    = make([]scdmodels.OVN, 0)
			previousSub *scdmodels.Subscription
		)
		if old != nil {
			version = old.Version + 1
			pastOVNs = append(old.PastOVNs, validParams.OVN)

			// Fetch the previous OIR's subscription if it exists
			if old.SubscriptionID != nil {
				previousSub, err = r.GetSubscription(ctx, *old.SubscriptionID)
				if err != nil {
					return stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
				}
			}
		}

		// Determine if the previous subscription is being replaced and if it will need to be cleaned up
		previousSubIsBeingReplaced := previousSub != nil && validParams.SubscriptionID != previousSub.ID
		removePreviousImplicitSubscription := false
		if previousSubIsBeingReplaced {
			removePreviousImplicitSubscription, err = actions.SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx, r, validParams.ID, previousSub)
			if err != nil {
				return stacktrace.Propagate(err, "Could not determine if previous Subscription can be removed")
			}
		}

		// attachedSub is the subscription that will end up being attached to the OIR
		// it defaults to the previous subscription (which may be nil), and may be updated if required by the parameters
		attachedSub := previousSub
		if validParams.SubscriptionID.Empty() {
			// No subscription ID was provided:
			// check if an implicit subscription should be created, otherwise do nothing
			if validParams.ImplicitSubscription.Requested {
				// Parameters for a new implicit subscription have been passed: we will create
				// a new implicit subscription even if another subscription was attached to this OIR before,
				// regardless of whether it was an implicit subscription or not.
				if attachedSub, err = actions.CreateAndStoreNewImplicitSubscription(ctx, r, manager, validParams); err != nil {
					return stacktrace.Propagate(err, "Failed to create implicit subscription")
				}
			} else {
				// If no subscription ID is provided and no implicit subscription is requested,
				// the OIR should have no attached subscription
				attachedSub = nil
			}
		} else {
			// Attempt to rely on the specified subscription
			// If it is different from the previous subscription, we need to fetch it from the store
			// in order to ensure it correctly covers the OIR.
			// We do the check below in order to avoid re-fetching the subscription if it has not changed
			if attachedSub == nil || previousSubIsBeingReplaced {
				attachedSub, err = r.GetSubscription(ctx, validParams.SubscriptionID)
				if err != nil {
					return stacktrace.Propagate(err, "Unable to get requested Subscription from store")
				}
				if attachedSub == nil {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", validParams.SubscriptionID)
				}
			}

			// We need to confirm that it is owned by the calling manager
			if attachedSub.Manager != manager {
				return stacktrace.Propagate(
					// We do a bit of wrapping gymnastics because the root error message will be sent in the response,
					// and we don't want to include the effective manager in there.
					stacktrace.NewErrorWithCode(
						dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					// The propagation message will end in the logs and help with debugging.
					"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent",
					validParams.SubscriptionID,
					attachedSub.Manager,
					manager,
				)
			}

			// We need to ensure the subscription covers the OIR's geo-temporal extent
			attachedSub, err = actions.EnsureSubscriptionCoversOIR(ctx, r, attachedSub, validParams)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to ensure subscription covers OIR")
			}
		}

		if validParams.State.RequiresKey() {
			responseConflict, err = actions.ValidateKeyAndProvideConflictResponse(ctx, r, manager, validParams, attachedSub)
			if err != nil {
				return stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Failed to validate key")
			}
		}

		// Construct the new OperationalIntent
		op := validParams.ToOIR(manager, attachedSub, version, pastOVNs)

		// Upsert the OperationalIntent
		op, err = r.UpsertOperationalIntent(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
		}

		// Check if the previously attached subscription should be removed
		if removePreviousImplicitSubscription {
			err = r.DeleteSubscription(ctx, previousSub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete previous implicit Subscription")
			}
		}

		notifyVolume, err := actions.ComputeNotificationVolume(old, validParams.UExtent)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to compute notification volume")
		}

		// Notify relevant Subscriptions
		subsToNotify, err := r.IncrementNotificationIndicesForOperationalIntents(ctx, notifyVolume)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to notify relevant Subscriptions")
		}

		// Return response to client
		responseOK = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *op.ToRest(),
			Subscribers:                makeSubscribersToNotify(subsToNotify),
		}

		return nil
	}

	_, err = a.Store.Transact(ctx, dssstore.NewFuncOperation(action))
	if err != nil {
		return nil, responseConflict, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return responseOK, responseConflict, nil
}
