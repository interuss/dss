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

		if old.OVN != ovn {
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", old.OVN, ovn)
		}

		// Get the Subscription supporting the OperationalIntent, if one is defined
		var sub *scdmodels.Subscription
		removeImplicitSubscription := false
		if old.SubscriptionID != nil {
			sub, err = r.GetSubscription(ctx, *old.SubscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
			}
			if sub == nil {
				return stacktrace.NewError("OperationalIntent's Subscription missing from repo")
			}

			if sub.ImplicitSubscription {
				// Get the Subscription's dependent OperationalIntents
				dependentOps, err := r.GetDependentOperationalIntents(ctx, sub.ID)
				if err != nil {
					return stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
				}
				if len(dependentOps) == 0 {
					return stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
				} else if len(dependentOps) == 1 {
					removeImplicitSubscription = true
				}
			}
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

		subsToNotify, err := getRelevantSubscriptionsAndIncrementIndices(ctx, r, notifyVolume)
		if err != nil {
			return stacktrace.Propagate(err, "could not obtain relevant subscriptions")
		}

		// Delete OperationalIntent from repo
		if err := r.DeleteOperationalIntent(ctx, id); err != nil {
			return stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
		}

		// removeImplicitSubscription is only true if the OIR had a subscription defined
		if removeImplicitSubscription {
			// Automatically remove a now-unused implicit Subscription
			err = r.DeleteSubscription(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
			}
		}

		// Return response to client
		response = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *old.ToRest(),
			Subscribers:                makeSubscribersToNotify(subsToNotify),
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

type validOIRParams struct {
	id             dssmodels.ID
	ovn            restapi.EntityOVN
	state          scdmodels.OperationalIntentState
	extents        []*dssmodels.Volume4D
	uExtent        *dssmodels.Volume4D
	cells          s2.CellUnion
	subscriptionID dssmodels.ID
	key            map[scdmodels.OVN]bool
}

// validateAndReturnUpsertParams checks that the parameters for an Operational Intent Reference upsert are valid.
// Note that this does NOT check for anything related to access controls: any error returned should be labeled
// as a dsserr.BadRequest.
func validateAndReturnUpsertParams(
	entityid restapi.EntityID,
	ovn restapi.EntityOVN,
	params *restapi.PutOperationalIntentReferenceParameters,
	allowHTTPBaseUrls bool,
) (*validOIRParams, error) {

	valid := &validOIRParams{}
	var err error

	valid.id, err = dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, stacktrace.NewError("Invalid ID format: `%s`", entityid)
	}

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewError("Missing required UssBaseUrl")
	}

	if !allowHTTPBaseUrls {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to validate base URL")
		}
	}

	valid.state = scdmodels.OperationalIntentState(params.State)
	if !valid.state.IsValidInDSS() {
		return nil, stacktrace.NewError("Invalid OperationalIntent state: %s", params.State)
	}

	valid.extents = make([]*dssmodels.Volume4D, len(params.Extents))

	for idx, extent := range params.Extents {
		cExtent, err := dssmodels.Volume4DFromSCDRest(&extent)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to parse extent %d", idx)
		}
		valid.extents[idx] = cExtent
	}

	valid.uExtent, err = dssmodels.UnionVolumes4D(valid.extents...)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to union extents")
	}

	if valid.uExtent.StartTime == nil {
		return nil, stacktrace.NewError("Missing time_start from extents")
	}
	if valid.uExtent.EndTime == nil {
		return nil, stacktrace.NewError("Missing time_end from extents")
	}

	if time.Now().After(*valid.uExtent.EndTime) {
		return nil, stacktrace.NewError("OperationalIntents may not end in the past")
	}

	valid.cells, err = valid.uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	if valid.uExtent.EndTime.Before(*valid.uExtent.StartTime) {
		return nil, stacktrace.NewError("End time is past the start time")
	}

	if ovn == "" && params.State != restapi.OperationalIntentState_Accepted {
		return nil, stacktrace.NewError("Invalid state for initial version: `%s`", params.State)
	}
	valid.ovn = ovn

	if params.SubscriptionId != nil {
		valid.subscriptionID, err = dssmodels.IDFromOptionalString(string(*params.SubscriptionId))
		if err != nil {
			return nil, stacktrace.NewError("Invalid ID format for Subscription ID: `%s`", *params.SubscriptionId)
		}
	}

	if params.NewSubscription != nil {
		// The spec states that NewSubscription.UssBaseUrl is required and an empty value
		// makes no sense, so we will fail if an implicit subscription is requested but the base url is empty
		if params.NewSubscription.UssBaseUrl == "" {
			return nil, stacktrace.NewError("Missing required USS base url for new subscription (in parameters for implicit subscription)")
		}
		// If an implicit subscription is requested, the Subscription ID cannot be present.
		if params.SubscriptionId != nil {
			return nil, stacktrace.NewError("Cannot provide both a Subscription ID and request an implicit subscription")
		}
	}

	// Check if a subscription is required for this request:
	// OIRs in an accepted state do not need a subscription.
	if valid.state.RequiresSubscription() &&
		valid.subscriptionID.Empty() &&
		(params.NewSubscription == nil ||
			params.NewSubscription.UssBaseUrl == "") {
		return nil, stacktrace.NewError("Provided Operational Intent Reference state `%s` requires either a subscription ID or information to create an implicit subscription", valid.state)
	}

	// Construct a hash set of OVNs as the key
	valid.key = map[scdmodels.OVN]bool{}
	if params.Key != nil {
		for _, ovn := range *params.Key {
			valid.key[scdmodels.OVN(ovn)] = true
		}
	}

	return valid, nil
}

