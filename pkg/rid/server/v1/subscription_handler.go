package v1

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/rid_v1"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	geoerr "github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv1 "github.com/interuss/dss/pkg/rid/models/api/v1"
	"github.com/interuss/stacktrace"
	"github.com/pkg/errors"
)

// DeleteSubscription deletes an existing subscription.
func (s *Server) DeleteSubscription(ctx context.Context, req *restapi.DeleteSubscriptionRequest,
) restapi.DeleteSubscriptionResponseSet {

	// TODO: simply verify the owner was set in an upper level.
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return restapi.DeleteSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context"))}}
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
	subscription, err := s.App.DeleteSubscription(ctx, id, owner, version)
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
		Subscription: *apiv1.ToSubscription(subscription),
	}}
}

// SearchSubscriptions queries for existing subscriptions in the given bounds.
func (s *Server) SearchSubscriptions(ctx context.Context, req *restapi.SearchSubscriptionsRequest,
) restapi.SearchSubscriptionsResponseSet {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return restapi.SearchSubscriptionsResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context"))}}
	}

	if req.Area == nil {
		return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area"))}}
	}
	cu, err := geo.AreaToCellIDs(string(*req.Area))
	if err != nil {
		if errors.Is(err, geoerr.ErrAreaTooLarge) {
			return restapi.SearchSubscriptionsResponseSet{Response413: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Invalid area"))}}
		}
		return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area"))}}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscriptions, err := s.App.SearchSubscriptionsByOwner(ctx, cu, owner)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not search Subscriptions")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.SearchSubscriptionsResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		} else {
			return restapi.SearchSubscriptionsResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	sp := make([]restapi.Subscription, 0, len(subscriptions))
	for _, sub := range subscriptions {
		sp = append(sp, *apiv1.ToSubscription(sub))
	}

	return restapi.SearchSubscriptionsResponseSet{Response200: &restapi.SearchSubscriptionsResponse{
		Subscriptions: sp,
	}}
}

// GetSubscription gets a single subscription based on ID.
func (s *Server) GetSubscription(ctx context.Context, req *restapi.GetSubscriptionRequest,
) restapi.GetSubscriptionResponseSet {

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
		Subscription: *apiv1.ToSubscription(subscription)}}
}

// CreateSubscription creates a single subscription.
func (s *Server) CreateSubscription(ctx context.Context, req *restapi.CreateSubscriptionRequest,
) restapi.CreateSubscriptionResponseSet {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return restapi.CreateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context"))}}
	}
	if req.BodyParseError != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.Callbacks.IdentificationServiceAreaUrl == nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required callbacks"))}}
	}
	if len(req.Body.Extents.SpatialVolume.Footprint.Vertices) == 0 {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents"))}}
	}
	extents, err := apiv1.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	if !s.EnableHTTP {
		err = ridmodels.ValidateURL(string(*req.Body.Callbacks.IdentificationServiceAreaUrl))
		if err != nil {
			return restapi.CreateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate IdentificationServiceAreaUrl"))}}
		}
	}

	sub := &ridmodels.Subscription{
		ID:     id,
		Owner:  owner,
		URL:    string(*req.Body.Callbacks.IdentificationServiceAreaUrl),
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
		} else {
			return restapi.CreateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	// Convert the ISAs to REST.
	restIsas := make([]restapi.IdentificationServiceArea, 0, len(isas))
	for _, isa := range isas {
		restIsas = append(restIsas, *apiv1.ToIdentificationServiceArea(isa))
	}

	return restapi.CreateSubscriptionResponseSet{Response200: &restapi.PutSubscriptionResponse{
		Subscription: *apiv1.ToSubscription(insertedSub),
		ServiceAreas: &restIsas,
	}}
}

// UpdateSubscription updates a single subscription.
func (s *Server) UpdateSubscription(ctx context.Context, req *restapi.UpdateSubscriptionRequest,
) restapi.UpdateSubscriptionResponseSet {

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

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return restapi.UpdateSubscriptionResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context"))}}
	}
	if req.BodyParseError != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.Callbacks.IdentificationServiceAreaUrl == nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required callbacks"))}}
	}
	if len(req.Body.Extents.SpatialVolume.Footprint.Vertices) == 0 {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents"))}}
	}
	extents, err := apiv1.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.UpdateSubscriptionResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}

	sub := &ridmodels.Subscription{
		ID:      id,
		Owner:   owner,
		URL:     string(*req.Body.Callbacks.IdentificationServiceAreaUrl),
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
		} else {
			return restapi.UpdateSubscriptionResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	// Convert the ISAs to REST.
	restIsas := make([]restapi.IdentificationServiceArea, 0, len(isas))
	for _, isa := range isas {
		restIsas = append(restIsas, *apiv1.ToIdentificationServiceArea(isa))
	}

	return restapi.UpdateSubscriptionResponseSet{Response200: &restapi.PutSubscriptionResponse{
		Subscription: *apiv1.ToSubscription(insertedSub),
		ServiceAreas: &restIsas,
	}}
}
