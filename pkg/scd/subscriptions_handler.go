package scd

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"github.com/jonboulle/clockwork"
)

var (
	DefaultClock = clockwork.NewRealClock()
)

func (a *Server) CreateSubscription(ctx context.Context, req *restapi.CreateSubscriptionRequest,
) restapi.CreateSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.CreateSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.CreateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	res, err := a.PutSubscription(ctx, *req.Auth.ClientID, req.Subscriptionid, "", req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.CreateSubscriptionResponseSet{Response403: errResp}
		case dsserr.AlreadyExists, dsserr.VersionMismatch:
			return restapi.CreateSubscriptionResponseSet{Response409: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.CreateSubscriptionResponseSet{Response400: errResp}
		default:
			return restapi.CreateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.CreateSubscriptionResponseSet{Response200: res}
}

func (a *Server) UpdateSubscription(ctx context.Context, req *restapi.UpdateSubscriptionRequest,
) restapi.UpdateSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.UpdateSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.UpdateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	res, err := a.PutSubscription(ctx, *req.Auth.ClientID, req.Subscriptionid, req.Version, req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateSubscriptionResponseSet{Response403: errResp}
		case dsserr.AlreadyExists, dsserr.VersionMismatch:
			return restapi.UpdateSubscriptionResponseSet{Response409: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.UpdateSubscriptionResponseSet{Response400: errResp}
		default:
			return restapi.UpdateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.UpdateSubscriptionResponseSet{Response200: res}
}

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, manager string, subscriptionid restapi.SubscriptionID, version string, params *restapi.PutSubscriptionParameters,
) (*restapi.PutSubscriptionResponse, error) {
	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(string(subscriptionid))

	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", subscriptionid)
	}

	if !a.EnableHTTP {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	// Parse extents
	extents, err := dssmodels.Volume4DFromSCDRest(&params.Extents)
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

	var result *restapi.PutSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Check existing Subscription (if any)
		old, err := r.GetSubscription(ctx, subreq.ID)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get Subscription from repo")
		}

		// Validate and perhaps correct StartTime and EndTime.
		if err := subreq.AdjustTimeRange(DefaultClock.Now(), old); err != nil {
			return stacktrace.Propagate(err, "Error adjusting time range of Subscription")
		}

		var dependentOpIds []dssmodels.ID

		if old == nil {
			// There is no previous Subscription (this is a creation attempt)
			if subreq.Version.String() != "" {
				// The user wants to update an existing Subscription, but one wasn't found.
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", subreq.ID.String())
			}
		} else {
			// There is a previous Subscription (this is an update attempt)
			switch {
			case subreq.Version.String() == "":
				// The user wants to create a new Subscription but it already exists.
				return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", subreq.ID.String())
			case subreq.Version.String() != old.Version.String():
				// The user wants to update a Subscription but the version doesn't match.
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", subreq.Version),
					"Current version is %s but client specified version %s", old.Version, subreq.Version)
			case old.Manager != subreq.Manager:
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
					"Subscription owned by %s, but %s attempted to modify", old.Manager, subreq.Manager)
			}

			subreq.NotificationIndex = old.NotificationIndex

			// Validate Subscription against DependentOperations
			dependentOpIds, err = r.GetDependentOperationalIntents(ctx, subreq.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Could not find dependent Operation Ids")
			}

			operations, err := GetOperations(ctx, r, dependentOpIds)
			if err != nil {
				return stacktrace.Propagate(err, "Could not get all dependent Operations")
			}
			if err := subreq.ValidateDependentOps(operations); err != nil {
				// The provided subscription does not cover all its dependent operations
				return err
			}
		}

		// Store Subscription model
		sub, err := r.UpsertSubscription(ctx, subreq)
		if err != nil {
			return stacktrace.Propagate(err, "Could not upsert Subscription into repo")
		}
		if sub == nil {
			return stacktrace.NewError("UpsertSubscription returned no Subscription for ID: %s", id)
		}

		// Find relevant Operations
		var relevantOperations []*scdmodels.OperationalIntent
		if len(sub.Cells) > 0 {
			ops, err := r.SearchOperationalIntents(ctx, &dssmodels.Volume4D{
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
				return stacktrace.Propagate(err, "Could not search Operations in repo")
			}
			relevantOperations = ops
		}

		// Convert Subscription to REST
		p, err := sub.ToRest(dependentOpIds)
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Subscription to REST model")
		}
		result = &restapi.PutSubscriptionResponse{
			Subscription: *p,
		}

		if sub.NotifyForOperationalIntents {
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
			constraints, err := r.SearchConstraints(ctx, extents)
			if err != nil {
				return stacktrace.Propagate(err, "Could not search Constraints in repo")
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

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	// Return response to client
	return result, nil
}

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *restapi.GetSubscriptionRequest,
) restapi.GetSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return restapi.GetSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Subscriptionid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.GetSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	var response *restapi.GetSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get Subscription from Store
		sub, err := r.GetSubscription(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get Subscription from repo")
		}
		if sub == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
		}

		// Check if the client is authorized to view this Subscription
		if dssmodels.Manager(*req.Auth.ClientID) != sub.Manager {
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to view", sub.Manager, *req.Auth.ClientID)
		}

		// Get dependent Operations
		dependentOps, err := r.GetDependentOperationalIntents(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not find dependent Operations")
		}

		// Convert Subscription to REST
		p, err := sub.ToRest(dependentOps)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to convert Subscription to REST")
		}

		// Return response to client
		response = &restapi.GetSubscriptionResponse{
			Subscription: *p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not get subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.GetSubscriptionResponseSet{Response403: errResp}
		case dsserr.BadRequest:
			return restapi.GetSubscriptionResponseSet{Response400: errResp}
		case dsserr.NotFound:
			return restapi.GetSubscriptionResponseSet{Response404: errResp}
		default:
			return restapi.GetSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.GetSubscriptionResponseSet{Response200: response}
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *restapi.QuerySubscriptionsRequest,
) restapi.QuerySubscriptionsResponseSet {
	nowMarker := time.Now()
	if req.Auth.Error != nil {
		resp := restapi.QuerySubscriptionsResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.QuerySubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return restapi.QuerySubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest"))}}
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QuerySubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QuerySubscriptionsResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	var response *restapi.QuerySubscriptionsResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		subs, err := r.SearchSubscriptions(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Error searching Subscriptions in repo")
		}

		// Return response to client
		response = &restapi.QuerySubscriptionsResponse{
			Subscriptions: make([]restapi.Subscription, 0),
		}
		for _, sub := range subs {
			// Do not return subscriptions which are expired.
			// TODO: Add reference to TSC investigation outcome.
			isExpired := sub.EndTime.Before(nowMarker)
			if !isExpired && sub.Manager == dssmodels.Manager(*req.Auth.ClientID) {
				// Get dependent Operations
				dependentOps, err := r.GetDependentOperationalIntents(ctx, sub.ID)
				if err != nil {
					return stacktrace.Propagate(err, "Could not find dependent Operations")
				}

				p, err := sub.ToRest(dependentOps)
				if err != nil {
					return stacktrace.Propagate(err, "Error converting Subscription model to REST")
				}
				response.Subscriptions = append(response.Subscriptions, *p)
			}
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {

		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.QuerySubscriptionsResponseSet{Response400: errResp}
		case dsserr.PermissionDenied:
			return restapi.QuerySubscriptionsResponseSet{Response403: errResp}
		case dsserr.AreaTooLarge:
			return restapi.QuerySubscriptionsResponseSet{Response413: errResp}
		default:
			return restapi.QuerySubscriptionsResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}

		}
	}

	return restapi.QuerySubscriptionsResponseSet{Response200: response}
}