// checkUpsertPermissions verifies that the client has the necessary permissions to upsert an Operational Intent with the requested state.
func checkUpsertPermissionsAndReturnManager(authorizedManager *api.AuthorizationResult, requestedState scdmodels.OperationalIntentState) (dssmodels.Manager, error) {
	if authorizedManager.ClientID == nil {
		return "", stacktrace.NewError("Missing manager")
	}
	hasCMSARole := auth.HasScope(authorizedManager.Scopes, restapi.UtmConformanceMonitoringSaScope)
	if requestedState.RequiresCMSA() && !hasCMSARole {
		return "", stacktrace.NewError("Missing `%s` Conformance Monitoring for Situational Awareness scope to transition to CMSA state: %s (see SCD0100)", restapi.UtmConformanceMonitoringSaScope, requestedState)
	}
	return dssmodels.Manager(*authorizedManager.ClientID), nil
}

// validateUpsertRequestAgainstPreviousOIR checks that the client requesting an OIR upsert has the necessary permissions and that the request is valid.
// On success, the version of the OIR is returned:
//   - upon initial creation (if no previous OIR exists), it is 0
//   - otherwise, it is the version of the previous OIR
func validateUpsertRequestAgainstPreviousOIR(
	requestingManager dssmodels.Manager,
	providedOVN restapi.EntityOVN,
	previousOIR *scdmodels.OperationalIntent,
) error {

	if previousOIR != nil {
		if previousOIR.Manager != requestingManager {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to modify", previousOIR.Manager, requestingManager)
		}
		if previousOIR.OVN != scdmodels.OVN(providedOVN) {
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", previousOIR.OVN, providedOVN)
		}

		return nil
	}

	if providedOVN != "" {
		return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent does not exist and therefore is not version %s", providedOVN)
	}

	return nil
}

// computeNotificationVolume computes the volume that needs to be queried for subscriptions
// given the requested extent and the (possibly nil) previous operational intent.
// The returned volume is either the union of the requested extent and the previous OIR's extent, or just the requested extent
// if the previous OIR is nil.
func computeNotificationVolume(
	previousOIR *scdmodels.OperationalIntent,
	requestedExtent *dssmodels.Volume4D) (*dssmodels.Volume4D, error) {

	if previousOIR == nil {
		return requestedExtent, nil
	}

	// Compute total affected Volume4D for notification purposes
	oldVolume := &dssmodels.Volume4D{
		StartTime: previousOIR.StartTime,
		EndTime:   previousOIR.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: previousOIR.AltitudeUpper,
			AltitudeLo: previousOIR.AltitudeLower,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return previousOIR.Cells, nil
			}),
		},
	}
	notifyVolume, err := dssmodels.UnionVolumes4D(requestedExtent, oldVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error constructing 4D volumes union")
	}

	return notifyVolume, nil
}

