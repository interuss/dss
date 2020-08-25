package scd

import (
	"context"
	"database/sql"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/palantir/stacktrace"
)

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *scdpb.DeleteConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	// Retrieve Constraint ID
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.ChangeConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Make sure deletion request is valid
		old, err := r.GetConstraint(ctx, id)
		switch {
		case err == sql.ErrNoRows:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
		case err != nil:
			return stacktrace.Propagate(err, "Unable to get Constraint from repo")
		case old.Owner != owner:
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to delete", old.Owner, owner)
		}

		// Find Subscriptions that may overlap the Constraint's Volume4D
		allsubs, err := r.SearchSubscriptions(ctx, &dssmodels.Volume4D{
			StartTime: old.StartTime,
			EndTime:   old.EndTime,
			SpatialVolume: &dssmodels.Volume3D{
				AltitudeHi: old.AltitudeUpper,
				AltitudeLo: old.AltitudeLower,
				Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
					return old.Cells, nil
				}),
			}})
		if err != nil {
			return stacktrace.Propagate(err, "Unable to search Subscriptions in repo")
		}

		// Limit Subscription notifications to only those interested in Constraints
		var subs repos.Subscriptions
		for _, sub := range allsubs {
			if sub.NotifyForConstraints {
				subs = append(subs, sub)
			}
		}

		// Delete Constraint in repo
		err = r.DeleteConstraint(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to delete Constraint from repo")
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to increment notification indices")
		}

		// Convert deleted Constraint to proto
		constraintProto, err := old.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Constraint to proto")
		}

		// Return response to client
		response = &scdpb.ChangeConstraintReferenceResponse{
			ConstraintReference: constraintProto,
			Subscribers:         makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// GetConstraintReference returns a single constraint ref for the given ID.
func (a *Server) GetConstraintReference(ctx context.Context, req *scdpb.GetConstraintReferenceRequest) (*scdpb.GetConstraintReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.GetConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		constraint, err := r.GetConstraint(ctx, id)
		switch {
		case err == sql.ErrNoRows:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
		case err != nil:
			return stacktrace.Propagate(err, "Unable to get Constraint from repo")
		}

		if constraint.Owner != owner {
			constraint.OVN = scdmodels.OVN("")
		}

		// Convert retrieved Constraint to proto
		p, err := constraint.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Constraint to proto")
		}

		// Return response to client
		response = &scdpb.GetConstraintReferenceResponse{
			ConstraintReference: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// PutConstraintReference creates a single contraint ref.
func (a *Server) PutConstraintReference(ctx context.Context, req *scdpb.PutConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityuuid())
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var (
		params  = req.GetParams()
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	// TODO: factor out logic below into common multi-vol4d parser and reuse with PutOperationReference
	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extents")
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to union extents")
	}

	if uExtent.StartTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_end from extents")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	var response *scdpb.ChangeConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get existing Constraint, if any, and validate request
		old, err := r.GetConstraint(ctx, id)
		switch {
		case err == sql.ErrNoRows:
			// No existing Constraint; verify that creation was requested
			if params.OldVersion != 0 {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %d does not exist", params.OldVersion)
			}
		case err != nil:
			return stacktrace.Propagate(err, "Could not get Constraint from repo")
		}
		if old != nil {
			if old.Owner != owner {
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
					"Constraint owned by %s, but %s attempted to modify", old.Owner, owner)
			}
			if old.Version != scdmodels.Version(params.OldVersion) {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
					"Current version is %d but client specified version %d", old.Version, params.OldVersion)
			}
		}

		// Compute total affected Volume4D for notification purposes
		var notifyVol4 *dssmodels.Volume4D
		if old == nil {
			notifyVol4 = uExtent
		} else {
			oldVol4 := &dssmodels.Volume4D{
				StartTime: old.StartTime,
				EndTime:   old.EndTime,
				SpatialVolume: &dssmodels.Volume3D{
					AltitudeHi: old.AltitudeUpper,
					AltitudeLo: old.AltitudeLower,
					Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
						return old.Cells, nil
					}),
				}}
			notifyVol4, err = dssmodels.UnionVolumes4D(uExtent, oldVol4)
			if err != nil {
				return stacktrace.Propagate(err, "Error constructing 4D volumes union")
			}
		}

		// Upsert the Constraint
		constraint, err := r.UpsertConstraint(ctx, &scdmodels.Constraint{
			ID:      id,
			Owner:   owner,
			Version: scdmodels.Version(params.OldVersion + 1),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,

			USSBaseURL: params.UssBaseUrl,
			Cells:      cells,
		})
		if err != nil {
			return err
		}

		// Find Subscriptions that may need to be notified
		allsubs, err := r.SearchSubscriptions(ctx, notifyVol4)
		if err != nil {
			return err
		}

		// Limit Subscription notifications to only those interested in Constraints
		var subs repos.Subscriptions
		for _, sub := range allsubs {
			if sub.NotifyForConstraints {
				subs = append(subs, sub)
			}
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return err
		}

		// Convert upserted Constraint to proto
		p, err := constraint.ToProto()
		if err != nil {
			return err
		}

		// Return response to client
		response = &scdpb.ChangeConstraintReferenceResponse{
			ConstraintReference: p,
			Subscribers:         makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// QueryConstraintReferences queries existing contraint refs in the given
// bounds.
func (a *Server) QueryConstraintReferences(ctx context.Context, req *scdpb.QueryConstraintReferencesRequest) (*scdpb.SearchConstraintReferencesResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, err
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.SearchConstraintReferencesResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		constraints, err := r.SearchConstraints(ctx, vol4)
		if err != nil {
			return err
		}

		// Create response for client
		response = &scdpb.SearchConstraintReferencesResponse{}
		for _, constraint := range constraints {
			p, err := constraint.ToProto()
			if err != nil {
				return err
			}
			if constraint.Owner != owner {
				p.Ovn = ""
			}
			response.ConstraintReferences = append(response.ConstraintReferences, p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}
