package scd

import (
	"context"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *restapi.DeleteConstraintReferenceRequest,
) restapi.DeleteConstraintReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteConstraintReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Retrieve Constraint ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.DeleteConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.ChangeConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Make sure deletion request is valid
		old, err := r.GetConstraint(ctx, id)
		switch {
		case err == pgx.ErrNoRows:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
		case err != nil:
			return stacktrace.Propagate(err, "Unable to get Constraint from repo")
		case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"Constraint owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
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
		subs := repos.Subscriptions{}
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

		// Return response to client
		response = &restapi.ChangeConstraintReferenceResponse{
			ConstraintReference: *old.ToRest(),
			Subscribers:         makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete constraint")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteConstraintReferenceResponseSet{Response403: errResp}
		case dsserr.BadRequest:
			return restapi.DeleteConstraintReferenceResponseSet{Response400: errResp}
		case dsserr.NotFound:
			return restapi.DeleteConstraintReferenceResponseSet{Response404: errResp}
		default:
			return restapi.DeleteConstraintReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.DeleteConstraintReferenceResponseSet{Response200: response}
}

// GetConstraintReference returns a single constraint ref for the given ID.
func (a *Server) GetConstraintReference(ctx context.Context, req *restapi.GetConstraintReferenceRequest,
) restapi.GetConstraintReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetConstraintReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.GetConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		constraint, err := r.GetConstraint(ctx, id)
		switch {
		case err == pgx.ErrNoRows:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
		case err != nil:
			return stacktrace.Propagate(err, "Unable to get Constraint from repo")
		}

		if constraint.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			constraint.OVN = scdmodels.NoOvnPhrase
		}

		// Return response to client
		response = &restapi.GetConstraintReferenceResponse{
			ConstraintReference: *constraint.ToRest(),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not get constraint")
		if stacktrace.GetCode(err) == dsserr.NotFound {
			return restapi.GetConstraintReferenceResponseSet{Response404: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.GetConstraintReferenceResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.GetConstraintReferenceResponseSet{Response200: response}
}

func (a *Server) CreateConstraintReference(ctx context.Context, req *restapi.CreateConstraintReferenceRequest,
) restapi.CreateConstraintReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.CreateConstraintReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.CreateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.CreateConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	res, err := a.PutConstraintReference(ctx, *req.Auth.ClientID, req.Entityid, "", req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put constraint")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.CreateConstraintReferenceResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.CreateConstraintReferenceResponseSet{Response409: errResp}
		case dsserr.BadRequest:
			return restapi.CreateConstraintReferenceResponseSet{Response400: errResp}
		default:
			return restapi.CreateConstraintReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.CreateConstraintReferenceResponseSet{Response201: res}
}

func (a *Server) UpdateConstraintReference(ctx context.Context, req *restapi.UpdateConstraintReferenceRequest,
) restapi.UpdateConstraintReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.UpdateConstraintReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.UpdateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.UpdateConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	res, err := a.PutConstraintReference(ctx, *req.Auth.ClientID, req.Entityid, req.Ovn, req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put constraint")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateConstraintReferenceResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.UpdateConstraintReferenceResponseSet{Response409: errResp}
		case dsserr.BadRequest:
			return restapi.UpdateConstraintReferenceResponseSet{Response400: errResp}
		default:
			return restapi.UpdateConstraintReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.UpdateConstraintReferenceResponseSet{Response200: res}
}

// PutConstraintReference inserts or updates a Constraint.
// If the ovn argument is empty (""), it will attempt to create a new Constraint.
func (a *Server) PutConstraintReference(ctx context.Context, manager string, entityid restapi.EntityID, ovn restapi.EntityOVN, params *restapi.PutConstraintReferenceParameters,
) (*restapi.ChangeConstraintReferenceResponse, error) {
	id, err := dssmodels.IDFromString(string(entityid))

	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", entityid)
	}

	var extents = make([]*dssmodels.Volume4D, len(params.Extents))

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	if !a.AllowHTTPBaseUrls {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	// TODO: factor out logic below into common multi-vol4d parser and reuse with PutOperationReference
	for idx, extent := range params.Extents {
		cExtent, err := dssmodels.Volume4DFromSCDRest(&extent)
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

	var response *restapi.ChangeConstraintReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		var version int32 // Version of the Constraint (0 means creation requested).

		// Get existing Constraint, if any, and validate request
		old, err := r.GetConstraint(ctx, id)
		switch {
		case err == pgx.ErrNoRows:
			// No existing Constraint; verify that creation was requested
			if ovn != "" {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Old version %s does not exist", ovn)
			}
			version = 0
		case err != nil:
			return stacktrace.Propagate(err, "Could not get Constraint from repo")
		}
		if old != nil {
			if old.Manager != dssmodels.Manager(manager) {
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
			Manager: dssmodels.Manager(manager),
			Version: scdmodels.VersionNumber(version + 1),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,

			USSBaseURL: string(params.UssBaseUrl),
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
		subs := repos.Subscriptions{}
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

		// Return response to client
		response = &restapi.ChangeConstraintReferenceResponse{
			ConstraintReference: *constraint.ToRest(),
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
func (a *Server) QueryConstraintReferences(ctx context.Context, req *restapi.QueryConstraintReferencesRequest,
) restapi.QueryConstraintReferencesResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.QueryConstraintReferencesResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.QueryConstraintReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return restapi.QueryConstraintReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest"))}}
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryConstraintReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model"))}}
	}

	var response *restapi.QueryConstraintReferencesResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		constraints, err := r.SearchConstraints(ctx, vol4)
		if err != nil {
			return err
		}

		// Create response for client
		response = &restapi.QueryConstraintReferencesResponse{
			ConstraintReferences: make([]restapi.ConstraintReference, 0, len(constraints)),
		}
		for _, constraint := range constraints {
			p := constraint.ToRest()
			if constraint.Manager != dssmodels.Manager(*req.Auth.ClientID) {
				noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
				p.Ovn = &noOvnPhrase
			}
			response.ConstraintReferences = append(response.ConstraintReferences, *p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return restapi.QueryConstraintReferencesResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.QueryConstraintReferencesResponseSet{Response200: response}
}
