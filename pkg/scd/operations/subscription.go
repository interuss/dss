package operations

import (
	"context"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.CreateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.CreateSubscriptionRequest],
		Execute: executePutSubscription,
	}
	Registry[restapi.UpdateSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.UpdateSubscriptionRequest],
		Execute: executePutSubscription,
	}
	Registry[restapi.DeleteSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteSubscriptionRequest],
		Execute: executeDeleteSubscription,
	}
	Registry[restapi.GetSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetSubscriptionRequest],
		Execute:    executeGetSubscription,
		IsReadOnly: true,
	}
	Registry[restapi.QuerySubscriptionsOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QuerySubscriptionsRequest],
		Execute:    executeQuerySubscriptions,
		IsReadOnly: true,
	}
}

func executePutSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		manager        string
		subscriptionid restapi.SubscriptionID
		version        string
		params         *restapi.PutSubscriptionParameters
	)

	switch req := request.(type) {
	case *restapi.CreateSubscriptionRequest:
		manager, subscriptionid, params = *req.Auth.ClientID, req.Subscriptionid, req.Body
	case *restapi.UpdateSubscriptionRequest:
		manager, subscriptionid, version, params = *req.Auth.ClientID, req.Subscriptionid, req.Version, req.Body
	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.CreateSubscriptionOperationID)
	}

	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(string(subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", subscriptionid)
	}

	// Parse extents
	// If end time is not specified, the value will be chosen automatically by the DSS.
	// If start time is not specified, it will default to the time the request is processed.
	extents, err := scdmodels.Volume4DFromSCDRest(&params.Extents)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to parse extents")
	}

	// Construct requested Subscription model
	cells, err := extents.CalculateSpatialCovering()
	switch err {
	case nil, geo.ErrMissingSpatialVolume, geo.ErrMissingFootprint:
		// We may be able to fill these values from a previous Subscription or via defaults.
	default:
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	subreq := &scdmodels.Subscription{
		ID:      id,
		Manager: dssmodels.Manager(manager),
		Version: scdmodels.OVN(version),

		StartTime:  extents.StartTime,
		EndTime:    extents.EndTime,
		AltitudeLo: extents.SpatialVolume.AltitudeLo,
		AltitudeHi: extents.SpatialVolume.AltitudeHi,
		Cells:      cells,

		USSBaseURL: string(params.UssBaseUrl),
	}
	if params.NotifyForOperationalIntents != nil {
		subreq.NotifyForOperationalIntents = *params.NotifyForOperationalIntents
	}
	if params.NotifyForConstraints != nil {
		subreq.NotifyForConstraints = *params.NotifyForConstraints
	}

	// Validate requested Subscription
	if !subreq.NotifyForOperationalIntents && !subreq.NotifyForConstraints {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "No notification triggers requested for Subscription")
	}

	// TODO: Check scopes to verify requested information (op intents or constraints) may be requested

	// Check existing Subscription (if any)
	old, err := repo.GetSubscription(ctx, subreq.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := subreq.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range of Subscription")
	}

	var dependentOpIds []dssmodels.ID

	if old == nil {
		// There is no previous Subscription (this is a creation attempt)
		if subreq.Version.String() != "" {
			// The user wants to update an existing Subscription, but one wasn't found.
			return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", subreq.ID.String())
		}
	} else {
		// There is a previous Subscription (this is an update attempt)
		switch {
		case subreq.Version.String() == "":
			// The user wants to create a new Subscription but it already exists.
			return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", subreq.ID.String())
		case subreq.Version.String() != old.Version.String():
			// The user wants to update a Subscription but the version doesn't match.
			return nil, stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", subreq.Version),
				"Current version is %s but client specified version %s", old.Version, subreq.Version)
		case old.Manager != subreq.Manager:
			return nil, stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to modify", old.Manager, subreq.Manager)
		}

		subreq.NotificationIndex = old.NotificationIndex

		// Validate Subscription against DependentOperations
		dependentOpIds, err = repo.GetDependentOperationalIntents(ctx, subreq.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not find dependent Operation Ids")
		}

		operations, err := GetOperations(ctx, repo, dependentOpIds)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not get all dependent Operations")
		}
		if err := subreq.ValidateDependentOps(operations); err != nil {
			// The provided subscription does not cover all its dependent operations
			return nil, err
		}
	}

	// Store Subscription model
	sub, err := repo.UpsertSubscription(ctx, subreq)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not upsert Subscription into repo")
	}
	if sub == nil {
		return nil, stacktrace.NewError("UpsertSubscription returned no Subscription for ID: %s", id)
	}

	// Convert Subscription to REST
	p, err := sub.ToRest(dependentOpIds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not convert Subscription to REST model")
	}
	result := &restapi.PutSubscriptionResponse{
		Subscription: *p,
	}

	if sub.NotifyForOperationalIntents {
		// Find relevant Operations
		var relevantOperations []*scdmodels.OperationalIntent
		if len(sub.Cells) > 0 {
			ops, err := repo.SearchOperationalIntents(ctx, &dssmodels.Volume4D{
				StartTime: sub.StartTime,
				EndTime:   sub.EndTime,
				SpatialVolume: &dssmodels.Volume3D{
					AltitudeLo: sub.AltitudeLo,
					AltitudeHi: sub.AltitudeHi,
					Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
						return sub.Cells, nil
					}),
				},
			})
			if err != nil {
				return nil, stacktrace.Propagate(err, "Could not search Operations in repo")
			}
			relevantOperations = ops
		}
		// Attach Operations to response
		opIntentRefs := make([]restapi.OperationalIntentReference, 0, len(relevantOperations))
		for _, op := range relevantOperations {
			if op.Manager != dssmodels.Manager(manager) {
				op.OVN = scdmodels.NoOvnPhrase
			}

			opIntentRefs = append(opIntentRefs, *op.ToRest())
		}
		result.OperationalIntentReferences = &opIntentRefs
	}

	if sub.NotifyForConstraints {
		// Query relevant Constraints
		constraints, err := repo.SearchConstraints(ctx, extents)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not search Constraints in repo")
		}

		// Attach Constraints to response
		constraintRefs := make([]restapi.ConstraintReference, 0, len(constraints))
		for _, constraint := range constraints {
			p := constraint.ToRest()
			if constraint.Manager != dssmodels.Manager(manager) {
				noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
				p.Ovn = &noOvnPhrase
			}

			constraintRefs = append(constraintRefs, *p)
		}
		result.ConstraintReferences = &constraintRefs
	}

	return result, nil
}

