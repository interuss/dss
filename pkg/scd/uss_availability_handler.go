package scd

import (
	"context"
	"strings"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/scd/actions"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetUssAvailability(ctx context.Context, req *restapi.GetUssAvailabilityRequest,
) restapi.GetUssAvailabilityResponseSet {

	id := dssmodels.ManagerFromString(req.UssId)
	if id == "" {
		return restapi.GetUssAvailabilityResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "UssId not provided"))}}
	}

	var response *restapi.UssAvailabilityStatusResponse
	result, err := a.Store.Transact(ctx, &actions.GetUssAvailabilityAction{ID: id})
	if err != nil {
		// In case of older DB versions where availability table doesn't exist
		if strings.Contains(err.Error(), "does not exist") {
			response = actions.GetDefaultAvailabilityResponse(id)
		} else {
			// No need to Propagate this error as this is not a useful stacktrace line
			return restapi.GetUssAvailabilityResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, err)}}
		}
	} else {
		var ok bool
		response, ok = result.(*restapi.UssAvailabilityStatusResponse)
		if !ok || response == nil {
			return restapi.GetUssAvailabilityResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("Invalid result %T", result))}}
		}
	}
	return restapi.GetUssAvailabilityResponseSet{Response200: response}
}

func (a *Server) SetUssAvailability(ctx context.Context, req *restapi.SetUssAvailabilityRequest,
) restapi.SetUssAvailabilityResponseSet {

	if req.UssId == "" {
		return restapi.SetUssAvailabilityResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "ussID not provided."))}}
	}
	if req.BodyParseError != nil {
		return restapi.SetUssAvailabilityResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	// Retrieve USS availability status from request params
	availability, err := scdmodels.UssAvailabilityStateFromRest(req.Body.Availability)
	if err != nil {
		return restapi.SetUssAvailabilityResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid availability state"))}}
	}
	id := dssmodels.ManagerFromString(req.UssId)
	version := scdmodels.OVN(req.Body.OldVersion)

	var result *restapi.UssAvailabilityStatusResponse
	res, err := a.Store.Transact(ctx, &actions.SetUssAvailabilityAction{ID: id, Version: version, Availability: availability})
	if err != nil {
		// In case of older DB versions where availability table doesn't exist
		if strings.Contains(err.Error(), "does not exist") {
			result = actions.GetDefaultAvailabilityResponse(id)
		} else {
			err = stacktrace.Propagate(err, "Could not set USS availability status")
			errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
			switch stacktrace.GetCode(err) {
			case dsserr.AlreadyExists, dsserr.VersionMismatch:
				return restapi.SetUssAvailabilityResponseSet{Response409: errResp}
			default:
				return restapi.SetUssAvailabilityResponseSet{Response500: &api.InternalServerErrorBody{
					ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
			}
		}
	} else {
		var ok bool
		result, ok = res.(*restapi.UssAvailabilityStatusResponse)
		if !ok || result == nil {
			return restapi.SetUssAvailabilityResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.NewError("Invalid result %T", res))}}
		}
	}

	// Return response to client
	return restapi.SetUssAvailabilityResponseSet{Response200: result}
}