// DeleteSubscription deletes a single subscription for a given ID.
func (a *Server) DeleteSubscription(ctx context.Context, req *restapi.DeleteSubscriptionRequest,
) restapi.DeleteSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return restapi.DeleteSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	// Retrieve Subscription Version
	version := scdmodels.OVN(req.Version)
	if version == "" {
		return restapi.DeleteSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing version"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	var response *restapi.DeleteSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Check to make sure it's ok to delete this Subscription
		old, err := r.GetSubscription(ctx, id)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Could not get Subscription from repo")
		case old == nil: // Return a 404 here.
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
		case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
		case old.Version != version:
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", version)
		}

		// Get dependent Operations
		dependentOps, err := r.GetDependentOperationalIntents(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not find dependent Operations")
		}
		if len(dependentOps) > 0 {
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscriptions with dependent Operations may not be removed"),
				"Subscription had %d dependent Operations", len(dependentOps))
		}

		// Delete Subscription in repo
		err = r.DeleteSubscription(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not delete Subscription from repo")
		}

		// Convert deleted Subscription to REST
		p, err := old.ToRest(dependentOps)
		if err != nil {
			return stacktrace.Propagate(err, "Error converting Subscription model to REST")
		}

		// Create response for client
		response = &restapi.DeleteSubscriptionResponse{
			Subscription: *p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteSubscriptionResponseSet{Response403: errResp}
		case dsserr.BadRequest:
			return restapi.DeleteSubscriptionResponseSet{Response400: errResp}
		case dsserr.NotFound:
			return restapi.DeleteSubscriptionResponseSet{Response404: errResp}
		case dsserr.VersionMismatch:
			return restapi.DeleteSubscriptionResponseSet{Response409: errResp}
		default:
			return restapi.DeleteSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.DeleteSubscriptionResponseSet{Response200: response}
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
