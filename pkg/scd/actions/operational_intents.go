package actions

import (
	"context"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.GetOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetOperationalIntentReferenceRequest],
		Execute:    ExecuteGetOperationalIntentReference,
		IsReadOnly: true,
	}
	Registry[restapi.QueryOperationalIntentReferencesOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QueryOperationalIntentReferencesRequest],
		Execute:    ExecuteQueryOperationalIntentReferences,
		IsReadOnly: true,
	}
	Registry[restapi.DeleteOperationalIntentReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteOperationalIntentReferenceRequest],
		Execute: ExecuteDeleteOperationalIntentReference,
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

// ExecuteDeleteOperationalIntentReference deletes a single operational intent ref for a given ID
// at the specified version.
func ExecuteDeleteOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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

	subsToNotify, err := repo.IncrementNotificationIndicesForOperationalIntents(ctx, notifyVolume)
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

func ExecuteGetOperationalIntentReference(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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

func ExecuteQueryOperationalIntentReferences(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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