// getRelevantSubscriptionsAndIncrementIndices retrieves the subscriptions relevant to the passed volume and increments their notification indices
// before returning them.
func getRelevantSubscriptionsAndIncrementIndices(
	ctx context.Context,
	r repos.Repository,
	notifyVolume *dssmodels.Volume4D,
) (repos.Subscriptions, error) {

	// Find Subscriptions that may need to be notified
	allsubs, err := r.SearchSubscriptions(ctx, notifyVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to search for impacted subscriptions.")
	}

	// Limit Subscription notifications to only those interested in OperationalIntents
	subs := repos.Subscriptions{}
	for _, sub := range allsubs {
		if sub.NotifyForOperationalIntents {
			subs = append(subs, sub)
		}
	}

	// Increment notification indices for relevant Subscriptions
	if err := subs.IncrementNotificationIndices(ctx, r); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to increment notification indices of relevant subscriptions")
	}

	return subs, nil
}

// validateKeyAndProvideConflictResponse ensures that the provided key contains all the necessary OVNs relevant for the area covered by the OperationalIntent.
// - If all required keys are provided, (nil, nil) will be returned.
// - If keys are missing, the conflict response to be sent back as well as an error with the dsserr.MissingOVNs code will be returned.
// - In case of any other error, (nil, error) will be returned.
func validateKeyAndProvideConflictResponse(
	ctx context.Context,
	r repos.Repository,
	requestingManager dssmodels.Manager,
	params *validOIRParams,
	attachedSubscription *scdmodels.Subscription,
) (*restapi.AirspaceConflictResponse, error) {

	// Identify OperationalIntents missing from the key
	var missingOps []*scdmodels.OperationalIntent
	relevantOps, err := r.SearchOperationalIntents(ctx, params.uExtent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to SearchOperations")
	}
	for _, relevantOp := range relevantOps {
		_, ok := params.key[relevantOp.OVN]
		// Note: The OIR being mutated does not need to be specified in the key:
		if !ok && relevantOp.RequiresKey() && relevantOp.ID != params.id {
			missingOps = append(missingOps, relevantOp)
		}
	}

	// Identify Constraints missing from the key
	var missingConstraints []*scdmodels.Constraint
	if attachedSubscription != nil && attachedSubscription.NotifyForConstraints {
		constraints, err := r.SearchConstraints(ctx, params.uExtent)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to SearchConstraints")
		}
		for _, relevantConstraint := range constraints {
			if _, ok := params.key[relevantConstraint.OVN]; !ok {
				missingConstraints = append(missingConstraints, relevantConstraint)
			}
		}
	}

	// If the client is missing some OVNs, provide the pointers to the
	// information they need
	if len(missingOps) > 0 || len(missingConstraints) > 0 {
		msg := "Current OVNs not provided for one or more OperationalIntents or Constraints"
		responseConflict := &restapi.AirspaceConflictResponse{Message: &msg}

		if len(missingOps) > 0 {
			responseConflict.MissingOperationalIntents = new([]restapi.OperationalIntentReference)
			for _, missingOp := range missingOps {
				p := missingOp.ToRest()
				// We scrub the OVNs of entities not owned by the requesting manager to make sure
				// they have really contacted the managing USS
				if missingOp.Manager != requestingManager {
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
				// We scrub the OVNs of entities not owned by the requesting manager to make sure
				// they have really contacted the managing USS
				if missingConstraint.Manager != requestingManager {
					noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
					c.Ovn = &noOvnPhrase
				}
				*responseConflict.MissingConstraints = append(*responseConflict.MissingConstraints, *c)
			}
		}

		return responseConflict, stacktrace.NewErrorWithCode(dsserr.MissingOVNs, "Missing OVNs: %v", msg)
	}

	return nil, nil
}