// GetOperations gets operations by given ids
func GetOperations(ctx context.Context, r repos.Repository, opIDs []dssmodels.ID) ([]*scdmodels.OperationalIntent, error) {
	var res []*scdmodels.OperationalIntent
	for _, opID := range opIDs {
		operation, err := r.GetOperationalIntent(ctx, opID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not retrieve dependent Operation %s", opID)
		}
		res = append(res, operation)
	}
	return res, nil
}

func executeDeleteSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.DeleteSubscriptionRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteSubscriptionOperationID)
	}

	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid subscription ID: %s", req.Subscriptionid)
	}

	// Check to make sure it's ok to delete this Subscription
	old, err := repo.GetSubscription(ctx, id)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	case old == nil: // Return a 404 here.
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.Version != scdmodels.OVN(req.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", scdmodels.OVN(req.Version))
	}

	// Get dependent Operations
	dependentOps, err := repo.GetDependentOperationalIntents(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
	}
	if len(dependentOps) > 0 {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscriptions with dependent Operations may not be removed"),
			"Subscription had %d dependent Operations", len(dependentOps))
	}

	// Delete Subscription in repo
	err = repo.DeleteSubscription(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not delete Subscription from repo")
	}

	// Convert deleted Subscription to REST
	p, err := old.ToRest(dependentOps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting Subscription model to REST")
	}

	return &restapi.DeleteSubscriptionResponse{Subscription: *p}, nil
}

func executeGetSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.GetSubscriptionRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetSubscriptionOperationID)
	}

	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Subscriptionid)
	}

	// Get Subscription from Store
	sub, err := repo.GetSubscription(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get Subscription from repo")
	}
	if sub == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
	}

	// Check if the client is authorized to view this Subscription
	if dssmodels.Manager(*req.Auth.ClientID) != sub.Manager {
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
			"Subscription owned by %s, but %s attempted to view", sub.Manager, *req.Auth.ClientID)
	}

	// Get dependent Operations
	dependentOps, err := repo.GetDependentOperationalIntents(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
	}

	// Convert Subscription to REST
	p, err := sub.ToRest(dependentOps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to convert Subscription to REST")
	}

	// Return response to client
	return &restapi.GetSubscriptionResponse{Subscription: *p}, nil
}

func executeQuerySubscriptions(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.QuerySubscriptionsRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.QuerySubscriptionsOperationID)
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model")
	}

	// Perform search query on Store
	subs, err := repo.SearchSubscriptions(ctx, vol4)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error searching Subscriptions in repo")
	}

	nowMarker := timestamp.MustFromContext(ctx)

	// Return response to client
	response := &restapi.QuerySubscriptionsResponse{
		Subscriptions: make([]restapi.Subscription, 0),
	}
	for _, sub := range subs {
		// Do not return subscriptions which are expired.
		// This implementation decision is described and motivated in https://github.com/interuss/tsc/pull/12.
		isExpired := sub.EndTime.Before(nowMarker)
		if !isExpired && sub.Manager == dssmodels.Manager(*req.Auth.ClientID) {
			// Get dependent Operations
			dependentOps, err := repo.GetDependentOperationalIntents(ctx, sub.ID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Could not find dependent Operations")
			}

			p, err := sub.ToRest(dependentOps)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Error converting Subscription model to REST")
			}
			response.Subscriptions = append(response.Subscriptions, *p)
		}
	}

	return response, nil
}
