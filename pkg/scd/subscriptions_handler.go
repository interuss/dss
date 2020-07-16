package scd

import (
	"context"
	"fmt"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
)

var (
	DefaultClock = clockwork.NewRealClock()
)

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, req *scdpb.PutSubscriptionRequest) (*scdpb.PutSubscriptionResponse, error) {
	// Retrieve Subscription ID
	id := scdmodels.ID(req.GetSubscriptionid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}

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

	subreq := &scdmodels.Subscription{
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
	if !subreq.NotifyForOperations && !subreq.NotifyForConstraints {
		return nil, dsserr.BadRequest("no notification triggers requested for Subscription")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := subreq.AdjustTimeRange(DefaultClock.Now(), subreq); err != nil {
		return nil, err
	}

	var result *scdpb.PutSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Check existing Subscription (if any)
		old, err := r.GetSubscription(ctx, subreq.ID)
		if err != nil {
			return err
		}

		if old == nil {
			// There is no previous Subscription (this is a creation attempt)
			if !subreq.Version.Empty() {
				// The user wants to update an existing Subscription, but one wasn't found.
				return dsserr.NotFound(subreq.ID.String())
			}
		} else {
			// There is a previous Subscription (this is an update attempt)
			switch {
			case subreq.Version.Empty():
				// The user wants to create a new Subscription but it already exists.
				return dsserr.AlreadyExists(subreq.ID.String())
			case !subreq.Version.Matches(old.Version):
				// The user wants to update a Subscription but the version doesn't match.
				return dsserr.VersionMismatch("old version")
			case old.Owner != subreq.Owner:
				return dsserr.PermissionDenied(fmt.Sprintf("Subscription is owned by %s", old.Owner))
			}

			// TODO: validate against DependentOperations when available
		}

		// Store Subscription model
		sub, err := r.UpsertSubscription(ctx, subreq)
		if err != nil {
			return err
		}
		if sub == nil {
			return dsserr.Internal(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
		}

		// Find relevant Operations
		var relevantOperations []*scdmodels.Operation
		if len(sub.Cells) > 0 {
			ops, err := r.SearchOperations(ctx, &dssmodels.Volume4D{
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
				return err
			}
			relevantOperations = ops
		}

		// Convert Subscription to proto
		p, err := sub.ToProto()
		if err != nil {
			return dsserr.Internal(err.Error())
		}
		result = &scdpb.PutSubscriptionResponse{
			Subscription: p,
		}

		if sub.NotifyForOperations {
			// Attach Operations to response
			for _, op := range relevantOperations {
				if op.Owner != owner {
					op.OVN = scdmodels.OVN("")
				}
				pop, _ := op.ToProto()
				result.Operations = append(result.Operations, pop)
			}
		}

		if sub.NotifyForConstraints {
			// Query relevant Constraints
			constraints, err := r.SearchConstraints(ctx, extents)
			if err != nil {
				return err
			}

			// Attach Constraints to response
			for _, constraint := range constraints {
				p, err := constraint.ToProto()
				if err != nil {
					return err
				}
				if constraint.Owner != owner {
					p.Ovn = ""
				}
				result.Constraints = append(result.Constraints, p)
			}
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
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
	id := scdmodels.ID(req.GetSubscriptionid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.GetSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get Subscription from Store
		sub, err := r.GetSubscription(ctx, id)
		if err != nil {
			return err
		}
		if sub == nil {
			return dsserr.NotFound(id.String())
		}

		// Check if the client is authorized to view this Subscription
		if owner != sub.Owner {
			return dsserr.PermissionDenied("Subscription owned by a different client")
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

	err := a.Store.Transact(ctx, action)
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
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		subs, err := r.SearchSubscriptions(ctx, vol4)
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

	err = a.Store.Transact(ctx, action)
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
	id := scdmodels.ID(req.GetSubscriptionid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var response *scdpb.DeleteSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Check to make sure it's ok to delete this Subscription
		old, err := r.GetSubscription(ctx, id)
		switch {
		case err != nil:
			return err
		case old == nil: // Return a 404 here.
			return dsserr.NotFound(id.String())
		case old.Owner != owner:
			return dsserr.PermissionDenied(fmt.Sprintf("Subscription is owned by %s", old.Owner))
		}

		// Delete Subscription in Store
		err = r.DeleteSubscription(ctx, id)
		if err != nil {
			return err
		}

		// Convert deleted Subscription to proto
		p, err := old.ToProto()
		if err != nil {
			return dsserr.Internal("error converting Subscription model to proto")
		}

		// Create response for client
		response = &scdpb.DeleteSubscriptionResponse{
			Subscription: p,
		}

		return nil
	}

	err := a.Store.Transact(ctx, action)
	if err != nil {
		// TODO: wrap err in dss.Internal?
		return nil, err
	}

	return response, nil
}
