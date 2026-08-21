package scd

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *restapi.DeleteConstraintReferenceRequest,
) restapi.DeleteConstraintReferenceResponseSet {

	// Retrieve Constraint ID
	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.DeleteConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	// Retrieve ID of client making call
	if req.Auth.ClientID == nil {
		return restapi.DeleteConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	// Retrieve OVN
	ovn := scdmodels.OVN(req.Ovn)
	if ovn == "" {
		return restapi.DeleteConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing OVN for constraint to modify"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.ChangeConstraintReferenceResponse](ctx, a.Store, req)
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
		case dsserr.VersionMismatch:
			return restapi.DeleteConstraintReferenceResponseSet{Response409: errResp}
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

	_, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return restapi.GetConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid))}}
	}

	if req.Auth.ClientID == nil {
		return restapi.GetConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.GetConstraintReferenceResponse](ctx, a.Store, req)
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

	if req.BodyParseError != nil {
		return restapi.CreateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.CreateConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	err := validateConstraintUpsertRequest(req.Entityid, req.Body.UssBaseUrl, a.AllowHTTPBaseUrls)
	if err != nil {
		return restapi.CreateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Constraint upsert parameters"))}}
	}

	res, err := dssstore.TransactWithResult[repos.Repository, *restapi.ChangeConstraintReferenceResponse](ctx, a.Store, req)
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

	if req.BodyParseError != nil {
		return restapi.UpdateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Auth.ClientID == nil {
		return restapi.UpdateConstraintReferenceResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing manager"))}}
	}

	err := validateConstraintUpsertRequest(req.Entityid, req.Body.UssBaseUrl, a.AllowHTTPBaseUrls)
	if err != nil {
		return restapi.UpdateConstraintReferenceResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate Constraint upsert parameters"))}}
	}

	res, err := dssstore.TransactWithResult[repos.Repository, *restapi.ChangeConstraintReferenceResponse](ctx, a.Store, req)
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

// validateConstraintUpsertRequest performs handler-side only validation of Constraint upsert requests.
// Note that this does NOT check for anything related to access controls: any error returned should be labeled as a dsserr.BadRequest.
func validateConstraintUpsertRequest(entityid restapi.EntityID, ussBaseUrl restapi.ConstraintUssBaseURL, allowHTTPBaseUrls bool) error {
	_, err := dssmodels.IDFromString(string(entityid))
	if err != nil {
		return stacktrace.NewError("Invalid ID format: `%s`", entityid)
	}

	if len(ussBaseUrl) == 0 {
		return stacktrace.NewError("Missing required UssBaseUrl")
	}

	if !allowHTTPBaseUrls {
		err := scdmodels.ValidateUSSBaseURL(string(ussBaseUrl))
		if err != nil {
			return stacktrace.Propagate(err, "Failed to validate base URL")
		}
	}

	return nil
}

// QueryConstraintReferences queries existing contraint refs in the given
// bounds.
func (a *Server) QueryConstraintReferences(ctx context.Context, req *restapi.QueryConstraintReferencesRequest,
) restapi.QueryConstraintReferencesResponseSet {

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
	_, err := scdmodels.Volume4DFromSCDRest(aoi)
	if err != nil {
		return restapi.QueryConstraintReferencesResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to convert to internal geometry model"))}}
	}

	response, err := dssstore.TransactWithResult[repos.Repository, *restapi.QueryConstraintReferencesResponse](ctx, a.Store, req)
	if err != nil {
		return restapi.QueryConstraintReferencesResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
	}

	return restapi.QueryConstraintReferencesResponseSet{Response200: response}
}
