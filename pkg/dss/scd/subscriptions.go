package scd

import (
	"context"
	"fmt"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// PutSubscription creates a single subscription.
func (a *Server) putSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error) {
	// Store Subscription model
	return a.Store.UpsertSubscription(ctx, sub)
}

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, req *scdpb.PutSubscriptionRequest) (*scdpb.PutSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		params = req.GetParams()
	)

	// Parse extents
	extents, err := params.GetExtents().ToCommon()
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("unable to parse extents: %s", err))
	}

	// Construct requested Subscription model
	cells, err := extents.CalculateSpatialCovering()
	switch err {
	case nil, dssmodels.ErrMissingSpatialVolume, dssmodels.ErrMissingFootprint:
		// All good, let's go ahead.
	default:
		return nil, dssErrorOfAreaError(err)
	}

	sub := &scdmodels.Subscription{
		ID:      id,
		Owner:   owner,
		Version: scdmodels.Version(params.OldVersion),

		StartTime:  extents.StartTime,
		EndTime:    extents.EndTime,
		AltitudeLo: extents.SpatialVolume.AltitudeLo,
		AltitudeHi: extents.SpatialVolume.AltitudeHi,
		Cells:      cells,

		BaseURL:              params.UssBaseUrl,
		NotifyForOperations:  params.NotifyForOperations,
		NotifyForConstraints: params.NotifyForConstraints,
	}

	// Store Subscription model
	sub, ops, err := a.putSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
	}
	for _, op := range ops {
		if op.Owner != owner {
			op.OVN = scdmodels.OVN("")
		}
	}

	// Convert Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	result := &scdpb.PutSubscriptionResponse{
		Subscription: p,
	}
	for _, op := range ops {
		if op.Owner != owner {
			op.OVN = scdmodels.OVN("")
		}
		pop, _ := op.ToProto()
		result.Operations = append(result.Operations, pop)
	}
	// Return response to client
	return result, nil
}

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *scdpb.GetSubscriptionRequest) (*scdpb.GetSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Get Subscription from Store
	sub, err := a.Store.GetSubscription(ctx, id, owner)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("GetSubscription returned no Subscription for ID: %s", id))
	}

	// Convert Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.GetSubscriptionResponse{
		Subscription: p,
	}, nil
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *scdpb.QuerySubscriptionsRequest) (*scdpb.SearchSubscriptionsResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := aoi.ToCommon()
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Extract S2 cells from area of interest
	cells, err := vol4.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Perform search query on Store
	subs, err := a.Store.SearchSubscriptions(ctx, cells, owner) //TODO: incorporate time bounds into query
	if err != nil {
		return nil, err
	}
	if subs == nil {
		return nil, dsserr.Internal("SearchSubscriptions returned nil subscriptions")
	}

	// Return response to client
	response := &scdpb.SearchSubscriptionsResponse{}
	for _, sub := range subs {
		p, err := sub.ToProto()
		if err != nil {
			return nil, dsserr.Internal("error converting Subscription model to proto")
		}
		response.Subscriptions = append(response.Subscriptions, p)
	}
	return response, nil
}

// DeleteSubscription deletes a single subscription for a given ID at the
// specified version.
func (a *Server) DeleteSubscription(ctx context.Context, req *scdpb.DeleteSubscriptionRequest) (*scdpb.DeleteSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Delete Subscription in Store
	sub, err := a.Store.DeleteSubscription(ctx, id, owner, scdmodels.Version(0))
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("DeleteSubscription returned no Subscription for ID: %s", id))
	}

	// Convert deleted Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}
