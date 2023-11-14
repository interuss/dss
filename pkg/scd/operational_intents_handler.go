package scd

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

// DeleteOperationalIntentReference deletes a single operational intent ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationalIntentReference(ctx context.Context, req *restapi.DeleteOperationalIntentReferenceRequest,
) restapi.DeleteOperationalIntentReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Retrieve OperationalIntent ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.ChangeOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get OperationalIntent to delete
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationIntent from repo")
		}
		if old == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		// Validate deletion request
		if old.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
		}

		// Get the Subscription supporting the OperationalIntent
		sub, err := r.GetSubscription(ctx, old.SubscriptionID)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationalIntent's Subscription from repo")
		}
		if sub == nil {
			return stacktrace.NewError("OperationalIntent's Subscription missing from repo")
		}

		removeImplicitSubscription := false
		if sub.ImplicitSubscription {
			// Get the Subscription's dependent OperationalIntents
			dependentOps, err := r.GetDependentOperationalIntents(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
			}
			if len(dependentOps) == 0 {
				return stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
			} else if len(dependentOps) == 1 {
				removeImplicitSubscription = true
			}
		}

		// Find Subscriptions that may overlap the OperationalIntent's Volume4D
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

		// Limit Subscription notifications to only those interested in OperationalIntents
		subs := repos.Subscriptions{}
		for _, s := range allsubs {
			if s.NotifyForOperationalIntents {
				subs = append(subs, s)
			}
		}

		// Increment notification indices for Subscriptions to be notified
		if err := subs.IncrementNotificationIndices(ctx, r); err != nil {
			return stacktrace.Propagate(err, "Unable to increment notification indices")
		}

		// Delete OperationalIntent from repo
		if err := r.DeleteOperationalIntent(ctx, id); err != nil {
			return stacktrace.Propagate(err, "Unable to delete OperationalIntent from repo")
		}

		if removeImplicitSubscription {
			// Automatically remove a now-unused implicit Subscription
			err = r.DeleteSubscription(ctx, sub.ID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to delete associated implicit Subscription")
			}
		}

		// Return response to client
		response = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *old.ToRest(),
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete operational intent")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.NotFound:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response404: errResp}
		default:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.DeleteOperationalIntentReferenceResponseSet{Response200: response}
}

// GetOperationalIntentReference returns a single operation intent ref for the given ID.
func (a *Server) GetOperationalIntentReference(ctx context.Context, req *restapi.GetOperationalIntentReferenceRequest,
) restapi.GetOperationalIntentReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.GetOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		op, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get OperationalIntent from repo")
		}
		if op == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent %s not found", id)
		}

		if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
			op.OVN = scdmodels.NoOvnPhrase
		}

		response = &restapi.GetOperationalIntentReferenceResponse{
			OperationalIntentReference: *op.ToRest(),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not get operational intent")
		if stacktrace.GetCode(err) == dsserr.NotFound {
			return restapi.GetOperationalIntentReferenceResponseSet{Response404: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.GetOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.GetOperationalIntentReferenceResponseSet{Response200: response}
}

// QueryOperationalIntentReferences queries existing operational intent refs in the given
// bounds.
func (a *Server) QueryOperationalIntentReferences(ctx context.Context, req *restapi.QueryOperationalIntentReferencesRequest,
) restapi.QueryOperationalIntentReferencesResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.QueryOperationalIntentReferencesResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	// Retrieve the area of interest parameter
	aoi := req.Body.AreaOfInterest
	if aoi == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest"))}}
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	var response *restapi.QueryOperationalIntentReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		ops, err := r.SearchOperationalIntents(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to query for OperationalIntents in repo")
		}

		// Create response for client
		response = &restapi.QueryOperationalIntentReferenceResponse{
			OperationalIntentReferences: make([]restapi.OperationalIntentReference, 0, len(ops)),
		}
		for _, op := range ops {
			p := op.ToRest()
			if op.Manager != dssmodels.Manager(*req.Auth.ClientID) {
				noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
				p.Ovn = &noOvnPhrase
			}
			response.OperationalIntentReferences = append(response.OperationalIntentReferences, *p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not query operational intent")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		}
		return restapi.QueryOperationalIntentReferencesResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.QueryOperationalIntentReferencesResponseSet{Response200: response}
}

func (a *Server) CreateOperationalIntentReference(ctx context.Context, req *restapi.CreateOperationalIntentReferenceRequest,
) restapi.CreateOperationalIntentReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.CreateOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	respOK, respConflict, err := a.PutOperationalIntentReference(ctx, *req.Auth.ClientID, req.Entityid, "", req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response400: errResp}
		case dsserr.VersionMismatch:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response409: &restapi.AirspaceConflictResponse{
				Message: dsserr.Handle(ctx, err)}}
		case dsserr.MissingOVNs:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.CreateOperationalIntentReferenceResponseSet{Response201: respOK}
}

func (a *Server) UpdateOperationalIntentReference(ctx context.Context, req *restapi.UpdateOperationalIntentReferenceRequest,
) restapi.UpdateOperationalIntentReferenceResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.UpdateOperationalIntentReferenceResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.BodyParseError != nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	respOK, respConflict, err := a.PutOperationalIntentReference(ctx, *req.Auth.ClientID, req.Entityid, req.Ovn, req.Body)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put subscription")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: errResp}
		case dsserr.VersionMismatch:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response409: &restapi.AirspaceConflictResponse{
				Message: dsserr.Handle(ctx, err)}}
		case dsserr.MissingOVNs:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	return restapi.UpdateOperationalIntentReferenceResponseSet{Response200: respOK}
}

