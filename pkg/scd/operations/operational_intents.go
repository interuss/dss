package operations

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.GetOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetOperationalIntentReferenceRequest],
		Execute:    executeGetOperationalIntentReference,
		IsReadOnly: true,
	}
	Registry[restapi.QueryOperationalIntentReferencesOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QueryOperationalIntentReferencesRequest],
		Execute:    executeQueryOperationalIntentReferences,
		IsReadOnly: true,
	}
	Registry[restapi.DeleteOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteOperationalIntentReferenceRequest],
		Execute: executeDeleteOperationalIntentReference,
	}
	Registry[restapi.CreateOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.CreateOperationalIntentReferenceRequest],
		Execute: executePutOperationalIntentReference,
	}
	Registry[restapi.UpdateOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.UpdateOperationalIntentReferenceRequest],
		Execute: executePutOperationalIntentReference,
	}
}

// SubscriptionIsImplicitAndOnlyAttachedToOIR will check if:
// - the subscription is defined and is implicit
// - the subscription is attached to the specified operational intent
// - the subscription is not attached to any other operational intent
//
// This is to be used in contexts where an implicit subscription may need to be cleaned up: if true is returned,
// the subscription can be safely removed after the operational intent is deleted or attached to another subscription.
//
// NOTE: this should eventually be pushed down the datastore as part of the queries being executed in the callers of this method.
//
//	See https://github.com/interuss/dss/issues/1059 for more details
func SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx context.Context, r repos.Repository, oirID dssmodels.ID, subscription *scdmodels.Subscription) (bool, error) {
	if subscription == nil {
		return false, nil
	}
	if !subscription.ImplicitSubscription {
		return false, nil
	}
	// Get the Subscription's dependent OperationalIntents
	dependentOps, err := r.GetDependentOperationalIntents(ctx, subscription.ID)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
	}
	if len(dependentOps) == 0 {
		return false, stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
	} else if len(dependentOps) == 1 && dependentOps[0] == oirID {
		return true, nil
	}
	return false, nil
}

// GetRelevantSubscriptionsAndIncrementIndices retrieves the subscriptions relevant to the passed volume and increments their notification indices
// before returning them.
func GetRelevantSubscriptionsAndIncrementIndices(
	ctx context.Context,
	r repos.Repository,
	notifyVolume *dssmodels.Volume4D,
) (repos.Subscriptions, error) {

	// Find the Subscriptions interested in OperationalIntents and increment their
	// notification indices
	subs, err := r.IncrementNotificationIndicesForOperationalIntents(ctx, notifyVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to increment notification indices of relevant subscriptions")
	}

	return subs, nil
}

// executeDeleteOperationalIntentReference deletes a single operational intent ref for a given ID
// at the specified version.
func executeDeleteOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.DeleteOperationalIntentReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteOperationalIntentReferenceOperationID)
	}

	// Retrieve OperationalIntent ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	// Get OperationalIntent to delete
	old, err := repo.GetOperationalIntent(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get OperationIntent from repo")
	}
	if old == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
	}

	// Validate deletion request
	if old.Manager != dssmodels.Manager(*req.Auth.ClientID) {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"OperationalIntent owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	}

	if old.OVN != scdmodels.OVN(req.Ovn) {
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, scdmodels.OVN(req.Ovn))
	}

	// Lock subscriptions based on the cell and subscriptions we're going to use
	// to reduce the number of retries under concurrent load.
	// See issue #1002 for details.
	var subscriptionIds = make([]dssmodels.ID, 0)

	if old.SubscriptionID != nil {
		subscriptionIds = append(subscriptionIds, *old.SubscriptionID)
	}

	err = repo.LockSubscriptionsOnCells(ctx, old.Cells, subscriptionIds, old.StartTime, old.EndTime)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to acquire lock")
	}

	// Get the Subscription supporting the OperationalIntent, if one is defined
	var previousSubscription *scdmodels.Subscription
	if old.SubscriptionID != nil {
		previousSubscription, err = repo.GetSubscription(ctx, *old.SubscriptionID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
		}
		if previousSubscription == nil {
			return nil, stacktrace.NewError("OperationalIntent's Subscription missing from repo")
		}
	}

	removeImplicitSubscription, err := SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx, repo, id, previousSubscription)
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

	subsToNotify, err := GetRelevantSubscriptionsAndIncrementIndices(ctx, repo, notifyVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "could not obtain relevant subscriptions")
	}

	// Delete OperationalIntent from repo
	if err := repo.DeleteOperationalIntent(ctx, id); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
	}

	// removeImplicitSubscription is only true if the OIR had a subscription defined
	if removeImplicitSubscription {
		// Automatically remove a now-unused implicit Subscription
		err = repo.DeleteSubscription(ctx, previousSubscription.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
		}
	}

	// Return response to client
	return &restapi.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: *old.ToRest(),
		Subscribers:                makeSubscribersToNotify(subsToNotify),
	}, nil
}

func executeGetOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.GetOperationalIntentReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetOperationalIntentReferenceOperationID)
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	op, err := repo.GetOperationalIntent(ctx, id)
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

func executeQueryOperationalIntentReferences(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.QueryOperationalIntentReferencesRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QueryOperationalIntentReferencesOperationID)
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	// Perform search query on Store
	ops, err := repo.SearchOperationalIntents(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
	}

	// Create response for client
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

// CheckUpsertPermissionsAndReturnManager verifies that the client has the necessary permissions to upsert an Operational Intent with the requested state.
func CheckUpsertPermissionsAndReturnManager(authorizedManager *api.AuthorizationResult, requestedState scdmodels.OperationalIntentState) (dssmodels.Manager, error) {
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
	providedOVN scdmodels.OVN,
	previousOIR *scdmodels.OperationalIntent,
) error {

	if previousOIR != nil {
		if previousOIR.Manager != requestingManager {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to modify", previousOIR.Manager, requestingManager)
		}
		if previousOIR.OVN != providedOVN {
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

type validOIRParams struct {
	ID                   dssmodels.ID
	OVN                  scdmodels.OVN
	NewOVN               scdmodels.OVN
	State                scdmodels.OperationalIntentState
	UExtent              *dssmodels.Volume4D
	Cells                s2.CellUnion
	SubscriptionID       dssmodels.ID
	USSBaseURL           string
	ImplicitSubscription struct {
		Requested      bool
		BaseURL        string
		ForConstraints bool
	}
	Key map[scdmodels.OVN]bool
}

func (vp *validOIRParams) toOIR(manager dssmodels.Manager, attachedSub *scdmodels.Subscription, version scdmodels.VersionNumber, pastOVNs []scdmodels.OVN) *scdmodels.OperationalIntent {
	// For OIR's in the accepted state, we may not have a attachedSub available,
	// in such cases the attachedSub ID on scdmodels.OperationalIntent will be nil
	// and will be replaced with the 'NullV4UUID' when sent over to a client.
	var subID *dssmodels.ID
	if attachedSub != nil {
		// Note: do _not_ use vp.SubscriptionID here, as it may be empty
		subID = &attachedSub.ID
	}
	return &scdmodels.OperationalIntent{
		ID:       vp.ID,
		Manager:  manager,
		Version:  version,
		OVN:      vp.NewOVN, // non-empty only if the USS has requested an OVN
		PastOVNs: pastOVNs,

		StartTime:     vp.UExtent.StartTime,
		EndTime:       vp.UExtent.EndTime,
		AltitudeLower: vp.UExtent.SpatialVolume.AltitudeLo,
		AltitudeUpper: vp.UExtent.SpatialVolume.AltitudeHi,
		Cells:         vp.Cells,

		USSBaseURL:     vp.USSBaseURL,
		SubscriptionID: subID,
		State:          vp.State,
	}
}

// ValidateAndReturnOIRUpsertParams checks that the parameters for an Operational Intent Reference upsert are valid.
// Note that this does NOT check for anything related to access controls: any error returned should be labeled
// as a dsserr.BadRequest.
func ValidateAndReturnOIRUpsertParams(
	now time.Time,
	entityid restapi.EntityID,
	ovn restapi.EntityOVN,
	params *restapi.PutOperationalIntentReferenceParameters,
	allowHTTPBaseUrls bool,
) (*validOIRParams, error) {

	valid := &validOIRParams{}
	var err error

	valid.ID, err = dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, stacktrace.NewError("Invalid ID format: `%s`", entityid)
	}

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewError("Missing required UssBaseUrl")
	}

	valid.USSBaseURL = string(params.UssBaseUrl)

	if params.SubscriptionId != nil {
		valid.SubscriptionID, err = dssmodels.IDFromOptionalString(string(*params.SubscriptionId))
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
		valid.ImplicitSubscription.Requested = true
		valid.ImplicitSubscription.BaseURL = string(params.NewSubscription.UssBaseUrl)
		// notify for constraints defaults to false if not specified
		if params.NewSubscription.NotifyForConstraints != nil {
			valid.ImplicitSubscription.ForConstraints = *params.NewSubscription.NotifyForConstraints
		}
	}

	if !allowHTTPBaseUrls {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to validate base URL")
		}

		if params.NewSubscription != nil {
			err := scdmodels.ValidateUSSBaseURL(valid.ImplicitSubscription.BaseURL)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to validate USS base URL for subscription (in parameters for implicit subscription)")
			}
		}
	}

	valid.State = scdmodels.OperationalIntentState(params.State)
	if !valid.State.IsValidInDSS() {
		return nil, stacktrace.NewError("Invalid OperationalIntent state: %s", params.State)
	}

	// Start and end times, as well as lower and upper altitudes, are required for each volume
	// The end time may not be in the past.
	valid.UExtent, err = scdmodels.UnionVolumes4DFromSCDRest(
		params.Extents,
		scdmodels.WithRequireTimeBounds(),
		scdmodels.WithRequireAltitudeBounds(),
		scdmodels.WithRequireEndTimeAfter(now),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid extents")
	}
	valid.Cells, err = valid.UExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	if ovn == "" && params.State != restapi.OperationalIntentState_Accepted {
		return nil, stacktrace.NewError("Invalid state for initial version: `%s`", params.State)
	}
	valid.OVN = scdmodels.OVN(ovn)

	if params.RequestedOvnSuffix != nil {
		valid.NewOVN, err = scdmodels.NewOVNFromUUIDv7Suffix(now, valid.ID, string(*params.RequestedOvnSuffix))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Invalid requested OVN suffix")
		}
	}

	// Check if a subscription is required for this request:
	// OIRs in an accepted state do not need a subscription.
	if valid.State.RequiresSubscription() &&
		valid.SubscriptionID.Empty() &&
		(params.NewSubscription == nil ||
			params.NewSubscription.UssBaseUrl == "") {
		return nil, stacktrace.NewError("Provided Operational Intent Reference state `%s` requires either a subscription ID or information to create an implicit subscription", valid.State)
	}

	// Construct a hash set of OVNs as the key
	valid.Key = map[scdmodels.OVN]bool{}
	if params.Key != nil {
		for _, ovn := range *params.Key {
			valid.Key[scdmodels.OVN(ovn)] = true
		}
	}

	return valid, nil
}

