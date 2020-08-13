package scd

import (
	"context"
	"fmt"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/palantir/stacktrace"
)

var (
	DefaultClock = clockwork.NewRealClock()
)

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, req *scdpb.PutSubscriptionRequest) (*scdpb.PutSubscriptionResponse, error) {
	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(req.GetSubscriptionid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var (
		params = req.GetParams()
	)

	// Parse extents
	extents, err := dssmodels.Volume4DFromSCDProto(params.GetExtents())
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to parse extents")
	}

	// Construct requested Subscription model
	cells, err := extents.CalculateSpatialCovering()
	switch err {
	case nil, geo.ErrMissingSpatialVolume, geo.ErrMissingFootprint:
		// We may be able to fill these values from a previous Subscription or via defaults.
	default:
		return nil, stacktrace.Propagate(err, "Invalid area")
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
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "No notification triggers requested for Subscription")
	}

	var result *scdpb.PutSubscriptionResponse
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

		if old == nil {
			// There is no previous Subscription (this is a creation attempt)
			if !subreq.Version.Empty() {
				// The user wants to update an existing Subscription, but one wasn't found.
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", subreq.ID.String())
			}
		} else {
			// There is a previous Subscription (this is an update attempt)
			switch {
			case subreq.Version.Empty():
				// The user wants to create a new Subscription but it already exists.
				return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", subreq.ID.String())
			case !subreq.Version.Matches(old.Version):
				// The user wants to update a Subscription but the version doesn't match.
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %d is not current", subreq.Version)
			case old.Owner != subreq.Owner:
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client")
			}

			subreq.NotificationIndex = old.NotificationIndex

			// TODO(#386): validate against DependentOperations
		}

		// Store Subscription model
		sub, err := r.UpsertSubscription(ctx, subreq)
		if err != nil {
			return stacktrace.Propagate(err, "Could not upsert Subscription into repo")
		}
		if sub == nil {
			return stacktrace.NewError(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
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
				return stacktrace.Propagate(err, "Could not search Operations in repo")
			}
			relevantOperations = ops
		}

		// Get dependent Operations
		dependentOps, err := r.GetDependentOperations(ctx, sub.ID)
		if err != nil {
			return stacktrace.Propagate(err, "Could not find dependent Operations")
		}

		// Convert Subscription to proto
		p, err := sub.ToProto(dependentOps)
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Subscription to proto")
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
				return stacktrace.Propagate(err, "Could not search Constraints in repo")
			}

			// Attach Constraints to response
			for _, constraint := range constraints {
				p, err := constraint.ToProto()
				if err != nil {
					return stacktrace.Propagate(err, "Could not convert Constraint to proto")
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
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	// Return response to client
	return result, nil
}

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *scdpb.GetSubscriptionRequest) (*scdpb.GetSubscriptionResponse, error) {
	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(req.GetSubscriptionid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.GetSubscriptionResponse
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
		if owner != sub.Owner {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client")
		}

		// Get dependent Operations
		dependentOps, err := r.GetDependentOperations(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not find dependent Operations")
		}

		// Convert Subscription to proto
		p, err := sub.ToProto(dependentOps)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to convert Subscription to proto")
		}

		// Return response to client
		response = &scdpb.GetSubscriptionResponse{
			Subscription: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *scdpb.QuerySubscriptionsRequest) (*scdpb.SearchSubscriptionsResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert to internal geometry model")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.SearchSubscriptionsResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		subs, err := r.SearchSubscriptions(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Error searching Subscriptions in repo")
		}

		// Return response to client
		response = &scdpb.SearchSubscriptionsResponse{}
		for _, sub := range subs {
			if sub.Owner == owner {
				// Get dependent Operations
				dependentOps, err := r.GetDependentOperations(ctx, sub.ID)
				if err != nil {
					return stacktrace.Propagate(err, "Could not find dependent Operations")
				}

				p, err := sub.ToProto(dependentOps)
				if err != nil {
					return stacktrace.Propagate(err, "Error converting Subscription model to proto")
				}
				response.Subscriptions = append(response.Subscriptions, p)
			}
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// DeleteSubscription deletes a single subscription for a given ID at the
// specified version.
func (a *Server) DeleteSubscription(ctx context.Context, req *scdpb.DeleteSubscriptionRequest) (*scdpb.DeleteSubscriptionResponse, error) {
	// Retrieve Subscription ID
	id, err := dssmodels.IDFromString(req.GetSubscriptionid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.DeleteSubscriptionResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Check to make sure it's ok to delete this Subscription
		old, err := r.GetSubscription(ctx, id)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Could not get Subscription from repo")
		case old == nil: // Return a 404 here.
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
		case old.Owner != owner:
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client")
		}

		// Get dependent Operations
		dependentOps, err := r.GetDependentOperations(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not find dependent Operations")
		}
		if len(dependentOps) > 0 {
			return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscriptions with dependent Operations may not be removed")
		}

		// Delete Subscription in repo
		err = r.DeleteSubscription(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not delete Subscription from repo")
		}

		// Convert deleted Subscription to proto
		p, err := old.ToProto(dependentOps)
		if err != nil {
			return stacktrace.Propagate(err, "Error converting Subscription model to proto")
		}

		// Create response for client
		response = &scdpb.DeleteSubscriptionResponse{
			Subscription: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}