// PutOperationalIntentReference inserts or updates an Operational Intent.
// If the ovn argument is empty (""), it will attempt to create a new Operational Intent.
func (a *Server) PutOperationalIntentReference(ctx context.Context, manager string, entityid restapi.EntityID, ovn restapi.EntityOVN, params *restapi.PutOperationalIntentReferenceParameters,
) (*restapi.ChangeOperationalIntentReferenceResponse, *restapi.AirspaceConflictResponse, error) {
	id, err := dssmodels.IDFromString(string(entityid))
	if err != nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", entityid)
	}

	var (
		extents = make([]*dssmodels.Volume4D, len(params.Extents))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	if !a.EnableHTTP {
		err = scdmodels.ValidateUSSBaseURL(string(params.UssBaseUrl))
		if err != nil {
			return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	state := scdmodels.OperationalIntentState(params.State)
	if !state.IsValidInDSS() {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid OperationalIntent state: %s", params.State)
	}

	for idx, extent := range params.Extents {
		cExtent, err := dssmodels.Volume4DFromSCDRest(&extent)
		if err != nil {
			return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extent %d", idx)
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to union extents")
	}

	if uExtent.StartTime == nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_end from extents")
	}

	if time.Now().After(*uExtent.EndTime) {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "OperationalIntents may not end in the past")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	if uExtent.EndTime.Before(*uExtent.StartTime) {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is past the start time")
	}

	if ovn == "" && params.State != restapi.OperationalIntentState_Accepted {
		return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid state for initial version: `%s`", params.State)
	}

	subscriptionID := dssmodels.ID("")
	if params.SubscriptionId != nil {
		subscriptionID, err = dssmodels.IDFromOptionalString(string(*params.SubscriptionId))
		if err != nil {
			return nil, nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format for Subscription ID: `%s`", *params.SubscriptionId)
		}
	}

	var responseOK *restapi.ChangeOperationalIntentReferenceResponse
	var responseConflict *restapi.AirspaceConflictResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		var version int32 // Version of the Operational Intent (0 means creation requested).

		// Get existing OperationalIntent, if any, and validate request
		old, err := r.GetOperationalIntent(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get OperationalIntent from repo")
		}
		if old != nil {
			if old.Manager != dssmodels.Manager(manager) {
				return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
					"OperationalIntent owned by %s, but %s attempted to modify", old.Manager, manager)
			}
			if old.OVN != scdmodels.OVN(ovn) {
				return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
					"Current version is %s but client specified version %s", old.OVN, ovn)
			}

			version = int32(old.Version)
		} else {
			if ovn != "" {
				return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent does not exist and therefore is not version %s", ovn)
			}

			version = 0
		}

		var sub *scdmodels.Subscription
		if subscriptionID.Empty() {
			// Create implicit Subscription
			if params.NewSubscription == nil || params.NewSubscription.UssBaseUrl == "" {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing new_subscription or uss_base_url in new_subscription")
			}
			if !a.EnableHTTP {
				err := scdmodels.ValidateUSSBaseURL(string(params.NewSubscription.UssBaseUrl))
				if err != nil {
					return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate USS base URL")
				}
			}

			subToUpsert := scdmodels.Subscription{
				ID:                          dssmodels.ID(uuid.New().String()),
				Manager:                     dssmodels.Manager(manager),
				StartTime:                   uExtent.StartTime,
				EndTime:                     uExtent.EndTime,
				AltitudeLo:                  uExtent.SpatialVolume.AltitudeLo,
				AltitudeHi:                  uExtent.SpatialVolume.AltitudeHi,
				Cells:                       cells,
				USSBaseURL:                  string(params.NewSubscription.UssBaseUrl),
				NotifyForOperationalIntents: true,
				ImplicitSubscription:        true,
			}
			if params.NewSubscription.NotifyForConstraints != nil {
				subToUpsert.NotifyForConstraints = *params.NewSubscription.NotifyForConstraints
			}

			sub, err = r.UpsertSubscription(ctx, &subToUpsert)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to create implicit subscription")
			}

		} else {
			// Use existing Subscription
			sub, err = r.GetSubscription(ctx, subscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get Subscription")
			}
			if sub == nil {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription %s does not exist", subscriptionID)
			}
			if sub.Manager != dssmodels.Manager(manager) {
				return stacktrace.Propagate(
					stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Specificed Subscription is owned by different client"),
					"Subscription %s owned by %s, but %s attempted to use it for an OperationalIntent", subscriptionID, sub.Manager, manager)
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
				}
			}
			if !sub.Cells.Contains(cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, cells)
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the OperationalIntent")
				}
			}
			if updateSub {
				sub, err = r.UpsertSubscription(ctx, sub)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to update existing Subscription")
				}
			}
		}

		if state.RequiresKey() {
			// Construct a hash set of OVNs as the key
			key := map[scdmodels.OVN]bool{}
			if params.Key != nil {
				for _, ovn := range *params.Key {
					key[scdmodels.OVN(ovn)] = true
				}
			}

			// Identify OperationalIntents missing from the key
			var missingOps []*scdmodels.OperationalIntent
			relevantOps, err := r.SearchOperationalIntents(ctx, uExtent)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to SearchOperations")
			}
			for _, relevantOp := range relevantOps {
				_, ok := key[relevantOp.OVN]
				if !ok && relevantOp.RequiresKey() {
					if relevantOp.Manager != dssmodels.Manager(manager) {
						relevantOp.OVN = scdmodels.NoOvnPhrase
					}
					missingOps = append(missingOps, relevantOp)
				}
			}

			// Identify Constraints missing from the key
			var missingConstraints []*scdmodels.Constraint
			if sub.NotifyForConstraints {
				constraints, err := r.SearchConstraints(ctx, uExtent)
				if err != nil {
					return stacktrace.Propagate(err, "Unable to SearchConstraints")
				}
				for _, relevantConstraint := range constraints {
					if _, ok := key[relevantConstraint.OVN]; !ok {
						if relevantConstraint.Manager != dssmodels.Manager(manager) {
							relevantConstraint.OVN = scdmodels.NoOvnPhrase
						}
						missingConstraints = append(missingConstraints, relevantConstraint)
					}
				}
			}

			// If the client is missing some OVNs, provide the pointers to the
			// information they need
			if len(missingOps) > 0 || len(missingConstraints) > 0 {
				msg := "Current OVNs not provided for one or more OperationalIntents or Constraints"
				responseConflict = &restapi.AirspaceConflictResponse{Message: &msg}

				if len(missingOps) > 0 {
					responseConflict.MissingOperationalIntents = new([]restapi.OperationalIntentReference)
					for _, missingOp := range missingOps {
						*responseConflict.MissingOperationalIntents = append(*responseConflict.MissingOperationalIntents, *missingOp.ToRest())
					}
				}

				if len(missingConstraints) > 0 {
					responseConflict.MissingConstraints = new([]restapi.ConstraintReference)
					for _, missingConstraint := range missingConstraints {
						*responseConflict.MissingConstraints = append(*responseConflict.MissingConstraints, *missingConstraint.ToRest())
					}
				}

				return stacktrace.NewErrorWithCode(dsserr.MissingOVNs, "Missing OVNs: %v", msg)
			}
		}

		// Construct the new OperationalIntent
		op := &scdmodels.OperationalIntent{
			ID:      id,
			Manager: dssmodels.Manager(manager),
			Version: scdmodels.VersionNumber(version + 1),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
			Cells:         cells,

			USSBaseURL:     string(params.UssBaseUrl),
			SubscriptionID: sub.ID,
			State:          state,
		}
		err = op.ValidateTimeRange()
		if err != nil {
			return stacktrace.Propagate(err, "Error validating time range")
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

		// Upsert the OperationalIntent
		op, err = r.UpsertOperationalIntent(ctx, op)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to upsert OperationalIntent in repo")
		}

		// Find Subscriptions that may need to be notified
		allsubs, err := r.SearchSubscriptions(ctx, notifyVol4)
		if err != nil {
			return err
		}

		// Limit Subscription notifications to only those interested in OperationalIntents
		subs := repos.Subscriptions{}
		for _, sub := range allsubs {
			if sub.NotifyForOperationalIntents {
				subs = append(subs, sub)
			}
		}

		// Increment notification indices for relevant Subscriptions
		err = subs.IncrementNotificationIndices(ctx, r)
		if err != nil {
			return err
		}

		// Return response to client
		responseOK = &restapi.ChangeOperationalIntentReferenceResponse{
			OperationalIntentReference: *op.ToRest(),
			Subscribers:                makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, responseConflict, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return responseOK, responseConflict, nil
}