// createAndStoreNewImplicitSubscription will create a brand new implicit subscription based on the provided parameters,
// store it and return it.
func createAndStoreNewImplicitSubscription(ctx context.Context, r repos.Repository, manager dssmodels.Manager, validParams *validOIRParams) (*scdmodels.Subscription, error) {
	id, err := scdmodels.NewDeterministicImplicitSubscriptionID(timestamp.MustGetRequestTimestamp(ctx), validParams.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create implicit subscription ID")
	}

	subToUpsert := scdmodels.Subscription{
		ID:                          id,
		Manager:                     manager,
		StartTime:                   validParams.UExtent.StartTime,
		EndTime:                     validParams.UExtent.EndTime,
		AltitudeLo:                  validParams.UExtent.SpatialVolume.AltitudeLo,
		AltitudeHi:                  validParams.UExtent.SpatialVolume.AltitudeHi,
		Cells:                       validParams.Cells,
		USSBaseURL:                  validParams.ImplicitSubscription.BaseURL,
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        validParams.ImplicitSubscription.ForConstraints,
		ImplicitSubscription:        true,
	}

	return r.UpsertSubscription(ctx, &subToUpsert)
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
	relevantOps, err := r.SearchOperationalIntents(ctx, params.UExtent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to SearchOperations")
	}
	for _, relevantOp := range relevantOps {
		_, ok := params.Key[relevantOp.OVN]
		// Note: The OIR being mutated does not need to be specified in the key:
		if !ok && relevantOp.RequiresKey() && relevantOp.ID != params.ID {
			missingOps = append(missingOps, relevantOp)
		}
	}

	// Identify Constraints missing from the key
	var missingConstraints []*scdmodels.Constraint
	if attachedSubscription != nil && attachedSubscription.NotifyForConstraints {
		constraints, err := r.SearchConstraints(ctx, params.UExtent)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to SearchConstraints")
		}
		for _, relevantConstraint := range constraints {
			if _, ok := params.Key[relevantConstraint.OVN]; !ok {
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

// ensureSubscriptionCoversOIR ensures that the subscription covers the requested geo-temporal extent, extending it if both possible and required,
// or failing otherwise.
// After this method returns successfully, the subscription will cover the requested geo-temporal extent.
func ensureSubscriptionCoversOIR(ctx context.Context, r repos.Repository, sub *scdmodels.Subscription, params *validOIRParams) (*scdmodels.Subscription, error) {

	updateSub := false
	if sub.StartTime != nil && sub.StartTime.After(*params.UExtent.StartTime) {
		if sub.ImplicitSubscription {
			sub.StartTime = params.UExtent.StartTime
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
		}
	}
	if sub.EndTime != nil && sub.EndTime.Before(*params.UExtent.EndTime) {
		if sub.ImplicitSubscription {
			sub.EndTime = params.UExtent.EndTime
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
		}
	}
	if !sub.Cells.Contains(params.Cells) {
		if sub.ImplicitSubscription {
			sub.Cells = s2.CellUnionFromUnion(sub.Cells, params.Cells)
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the OperationalIntent")
		}
	}
	if updateSub {
		upsertedSub, err := r.UpsertSubscription(ctx, sub)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to update existing Subscription")
		}
		return upsertedSub, nil
	}

	return sub, nil
}

// executePutOperationalIntentReference inserts or updates an Operational Intent.
// If the ovn argument is empty (""), it will attempt to create a new Operational Intent.
func executePutOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		entityid restapi.EntityID
		ovn      restapi.EntityOVN
		params   *restapi.PutOperationalIntentReferenceParameters
		auth     *api.AuthorizationResult
	)

	switch req := request.(type) {
	case *restapi.CreateOperationalIntentReferenceRequest:
		entityid, params, auth = req.Entityid, req.Body, &req.Auth
	case *restapi.UpdateOperationalIntentReferenceRequest:
		entityid, ovn, params, auth = req.Entityid, req.Ovn, req.Body, &req.Auth
	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.CreateOperationalIntentReferenceOperationID)
	}

	now := timestamp.MustGetRequestTimestamp(ctx)

	// Base URL scheme validation is a pre-flight, request-only check performed by the handler
	// before this action is proposed for consensus; skip it here (allowHTTPBaseUrls: true).
	validParams, err := ValidateAndReturnOIRUpsertParams(now, entityid, ovn, params, true)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Operational Intent Reference upsert parameters")
	}
	manager, err := CheckUpsertPermissionsAndReturnManager(auth, validParams.State)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.PermissionDenied, "Caller is not allowed to upsert with the requested state")
	}

	// Get existing OperationalIntent, if any
	old, err := repo.GetOperationalIntent(ctx, validParams.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
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

	err = repo.LockSubscriptionsOnCells(ctx, validParams.Cells, subscriptionIds, validParams.UExtent.StartTime, validParams.UExtent.EndTime)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to acquire lock")
	}

	// Validate the request against the previous OIR
	if err := validateUpsertRequestAgainstPreviousOIR(manager, validParams.OVN, old); err != nil {
		return nil, stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Request validation failed")
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
			previousSub, err = repo.GetSubscription(ctx, *old.SubscriptionID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
			}
		}
	}

	// Determine if the previous subscription is being replaced and if it will need to be cleaned up
	previousSubIsBeingReplaced := previousSub != nil && validParams.SubscriptionID != previousSub.ID
	removePreviousImplicitSubscription := false
	if previousSubIsBeingReplaced {
		removePreviousImplicitSubscription, err = SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx, repo, validParams.ID, previousSub)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not determine if previous Subscription can be removed")
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
			if attachedSub, err = createAndStoreNewImplicitSubscription(ctx, repo, manager, validParams); err != nil {
				return nil, stacktrace.Propagate(err, "Failed to create implicit subscription")
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
			attachedSub, err = repo.GetSubscription(ctx, validParams.SubscriptionID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Unable to get requested Subscription from store")
			}
			if attachedSub == nil {
				return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", validParams.SubscriptionID)
			}
		}

		// We need to confirm that it is owned by the calling manager
		if attachedSub.Manager != manager {
			return nil, stacktrace.Propagate(
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
		attachedSub, err = ensureSubscriptionCoversOIR(ctx, repo, attachedSub, validParams)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to ensure subscription covers OIR")
		}
	}

	var responseConflict *restapi.AirspaceConflictResponse
	if validParams.State.RequiresKey() {
		responseConflict, err = validateKeyAndProvideConflictResponse(ctx, repo, manager, validParams, attachedSub)
		if err != nil {
			// responseConflict is non-nil here on a dsserr.MissingOVNs error: return it alongside
			// the error so the handler can still send it to the client. See the doc comment above.
			return responseConflict, stacktrace.PropagateWithCode(err, stacktrace.GetCode(err), "Failed to validate key")
		}
	}

	// Construct the new OperationalIntent
	op := validParams.toOIR(manager, attachedSub, version, pastOVNs)

	// Upsert the OperationalIntent
	op, err = repo.UpsertOperationalIntent(ctx, op)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
	}

	// Check if the previously attached subscription should be removed
	if removePreviousImplicitSubscription {
		err = repo.DeleteSubscription(ctx, previousSub.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to delete previous implicit Subscription")
		}
	}

	notifyVolume, err := computeNotificationVolume(old, validParams.UExtent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to compute notification volume")
	}

	// Notify relevant Subscriptions
	subsToNotify, err := GetRelevantSubscriptionsAndIncrementIndices(ctx, repo, notifyVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to notify relevant Subscriptions")
	}

	// Return response to client
	return &restapi.ChangeOperationalIntentReferenceResponse{
		OperationalIntentReference: *op.ToRest(),
		Subscribers:                makeSubscribersToNotify(subsToNotify),
	}, nil
}
