package scd

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func (a *Server) CreateSubscription(ctx context.Context, req *restapi.CreateSubscriptionRequest,
) restapi.CreateSubscriptionResponseSet {

	if req.BodyParseError != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.CreateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if err := a.validatePutSubscriptionParams(req.Subscriptionid, req.Body); err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
	}

	res, err := dssstore.TransactWithResult[repos.Repository, *restapi.PutSubscriptionResponse](ctx, a.Store, req)
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

	if req.BodyParseError != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.UpdateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if err := a.validatePutSubscriptionParams(req.Subscriptionid, req.Body); err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
	}

	res, err := dssstore.TransactWithResult[repos.Repository, *restapi.PutSubscriptionResponse](ctx, a.Store, req)
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

// validatePutSubscriptionParams performs the request validation that can be done ahead of the transaction
func (a *Server) validatePutSubscriptionParams(subscriptionid restapi.SubscriptionID, params *restapi.PutSubscriptionParameters) error {
	// Retrieve Subscription ID
	if _, err := dssmodels.IDFromString(string(subscriptionid)); err != nil {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", subscriptionid)
	}

	if !a.AllowHTTPBaseUrls {
		err := scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	// Parse extents
	// If end time is not specified, the value will be chosen automatically by the DSS.
	// If start time is not specified, it will default to the time the request is processed.
	extents, err := scdmodels.Volume4DFromSCDRest(&params.Extents)
	if err != nil {
		return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to parse extents")
	}

	// Construct requested Subscription model
	_, err = extents.CalculateSpatialCovering()
	switch err {
	case nil, geo.ErrMissingSpatialVolume, geo.ErrMissingFootprint:
		// We may be able to fill these values from a previous Subscription or via defaults.
	default:
		return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	notifyForOperationalIntents := params.NotifyForOperationalIntents != nil && *params.NotifyForOperationalIntents
	notifyForConstraints := params.NotifyForConstraints != nil && *params.NotifyForConstraints

	// Validate requested Subscription
	if !notifyForOperationalIntents && !notifyForConstraints {
		return stacktrace.NewErrorWithCode(dsserr.BadRequest, "No notification triggers requested for Subscription")
	}

	// TODO: Check scopes to verify requested information (op intents or constraints) may be requested

	return nil
}

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *restapi.GetSubscriptionRequest,
) restapi.GetSubscriptionResponseSet {

	// Retrieve Subscription ID
	_, err := dssmodels.IDFromString(string(req.Subscriptionid))
	if err != nil {
		return restapi.GetSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Subscriptionid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.GetSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.GetSubscriptionResponse](ctx, a.Store, req)
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
	_, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QuerySubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QuerySubscriptionsResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.QuerySubscriptionsResponse](ctx, a.Store, req)
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

	// Retrieve Subscription ID
	_, err := dssmodels.IDFromString(string(req.Subscriptionid))
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

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.DeleteSubscriptionResponse](ctx, a.Store, req)
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
