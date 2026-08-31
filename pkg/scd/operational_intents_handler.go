package scd

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// DeleteOperationalIntentReference deletes a single operational intent ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationalIntentReference(ctx context.Context, req *restapi.DeleteOperationalIntentReferenceRequest,
) restapi.DeleteOperationalIntentReferenceResponseSet {

	// Retrieve OperationalIntent ID
	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	// Retrieve OVN
	ovn := scdmodels.OVN(req.Ovn)
	if ovn == "" {
		return restapi.DeleteOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing OVN for operational intent to modify"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.ChangeOperationalIntentReferenceResponse](ctx, a.Store, req)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete operational intent")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response403: errResp}
		case dsserr.NotFound:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response404: errResp}
		case dsserr.VersionMismatch:
			return restapi.DeleteOperationalIntentReferenceResponseSet{Response409: errResp}
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

	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetOperationalIntentReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.GetOperationalIntentReferenceResponse](ctx, a.Store, req)
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
	_, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry"))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.QueryOperationalIntentReferencesResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.QueryOperationalIntentReferenceResponse](ctx, a.Store, req)
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

	if req.BodyParseError != nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if _, err := operations.ValidateAndReturnOIRUpsertParams(timestamp.MustFromContext(ctx), req.Entityid, "", req.Body, a.AllowHTTPBaseUrls); err != nil {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Operational Intent Reference upsert parameters"))}}
	}

	// We pass any as the result type here because we might get a ChangeOperationalIntentReferenceResponse or an AirspaceConflictResponse back.
	result, err := dssstore.TransactWithResult[repos.Repository, any](ctx, a.Store, req)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not put Operational Intent Reference")
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
			respConflict, ok := result.(*restapi.AirspaceConflictResponse)
			if !ok {
				return restapi.CreateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
					ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("unexpected result type %T for MissingOVNs conflict response", result))}}
			}
			return restapi.CreateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.CreateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	respOK, ok := result.(*restapi.ChangeOperationalIntentReferenceResponse)
	if !ok {
		return restapi.CreateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("unexpected result type %T for Operational Intent Reference upsert response", result))}}
	}
	return restapi.CreateOperationalIntentReferenceResponseSet{Response201: respOK}
}

func (a *Server) UpdateOperationalIntentReference(ctx context.Context, req *restapi.UpdateOperationalIntentReferenceRequest,
) restapi.UpdateOperationalIntentReferenceResponseSet {

	if req.BodyParseError != nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if _, err := operations.ValidateAndReturnOIRUpsertParams(timestamp.MustFromContext(ctx), req.Entityid, req.Ovn, req.Body, a.AllowHTTPBaseUrls); err != nil {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Operational Intent Reference upsert parameters"))}}
	}

	result, err := dssstore.TransactWithResult[repos.Repository, any](ctx, a.Store, req)
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
			respConflict, ok := result.(*restapi.AirspaceConflictResponse)
			if !ok {
				return restapi.UpdateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
					ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("unexpected result type %T for MissingOVNs conflict response", result))}}
			}
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response409: respConflict}
		default:
			return restapi.UpdateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	respOK, ok := result.(*restapi.ChangeOperationalIntentReferenceResponse)
	if !ok {
		return restapi.UpdateOperationalIntentReferenceResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("unexpected result type %T for Operational Intent Reference upsert response", result))}}
	}
	return restapi.UpdateOperationalIntentReferenceResponseSet{Response200: respOK}
}