// upsertOperationalIntentReference inserts or updates an Operational Intent.
// If the ovn argument is empty (""), it will attempt to create a new Operational Intent.
func (a *Server) upsertOperationalIntentReference(ctx context.Context, authorizedManager *api.AuthorizationResult, entityid restapi.EntityID, ovn restapi.EntityOVN, params *restapi.PutOperationalIntentReferenceParameters,
) (*restapi.ChangeOperationalIntentReferenceResponse, *restapi.AirspaceConflictResponse, error) {
	// Note: validateAndReturnUpsertParams and checkUpsertPermissionsAndReturnManager could be moved out of this method and only the valid params passed,
	// but this requires some changes in the caller that go beyond the immediate scope of #1088 and can be done later.
	validParams, err := validateAndReturnUpsertParams(entityid, ovn, params, a.AllowHTTPBaseUrls)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Operational Intent Reference upsert parameters")
	}
	manager, err := checkUpsertPermissionsAndReturnManager(authorizedManager, validParams.state)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.PermissionDenied, "Caller is not allowed to upsert with the requested state")
	}

	var responseOK *restapi.ChangeOperationalIntentReferenceResponse
	var responseConflict *restapi.AirspaceConflictResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Lock subscriptions based on the cell to reduce the number of retries under concurrent load.
		// See issue #1002 for details.
		err = r.LockSubscriptionsOnCells(ctx, validParams.cells)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to acquire lock")
		}

		// Get existing OperationalIntent, if any
		old, err := r.GetOperationalIntent(ctx, validParams.id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
		}
		// Validate the request against the previous OIR and return the current version
		// (upon new OIR creation, version is 0)
		if err := validateUpsertRequestAgainstPreviousOIR(manager, validParams.ovn, old); err != nil {
			return stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Request validation failed")
		}

		// For an OIR being created, version starts at 0
		version := int32(0)
		if old != nil {
			version = int32(old.Version)
		}

		var sub *scdmodels.Subscription
		if validParams.subscriptionID.Empty() {
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

				subToUpsert := scdmodels.Subscription{
					ID:                          dssmodels.ID(uuid.New().String()),
					Manager:                     manager,
					StartTime:                   validParams.uExtent.StartTime,
					EndTime:                     validParams.uExtent.EndTime,
					AltitudeLo:                  validParams.uExtent.SpatialVolume.AltitudeLo,
					AltitudeHi:                  validParams.uExtent.SpatialVolume.AltitudeHi,
					Cells:                       validParams.cells,
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
			sub, err = r.GetSubscription(ctx, validParams.subscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get Subscription")
			}
			if sub == nil {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", validParams.subscriptionID)
			}
			if sub.Manager != manager {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent", validParams.subscriptionID, sub.Manager, manager)
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*validParams.uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = validParams.uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*validParams.uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = validParams.uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
				}
			}
			if !sub.Cells.Contains(validParams.cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, validParams.cells)
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

		if validParams.state.RequiresKey() {
			responseConflict, err = validateKeyAndProvideConflictResponse(ctx, r, manager, validParams, sub)
			if err != nil {
				return stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Failed to validate key")
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
			ID:      validParams.id,
			Manager: manager,
			Version: scdmodels.VersionNumber(version + 1),

			StartTime:     validParams.uExtent.StartTime,
			EndTime:       validParams.uExtent.EndTime,
			AltitudeLower: validParams.uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: validParams.uExtent.SpatialVolume.AltitudeHi,
			Cells:         validParams.cells,

			USSBaseURL:     string(params.UssBaseUrl),
			SubscriptionID: subID,
			State:          validParams.state,
		}
		err = op.ValidateTimeRange()
		if err != nil {
			return stacktrace.Propagate(err, "Error validating time range")
		}

		// Upsert the OperationalIntent
		op, err = r.UpsertOperationalIntent(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
		}

		notifyVolume, err := computeNotificationVolume(old, validParams.uExtent)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to compute notification volume")
		}

		// Notify relevant Subscriptions
		subsToNotify, err := getRelevantSubscriptionsAndIncrementIndices(ctx, r, notifyVolume)
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

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, responseConflict, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return responseOK, responseConflict, nil
}
