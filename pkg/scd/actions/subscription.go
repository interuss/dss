package actions

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.DeleteSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteSubscriptionRequest],
		Execute: ExecuteDeleteSubscription,
	}
	Registry[restapi.GetSubscriptionOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetSubscriptionRequest],
		Execute:    ExecuteGetSubscription,
		IsReadOnly: true,
	}
	Registry[restapi.QuerySubscriptionsOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.QuerySubscriptionsRequest],
		Execute:    ExecuteQuerySubscriptions,
		IsReadOnly: true,
	}
}

func ExecuteDeleteSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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

func ExecuteGetSubscription(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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

func ExecuteQuerySubscriptions(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
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

	nowMarker := timestamp.MustGetRequestTimestamp(ctx)

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
