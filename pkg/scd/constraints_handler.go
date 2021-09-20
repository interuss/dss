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
	"github.com/interuss/stacktrace"
)

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *scdpb.DeleteConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	// Retrieve Constraint ID
	id, err := dssmodels.IDFromString(req.GetEntityid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityid())
	}

	// Retrieve ID of client making call
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
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
		case old.Manager != manager:
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to delete", old.Manager, manager)
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
	id, err := dssmodels.IDFromString(req.GetEntityid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.GetEntityid())
	}

	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
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

		if constraint.Manager != manager {
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

func (a *Server) CreateConstraintReference(ctx context.Context, req *scdpb.CreateConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return a.PutConstraintReference(ctx, req.GetEntityid(), "", req.GetParams())
}

func (a *Server) UpdateConstraintReference(ctx context.Context, req *scdpb.UpdateConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return a.PutConstraintReference(ctx, req.GetEntityid(), req.GetOvn(), req.GetParams())
}

// PutConstraintReference inserts or updates a Constraint.
// If the ovn argument is empty (""), it will attempt to create a new Constraint.
func (a *Server) PutConstraintReference(ctx context.Context, entityid string, ovn string, params *scdpb.PutConstraintReferenceParameters) (*scdpb.ChangeConstraintReferenceResponse, error) {
	id, err := dssmodels.IDFromString(entityid)

	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", entityid)
	}

	// Retrieve ID of client making call
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	if !a.EnableHTTP {
		err = scdmodels.ValidateUSSBaseURL(params.UssBaseUrl)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	// TODO: factor out logic below into common multi-vol4d parser and reuse with PutOperationReference
	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extent %d", idx)
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
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	var response *scdpb.ChangeConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		var version int32 // Version of the Constraint (0 means creation requested).

		// Get existing Constraint, if any, and validate request
		old, err := r.GetConstraint(ctx, id)
		switch {
		case err == sql.ErrNoRows:
			// No existing Constraint; verify that creation was requested
			if ovn != "" {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %s does not exist", ovn)
			}
			version = 0
		case err != nil:
			return stacktrace.Propagate(err, "Could not get Constraint from repo")
		}
		if old != nil {
			if old.Manager != manager {
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
					"Constraint owned by %s, but %s attempted to modify", old.Manager, manager)
			}
			if old.OVN != scdmodels.OVN(ovn) {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
					"Current version is %s but client specified version %s", old.OVN, ovn)
			}
			version = int32(old.Version)
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
			Manager: manager,
			Version: scdmodels.Version(version + 1),

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
func (a *Server) QueryConstraintReferences(ctx context.Context, req *scdpb.QueryConstraintReferencesRequest) (*scdpb.QueryConstraintReferencesResponse, error) {
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
	manager, ok := auth.ManagerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager from context")
	}

	var response *scdpb.QueryConstraintReferencesResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		constraints, err := r.SearchConstraints(ctx, vol4)
		if err != nil {
			return err
		}

		// Create response for client
		response = &scdpb.QueryConstraintReferencesResponse{}
		for _, constraint := range constraints {
			p, err := constraint.ToProto()
			if err != nil {
				return err
			}
			if constraint.Manager != manager {
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
