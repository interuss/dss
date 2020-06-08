package scd

import (
	"context"
	"fmt"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"
)

var (
	DefaultClock = clockwork.NewRealClock()
)

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
	extents, err := dssmodels.Volume4DFromSCDProto(params.GetExtents())
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

	// Validate requested Subscription
	if !sub.NotifyForOperations && !sub.NotifyForConstraints {
		return nil, dsserr.BadRequest("no notification triggers requested for Subscription")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := sub.AdjustTimeRange(DefaultClock.Now(), sub); err != nil {
		return nil, err
	}

	var result *scdpb.PutSubscriptionResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// TODO: validate against DependentOperations when available

		// Store Subscription model
		sub, ops, err := store.UpsertSubscription(ctx, sub)
		if err != nil {
			return err
		}
		if sub == nil {
			return dsserr.Internal(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
		}
		for _, op := range ops {
			if op.Owner != owner {
				op.OVN = scdmodels.OVN("")
			}
		}

		// Convert Subscription to proto
		p, err := sub.ToProto()
		if err != nil {
			return dsserr.Internal(err.Error())
		}
		result = &scdpb.PutSubscriptionResponse{
			Subscription: p,
		}
		for _, op := range ops {
			if op.Owner != owner {
				op.OVN = scdmodels.OVN("")
			}
			pop, _ := op.ToProto()
			result.Operations = append(result.Operations, pop)
		}

		return nil
	}

	err = scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
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

	var response *scdpb.GetSubscriptionResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// Get Subscription from Store
		sub, err := store.GetSubscription(ctx, id, owner)
		if err != nil {
			return err
		}

		// Convert Subscription to proto
		p, err := sub.ToProto()
		if err != nil {
			return dsserr.Internal("unable to convert Subscription to proto")
		}

		// Return response to client
		response = &scdpb.GetSubscriptionResponse{
			Subscription: p,
		}

		return nil
	}

	err := scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *scdpb.QuerySubscriptionsRequest) (*scdpb.SearchSubscriptionsResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.SearchSubscriptionsResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// Perform search query on Store
		subs, err := store.SearchSubscriptions(ctx, vol4)
		if err != nil {
			return err
		}

		// Return response to client
		response = &scdpb.SearchSubscriptionsResponse{}
		for _, sub := range subs {
			if sub.Owner == owner {
				p, err := sub.ToProto()
				if err != nil {
					return dsserr.Internal("error converting Subscription model to proto")
				}
				response.Subscriptions = append(response.Subscriptions, p)
			}
		}

		return nil
	}

	err = scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
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

	var response *scdpb.DeleteSubscriptionResponse
	action := func(ctx context.Context, store scdstore.Store) (err error) {
		// Delete Subscription in Store
		sub, err := store.DeleteSubscription(ctx, id, owner, scdmodels.Version(0))
		if err != nil {
			return err
		}
		if sub == nil {
			return dsserr.Internal(fmt.Sprintf("DeleteSubscription returned no Subscription for ID: %s", id))
		}

		// Convert deleted Subscription to proto
		p, err := sub.ToProto()
		if err != nil {
			return dsserr.Internal("error converting Subscription model to proto")
		}

		// Create response for client
		response = &scdpb.DeleteSubscriptionResponse{
			Subscription: p,
		}

		return nil
	}

	err := scdstore.PerformOperationWithRetries(ctx, a.Transactor, action, 0)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}
