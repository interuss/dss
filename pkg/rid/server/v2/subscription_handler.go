package server

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv2 "github.com/interuss/dss/pkg/rid/models/api/v2"
	"github.com/interuss/stacktrace"
	"github.com/pkg/errors"
)

// DeleteSubscription deletes an existing subscription.
func (s *Server) DeleteSubscription(ctx context.Context, req *restapi.DeleteSubscriptionRequest,
) restapi.DeleteSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.Auth.ClientID == nil {
		return restapi.DeleteSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	version, err := dssmodels.VersionFromString(req.Version)
	if err != nil {
		return restapi.DeleteSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version"))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.DeleteSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}
	// TODO: put the context with timeout into an interceptor so it's always set.
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.DeleteSubscription(ctx, id, dssmodels.Owner(*req.Auth.ClientID), version)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete Subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteSubscriptionResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.DeleteSubscriptionResponseSet{Response409: errResp}
		case dsserr.NotFound:
			return restapi.DeleteSubscriptionResponseSet{Response404: errResp}
		default:
			return restapi.DeleteSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.DeleteSubscriptionResponseSet{Response200: &restapi.DeleteSubscriptionResponse{
		Subscription: *apiv2.ToSubscription(subscription),
	}}
}

// SearchSubscriptions queries for existing subscriptions in the given bounds.
func (s *Server) SearchSubscriptions(ctx context.Context, req *restapi.SearchSubscriptionsRequest,
) restapi.SearchSubscriptionsResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.SearchSubscriptionsResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.Auth.ClientID == nil {
		return restapi.SearchSubscriptionsResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	if req.Area == nil {
		return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area"))}}
	}
	cu, err := geo.AreaToCellIDs(string(*req.Area))
	if err != nil {
		if errors.Is(err, geo.ErrAreaTooLarge) {
			return restapi.SearchSubscriptionsResponseSet{Response413: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Invalid area"))}}
		}
		return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area"))}}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscriptions, err := s.App.SearchSubscriptionsByOwner(ctx, cu, dssmodels.Owner(*req.Auth.ClientID))
	if err != nil {
		err = stacktrace.Propagate(err, "Could not search Subscriptions")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.SearchSubscriptionsResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	sp := make([]restapi.Subscription, 0, len(subscriptions))
	for _, sub := range subscriptions {
		sp = append(sp, *apiv2.ToSubscription(sub))
	}

	return restapi.SearchSubscriptionsResponseSet{Response200: &restapi.SearchSubscriptionsResponse{
		Subscriptions: &sp,
	}}
}

// GetSubscription gets a single subscription based on ID.
func (s *Server) GetSubscription(ctx context.Context, req *restapi.GetSubscriptionRequest,
) restapi.GetSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.GetSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.GetSubscription(ctx, id)
	if err != nil {
		return restapi.GetSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not get Subscription"))}}
	}
	if subscription == nil {
		return restapi.GetSubscriptionResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", req.Id))}}
	}
	return restapi.GetSubscriptionResponseSet{Response200: &restapi.GetSubscriptionResponse{
		Subscription: *apiv2.ToSubscription(subscription)}}
}

// CreateSubscription creates a single subscription.
func (s *Server) CreateSubscription(ctx context.Context, req *restapi.CreateSubscriptionRequest,
) restapi.CreateSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.CreateSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	if req.Auth.ClientID == nil {
		return restapi.CreateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if req.BodyParseError != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.UssBaseUrl == "" {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL"))}}
	}
	extents, err := apiv2.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	if !s.AllowHTTPBaseUrls {
		err = ridmodels.ValidateURL(string(req.Body.UssBaseUrl))
		if err != nil {
			return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate UssBaseUrl"))}}
		}
	}

	sub := &ridmodels.Subscription{
		ID:     id,
		Owner:  dssmodels.Owner(*req.Auth.ClientID),
		URL:    string(req.Body.UssBaseUrl),
		Writer: s.Locality,
	}

	if err := sub.SetExtents(extents); err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents"))}}
	}

	insertedSub, err := s.App.InsertSubscription(ctx, sub)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not insert Subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.AlreadyExists:
			return restapi.CreateSubscriptionResponseSet{Response409: errResp}
		case dsserr.BadRequest:
			return restapi.CreateSubscriptionResponseSet{Response400: errResp}
		case dsserr.Exhausted:
			return restapi.CreateSubscriptionResponseSet{Response429: errResp}
		default:
			return restapi.CreateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	// Find ISAs that were in this subscription's area.
	isas, err := s.App.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not search ISAs")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.CreateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	// Convert the ISAs to REST.
	restIsas := make([]restapi.IdentificationServiceArea, 0, len(isas))
	for _, isa := range isas {
		restIsas = append(restIsas, *apiv2.ToIdentificationServiceArea(isa))
	}

	return restapi.CreateSubscriptionResponseSet{Response200: &restapi.PutSubscriptionResponse{
		Subscription: *apiv2.ToSubscription(insertedSub),
		ServiceAreas: &restIsas,
	}}
}

// UpdateSubscription updates a single subscription.
func (s *Server) UpdateSubscription(ctx context.Context, req *restapi.UpdateSubscriptionRequest,
) restapi.UpdateSubscriptionResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.UpdateSubscriptionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	version, err := dssmodels.VersionFromString(req.Version)
	if err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version"))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	if req.Auth.ClientID == nil {
		return restapi.UpdateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if req.BodyParseError != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.UssBaseUrl == "" {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL"))}}
	}
	extents, err := apiv2.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}

	sub := &ridmodels.Subscription{
		ID:      id,
		Owner:   dssmodels.Owner(*req.Auth.ClientID),
		URL:     string(req.Body.UssBaseUrl),
		Version: version,
		Writer:  s.Locality,
	}

	if err := sub.SetExtents(extents); err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents"))}}
	}

	insertedSub, err := s.App.UpdateSubscription(ctx, sub)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not update Subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateSubscriptionResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.UpdateSubscriptionResponseSet{Response409: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.UpdateSubscriptionResponseSet{Response400: errResp}
		case dsserr.Exhausted:
			return restapi.UpdateSubscriptionResponseSet{Response429: errResp}
		default:
			return restapi.UpdateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	// Find ISAs that were in this subscription's area.
	isas, err := s.App.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not search ISAs")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.UpdateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	// Convert the ISAs to REST.
	restIsas := make([]restapi.IdentificationServiceArea, 0, len(isas))
	for _, isa := range isas {
		restIsas = append(restIsas, *apiv2.ToIdentificationServiceArea(isa))
	}

	return restapi.UpdateSubscriptionResponseSet{Response200: &restapi.PutSubscriptionResponse{
		Subscription: *apiv2.ToSubscription(insertedSub),
		ServiceAreas: &restIsas,
	}}
}
