package scd

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

// subscriptionIsOnlyAttachedToOIR will check if:
// - the subscription exists and is implicit
// - the subscription is attached to the specified operational intent
// - the subscription is not attached to any other operational intent
//
// This is to be used in contexts where an implicit subscription may need to be cleaned up: if true is returned,
// the subscription can be safely removed after the operational intent is deleted or attached to another subscription.
//
// NOTE: this should eventually be pushed down to CRDB as part of the queries being executed in the callers of this method.
//
//	See https://github.com/interuss/dss/issues/1059 for more details
func subscriptionIsOnlyAttachedToOIR(ctx context.Context, r repos.Repository, oirID, subscriptionID *dssmodels.ID) (bool, error) {
	// Get the Subscription supporting the OperationalIntent, if one is defined
	if subscriptionID != nil {
		sub, err := r.GetSubscription(ctx, *subscriptionID)
		if err != nil {
			return false, stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
		}
		if sub == nil {
			return false, stacktrace.NewError("OperationalIntent's Subscription missing from repo")
		}

		if sub.ImplicitSubscription {
			// Get the Subscription's dependent OperationalIntents
			dependentOps, err := r.GetDependentOperationalIntents(ctx, sub.ID)
			if err != nil {
				return false, stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
			}
			if len(dependentOps) == 0 {
				return false, stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
			} else if len(dependentOps) == 1 && dependentOps[0] == *oirID {
				return true, nil
			}
		}
	}
	return false, nil
}

// DeleteOperationalIntentReference deletes a single operational intent ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationalIntentReference(ctx context.Context, req *restapi.DeleteOperationalIntentReferenceRequest,
) restapi.DeleteOperationalIntentReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Retrieve OperationalIntent ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
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

	var response *restapi.ChangeOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get OperationalIntent to delete
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationIntent from repo")
		}
		if old == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		// Validate deletion request
		if old.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
		}

<<<<<<< HEAD
        if old.OVN != ovn {
            return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
                "Current version is %s but client specified version %s", old.OVN, ovn)
        }

		removeImplicitSubscription, err := subscriptionCanBeRemoved(ctx, r, old.SubscriptionID)
=======
		removeImplicitSubscription, err := subscriptionIsOnlyAttachedToOIR(ctx, r, &id, old.SubscriptionID)
