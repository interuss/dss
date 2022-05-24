package server

import (
	"context"

	"github.com/interuss/dss/pkg/api/v1/ridpb"
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
func (s *Server) DeleteSubscription(
	ctx context.Context, req *ridpb.DeleteSubscriptionRequest) (
	*ridpb.DeleteSubscriptionResponse, error) {

	// TODO: simply verify the owner was set in an upper level.
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}
	//TODO: put the context with timeout into an interceptor so it's always set.
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.DeleteSubscription(ctx, id, owner, version)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not delete Subscription")
	}
	return &ridpb.DeleteSubscriptionResponse{
		Subscription: apiv1.ToSubscription(subscription),
	}, nil
}

// SearchSubscriptions queries for existing subscriptions in the given bounds.
func (s *Server) SearchSubscriptions(
	ctx context.Context, req *ridpb.SearchSubscriptionsRequest) (
	*ridpb.SearchSubscriptionsResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		if errors.Is(err, geoerr.ErrAreaTooLarge) {
			return nil, stacktrace.Propagate(err, "Invalid area")
		}
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscriptions, err := s.App.SearchSubscriptionsByOwner(ctx, cu, owner)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not search Subscriptions")
	}
	sp := make([]*ridpb.Subscription, len(subscriptions))
	for i := range subscriptions {
		sp[i] = apiv1.ToSubscription(subscriptions[i])
	}

	return &ridpb.SearchSubscriptionsResponse{
		Subscriptions: sp,
	}, nil
}

// GetSubscription gets a single subscription based on ID.
func (s *Server) GetSubscription(
	ctx context.Context, req *ridpb.GetSubscriptionRequest) (
	*ridpb.GetSubscriptionResponse, error) {

	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.GetSubscription(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get Subscription")
	}
	if subscription == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", req.GetId())
	}
	return &ridpb.GetSubscriptionResponse{
		Subscription: apiv1.ToSubscription(subscription),
	}, nil
}

// CreateSubscription creates a single subscription.
func (s *Server) CreateSubscription(
	ctx context.Context, req *ridpb.CreateSubscriptionRequest) (
	*ridpb.PutSubscriptionResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	if params == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Params not set")
	}
	if params.Callbacks == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required callbacks")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	extents, err := apiv1.FromVolume4D(params.Extents)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	if !s.EnableHTTP {
		err = ridmodels.ValidateURL(params.Callbacks.IdentificationServiceAreaUrl)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(
				err, dsserr.BadRequest, "Failed to validate IdentificationServiceAreaUrl")
		}
	}

	sub := &ridmodels.Subscription{
		ID:     id,
		Owner:  owner,
		URL:    params.Callbacks.IdentificationServiceAreaUrl,
		Writer: s.Locality,
	}

	if err := sub.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedSub, err := s.App.InsertSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not insert Subscription")
	}

	p := apiv1.ToSubscription(insertedSub)

	// Find ISAs that were in this subscription's area.
	isas, err := s.App.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not search ISAs")
	}

	// Convert the ISAs to protos.
	isaProtos := make([]*ridpb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i] = apiv1.ToIdentificationServiceArea(isa)
	}

	return &ridpb.PutSubscriptionResponse{
		Subscription: p,
		ServiceAreas: isaProtos,
	}, nil
}

// UpdateSubscription updates a single subscription.
func (s *Server) UpdateSubscription(
	ctx context.Context, req *ridpb.UpdateSubscriptionRequest) (
	*ridpb.PutSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	if params == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Params not set")
	}
	if params.Callbacks == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required callbacks")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	extents, err := apiv1.FromVolume4D(params.Extents)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D")
	}

	sub := &ridmodels.Subscription{
		ID:      id,
		Owner:   owner,
		URL:     params.Callbacks.IdentificationServiceAreaUrl,
		Version: version,
		Writer:  s.Locality,
	}

	if err := sub.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedSub, err := s.App.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not update Subscription")
	}

	p := apiv1.ToSubscription(insertedSub)

	// Find ISAs that were in this subscription's area.
	isas, err := s.App.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not search ISAs")
	}

	// Convert the ISAs to protos.
	isaProtos := make([]*ridpb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i] = apiv1.ToIdentificationServiceArea(isa)
	}

	return &ridpb.PutSubscriptionResponse{
		Subscription: p,
		ServiceAreas: isaProtos,
	}, nil
}
