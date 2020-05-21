package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// DeleteSubscription deletes an existing subscription.
func (s *Server) DeleteSubscription(
	ctx context.Context, req *ridpb.DeleteSubscriptionRequest) (
	*ridpb.DeleteSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.Subscription.Delete(ctx, dssmodels.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}
	p, err := subscription.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	return &ridpb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}

// SearchSubscriptions queries for existing subscriptions in the given bounds.
func (s *Server) SearchSubscriptions(
	ctx context.Context, req *ridpb.SearchSubscriptionsRequest) (
	*ridpb.SearchSubscriptionsResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		errMsg := fmt.Sprintf("bad area: %s", err)
		switch err.(type) {
		case *geo.ErrAreaTooLarge:
			return nil, dsserr.AreaTooLarge(errMsg)
		}
		return nil, dsserr.BadRequest(errMsg)
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscriptions, err := s.App.Subscription.Search(ctx, cu, owner)
	if err != nil {
		return nil, err
	}
	sp := make([]*ridpb.Subscription, len(subscriptions))
	for i := range subscriptions {
		sp[i], err = subscriptions[i].ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &ridpb.SearchSubscriptionsResponse{
		Subscriptions: sp,
	}, nil
}

// GetSubscription gets a single subscription based on ID.
func (s *Server) GetSubscription(
	ctx context.Context, req *ridpb.GetSubscriptionRequest) (
	*ridpb.GetSubscriptionResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.App.Subscription.Get(ctx, dssmodels.ID(req.GetId()))
	if err == sql.ErrNoRows {
		return nil, dsserr.NotFound(req.GetId())
	}
	if err != nil {
		return nil, err
	}
	p, err := subscription.ToProto()
	if err != nil {
		return nil, err
	}
	return &ridpb.GetSubscriptionResponse{
		Subscription: p,
	}, nil
}

func (s *Server) createOrUpdateSubscription(
	ctx context.Context, id string, version *dssmodels.Version, callbacks *ridpb.SubscriptionCallbacks, extents *ridpb.Volume4D) (
	*ridpb.PutSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if callbacks == nil {
		return nil, dsserr.BadRequest("missing required callbacks")
	}
	if extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(id),
		Owner:   owner,
		URL:     callbacks.IdentificationServiceAreaUrl,
		Version: version,
	}

	if err := sub.SetExtents(extents); err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedSub, err := s.App.Subscription.Insert(ctx, sub)
	if err != nil {
		return nil, err
	}

	p, err := insertedSub.ToProto()
	if err != nil {
		return nil, err
	}

	// Find ISAs that were in this subscription's area.
	isas, err := s.App.ISA.Search(ctx, sub.Cells, nil, nil)
	if err != nil {
		return nil, err
	}

	// Convert the ISAs to protos.
	isaProtos := make([]*ridpb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i], err = isa.ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &ridpb.PutSubscriptionResponse{
		Subscription: p,
		ServiceAreas: isaProtos,
	}, nil
}

// CreateSubscription creates a single subscription.
func (s *Server) CreateSubscription(
	ctx context.Context, req *ridpb.CreateSubscriptionRequest) (
	*ridpb.PutSubscriptionResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), nil, params.Callbacks, params.Extents)
}

// UpdateSubscription updates a single subscription.
func (s *Server) UpdateSubscription(
	ctx context.Context, req *ridpb.UpdateSubscriptionRequest) (
	*ridpb.PutSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), version, params.Callbacks, params.Extents)
}