>>>>>>> fdd5e11 (rename and make subscriptionCanBeRemoved more specific)
		if err != nil {
			return stacktrace.Propagate(err, "Could not determine if Subscription can be removed")
		}

		// Find Subscriptions that may overlap the OperationalIntent's Volume4D
		allsubs, err := r.SearchSubscriptions(ctx, &dssmodels.Volume4D{
			StartTime: old.StartTime,
			EndTime:   old.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeHi: old.AltitudeUpper,
				AltitudeLo: old.AltitudeLower,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return old.Cells, nil
				}),
			}})
		if err != nil {
			return stacktrace.Propagate(err, "Unable to search Subscriptions in repo")
		}

		// Limit Subscription notifications to only those interested in OperationalIntents
		subs := repos.Subscriptions{}
		for _, s := range allsubs {
			if s.NotifyForOperationalIntents {
				subs = append(subs, s)
			}
		}

		// Increment notification indices for Subscriptions to be notified
		if err := subs.IncrementNotificationIndices(ctx, r); err != nil {
			return stacktrace.Propagate(err, "Unable to increment notification indices")
		}

		// Delete OperationalIntent from repo
		if err := r.DeleteOperationalIntent(ctx, id); err != nil {
			return stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
		}

		// removeImplicitSubscription is only true if the OIR had a subscription defined
		if removeImplicitSubscription {
			// Automatically remove a now-unused implicit Subscription
			err = r.DeleteSubscription(ctx, *old.SubscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
			}
		}

		// Return response to client
		response = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *old.ToRest(),
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
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
	if req.Auth.Error != nil {
		resp := restapi.GetOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.GetOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		op, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationalIntent from repo")
		}
		if op == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			op.OVN = scdmodels.NoOvnPhrase
		}

		response = &restapi.GetOperationalIntentReferenceResponse{
			OperationalIntentReference: *op.ToRest(),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
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
	if req.Auth.Error != nil {
		resp := restapi.QueryOperationalIntentReferencesResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

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
	vol4, err := dssmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.QueryOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		ops, err := r.SearchOperationalIntents(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
		}

		// Create response for client
		response = &restapi.QueryOperationalIntentReferenceResponse{
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

		return nil
	}

	err = a.Store.Transact(ctx, action)
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
	if req.Auth.Error != nil {
		resp := restapi.CreateOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	respOK, respConflict, err := a.upsertOperationalIntentReference(ctx, &req.Auth, req.Entityid, "", req.Body)
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
	if req.Auth.Error != nil {
		resp := restapi.UpdateOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	respOK, respConflict, err := a.upsertOperationalIntentReference(ctx, &req.Auth, req.Entityid, req.Ovn, req.Body)
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
func (a *Server) upsertOperationalIntentReference(ctx context.Context, authorizedManager *api.AuthorizationResult, entityid restapi.EntityID, ovn restapi.EntityOVN, params *restapi.PutOperationalIntentReferenceParameters,
) (*restapi.ChangeOperationalIntentReferenceResponse, *restapi.AirspaceConflictResponse, error) {
	if authorizedManager.ClientID == nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager")
	}
	manager := dssmodels.Manager(*authorizedManager.ClientID)

	id, err := dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", entityid)
	}

	var (
		extents = make([]*dssmodels.Volume4D, len(params.Extents))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	if !a.AllowHTTPBaseUrls {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	state := scdmodels.OperationalIntentState(params.State)
	if !state.IsValidInDSS() {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid OperationalIntent state: %s", params.State)
	}
	hasCMSARole := auth.HasScope(authorizedManager.Scopes, restapi.UtmConformanceMonitoringSaScope)
	if state.RequiresCMSA() && !hasCMSARole {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing `%s` Conformance Monitoring for Situational Awareness scope to transition to CMSA state: %s (see SCD0100)", restapi.UtmConformanceMonitoringSaScope, params.State)
	}

	for idx, extent := range params.Extents {
		cExtent, err := dssmodels.Volume4DFromSCDRest(&extent)
		if err != nil {
			return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extent %d", idx)
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to union extents")
	}

	if uExtent.StartTime == nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_end from extents")
	}

	if time.Now().After(*uExtent.EndTime) {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "OperationalIntents may not end in the past")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	if uExtent.EndTime.Before(*uExtent.StartTime) {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is past the start time")
	}

	if ovn == "" && params.State != restapi.OperationalIntentState_Accepted {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid state for initial version: `%s`", params.State)
	}

	subscriptionID := dssmodels.ID("")
	if params.SubscriptionId != nil {
		subscriptionID, err = dssmodels.IDFromOptionalString(string(*params.SubscriptionId))
		if err != nil {
			return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format for Subscription ID: `%s`", *params.SubscriptionId)
		}
	}

	// Check if a subscription is required for this request:
	// OIRs in an accepted state do not need a subscription.
	if state.RequiresSubscription() &&
		subscriptionID.Empty() &&
		(params.NewSubscription == nil ||
			params.NewSubscription.UssBaseUrl == "") {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Provided Operational Intent Reference state `%s` requires either a subscription ID or information to create an implicit subscription", state)
	}

	var responseOK *restapi.ChangeOperationalIntentReferenceResponse
	var responseConflict *restapi.AirspaceConflictResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		var version int32 // Version of the Operational Intent (0 means creation requested).

		// Lock subscriptions based on the cell to reduce the number of retries under concurrent load.
		// See issue #1002 for details.
		err = r.LockSubscriptionsOnCells(ctx, cells)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to acquire lock")
		}

		// Get existing OperationalIntent, if any, and validate request
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
		}

		var previousSubscriptionID *dssmodels.ID
		if old != nil {
			if old.Manager != manager {
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
					"OperationalIntent owned by %s, but %s attempted to modify", old.Manager, manager)
			}
			if old.OVN != scdmodels.OVN(ovn) {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
					"Current version is %s but client specified version %s", old.OVN, ovn)
			}

			version = int32(old.Version)
			previousSubscriptionID = old.SubscriptionID
		} else {
			if ovn != "" {
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent does not exist and therefore is not version %s", ovn)
			}

			version = 0
		}

		var sub *scdmodels.Subscription
		removePreviousImplicitSubscription := false
		if subscriptionID.Empty() {
			// Create an implicit subscription if the implicit subscription params are set:
			// for situations where these params are required but have not been set,
			// an error will have been returned earlier.
			// If they are not set at this point, continue without creating an implicit subscription.
			if params.NewSubscription != nil && params.NewSubscription.UssBaseUrl != "" {
				if !a.AllowHTTPBaseUrls {
					err := scdmodels.ValidateUSSBaseURL(string(params.NewSubscription.UssBaseUrl))
					if err != nil {
						return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate USS base URL")
					}
				}

				removePreviousImplicitSubscription, err = subscriptionIsOnlyAttachedToOIR(ctx, r, &id, previousSubscriptionID)
				if err != nil {
					return stacktrace.Propagate(err, "Could not determine if previous Subscription can be removed")
				}

				// Note: parameters for a new implicit subscription have been passed, so we will create
				// a new implicit subscription even if another subscription was attaches to this OIR before,
				// (and regardless of whether it was an implicit subscription or not).
				subToUpsert := scdmodels.Subscription{
					ID:                          dssmodels.ID(uuid.New().String()),
					Manager:                     manager,
					StartTime:                   uExtent.StartTime,
					EndTime:                     uExtent.EndTime,
					AltitudeLo:                  uExtent.SpatialVolume.AltitudeLo,
					AltitudeHi:                  uExtent.SpatialVolume.AltitudeHi,
					Cells:                       cells,
					USSBaseURL:                  string(params.NewSubscription.UssBaseUrl),
					NotifyForOperationalIntents: true,
					ImplicitSubscription:        true,
				}
				if params.NewSubscription.NotifyForConstraints != nil {
					subToUpsert.NotifyForConstraints = *params.NewSubscription.NotifyForConstraints
				}

				sub, err = r.UpsertSubscription(ctx, &subToUpsert)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to create implicit subscription")
				}
			}

		} else {
			// Use existing Subscription
			sub, err = r.GetSubscription(ctx, subscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get Subscription")
			}
			if sub == nil {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", subscriptionID)
			}
			if sub.Manager != dssmodels.Manager(manager) {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent", subscriptionID, sub.Manager, manager)
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
				}
			}
			if !sub.Cells.Contains(cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, cells)
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the OperationalIntent")
				}
			}
			if updateSub {
				sub, err = r.UpsertSubscription(ctx, sub)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to update existing Subscription")
				}
			}
		}

		if state.RequiresKey() {
			// Construct a hash set of OVNs as the key
			key := map[scdmodels.OVN]bool{}
			if params.Key != nil {
				for _, ovn := range *params.Key {
					key[scdmodels.OVN(ovn)] = true
				}
			}

			// Identify OperationalIntents missing from the key
			var missingOps []*scdmodels.OperationalIntent
			relevantOps, err := r.SearchOperationalIntents(ctx, uExtent)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to SearchOperations")
			}
			for _, relevantOp := range relevantOps {
				_, ok := key[relevantOp.OVN]
				// Note: The OIR being mutated does not need to be specified in the key:
				if !ok && relevantOp.RequiresKey() && relevantOp.ID != id {
					if relevantOp.Manager != dssmodels.Manager(manager) {
						relevantOp.OVN = scdmodels.NoOvnPhrase
					}
					missingOps = append(missingOps, relevantOp)
				}
			}

			// Identify Constraints missing from the key
			var missingConstraints []*scdmodels.Constraint
			if sub != nil && sub.NotifyForConstraints {
				constraints, err := r.SearchConstraints(ctx, uExtent)
				if err != nil {
					return stacktrace.Propagate(err, "Unable to SearchConstraints")
				}
				for _, relevantConstraint := range constraints {
					if _, ok := key[relevantConstraint.OVN]; !ok {
						if relevantConstraint.Manager != dssmodels.Manager(manager) {
							relevantConstraint.OVN = scdmodels.NoOvnPhrase
						}
						missingConstraints = append(missingConstraints, relevantConstraint)
					}
				}
			}

			// If the client is missing some OVNs, provide the pointers to the
			// information they need
			if len(missingOps) > 0 || len(missingConstraints) > 0 {
				msg := "Current OVNs not provided for one or more OperationalIntents or Constraints"
				responseConflict = &restapi.AirspaceConflictResponse{Message: &msg}

				if len(missingOps) > 0 {
					responseConflict.MissingOperationalIntents = new([]restapi.OperationalIntentReference)
					for _, missingOp := range missingOps {
						p := missingOp.ToRest()
						if missingOp.Manager != manager {
							noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
							p.Ovn = &noOvnPhrase
						}
						*responseConflict.MissingOperationalIntents = append(*responseConflict.MissingOperationalIntents, *p)
					}
				}

				if len(missingConstraints) > 0 {
					responseConflict.MissingConstraints = new([]restapi.ConstraintReference)
					for _, missingConstraint := range missingConstraints {
						c := missingConstraint.ToRest()
						if missingConstraint.Manager != manager {
							noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
							c.Ovn = &noOvnPhrase
						}
						*responseConflict.MissingConstraints = append(*responseConflict.MissingConstraints, *c)
					}
				}

				return stacktrace.NewErrorWithCode(dsserr.MissingOVNs, "Missing OVNs: %v", msg)
			}
		}

		// For OIR's in the accepted state, we may not have a subscription available,
		// in such cases the subscription ID on scdmodels.OperationalIntent will be nil
		// and will be replaced with the 'NullV4UUID' when sent over to a client.
		var subID *dssmodels.ID = nil
		if sub != nil {
			subID = &sub.ID
		}

		// Construct the new OperationalIntent
		op := &scdmodels.OperationalIntent{
			ID:      id,
			Manager: dssmodels.Manager(manager),
			Version: scdmodels.VersionNumber(version + 1),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
			Cells:         cells,

			USSBaseURL:     string(params.UssBaseUrl),
			SubscriptionID: subID,
			State:          state,
		}
		err = op.ValidateTimeRange()
		if err != nil {
			return stacktrace.Propagate(err, "Error validating time range")
		}

		// Compute total affected Volume4D for notification purposes
		var notifyVol4 *dssmodels.Volume4D
		if old == nil {
			notifyVol4 = uExtent
		} else {
			oldVol4 := &dssmodels.Volume4D{
				StartTime: old.StartTime,
				EndTime:   old.EndTime,
				SpatialVolume: &dssmodels.Volume3D{
					AltitudeHi: old.AltitudeUpper,
					AltitudeLo: old.AltitudeLower,
					Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
						return old.Cells, nil
					}),
				}}
			notifyVol4, err = dssmodels.UnionVolumes4D(uExtent, oldVol4)
			if err != nil {
				return stacktrace.Propagate(err, "Error constructing 4D volumes union")
			}
		}

		// Upsert the OperationalIntent
		op, err = r.UpsertOperationalIntent(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
		}

		// Check if the previously attached subscription should be removed
		if removePreviousImplicitSubscription {
			err = r.DeleteSubscription(ctx, *previousSubscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete previous implicit Subscription")
			}
		}

		// Find Subscriptions that may need to be notified
		allsubs, err := r.SearchSubscriptions(ctx, notifyVol4)
		if err != nil {
			return err
		}

		// Limit Subscription notifications to only those interested in OperationalIntents
		subs := repos.Subscriptions{}
		for _, sub := range allsubs {
			if sub.NotifyForOperationalIntents {
				subs = append(subs, sub)
			}
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return err
		}

		// Return response to client
		responseOK = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *op.ToRest(),
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, responseConflict, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return responseOK, responseConflict, nil
}
