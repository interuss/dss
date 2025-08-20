package scd

import (
	"context"
	"strings"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

func GetDefaultAvailabilityResponse(id dssmodels.Manager) *restapi.UssAvailabilityStatusResponse {
	return &restapi.UssAvailabilityStatusResponse{
		Status: restapi.UssAvailabilityStatus{
			Availability: restapi.UssAvailabilityState_Unknown,
			Uss:          id.String()},
		Version: "",
	}
}

func (a *Server) GetUssAvailability(ctx context.Context, req *restapi.GetUssAvailabilityRequest,
) restapi.GetUssAvailabilityResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetUssAvailabilityResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id := dssmodels.ManagerFromString(req.UssId)
	if id == "" {
		return restapi.GetUssAvailabilityResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "UssId not provided"))}}
	}

	var response *restapi.UssAvailabilityStatusResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get USS availability from Store
		ussa, err := r.GetUssAvailability(ctx, id)
		if err != nil && err != pgx.ErrNoRows {
			return stacktrace.Propagate(err, "Could not get USS availability from repo")
		}
		if ussa == nil {
			// Return default availability status "Unknown"
			response = GetDefaultAvailabilityResponse(id)
			return nil
		}
		response = &restapi.UssAvailabilityStatusResponse{
			Status:  *ussa.ToRest(),
			Version: ussa.Version.String(),
		}
		return nil
	}

	err := a.Store.Transact(ctx, action)
	if err != nil {
		// In case of older DB versions where availability table doesn't exist
		if strings.Contains(err.Error(), "does not exist") {
			response = GetDefaultAvailabilityResponse(id)
		} else {
			// No need to Propagate this error as this is not a useful stacktrace line
			return restapi.GetUssAvailabilityResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, err)}}
		}
	}
	return restapi.GetUssAvailabilityResponseSet{Response200: response}
}

func (a *Server) SetUssAvailability(ctx context.Context, req *restapi.SetUssAvailabilityRequest,
) restapi.SetUssAvailabilityResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.SetUssAvailabilityResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

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
	ussareq := &scdmodels.UssAvailabilityStatus{
		Uss:          id,
		Availability: availability,
	}

	var result *restapi.UssAvailabilityStatusResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		old, err := r.GetUssAvailability(ctx, id)
		if err != nil && err != pgx.ErrNoRows {
			return stacktrace.Propagate(err, "Could not get USS availability from repo")
		}
		switch {
		case old == nil && !version.Empty():
			// The user wants set a new availability status but it already exists.
			return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "availability for USS %s already exists", id.String())
		case old != nil && old.Version != version:
			// The user wants to update an availability status but the version doesn't match.
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "USS availability version %s is not current", version),
				"Current version is %s but client specified version %s", old.Version, version)
		}

		// Upsert the USS availability
		ussa, err := r.UpsertUssAvailability(ctx, ussareq)
		if err != nil {
			return stacktrace.Propagate(err, "Could not upsert USS Availability into repo")
		}
		if ussa == nil {
			return stacktrace.NewError("UpsertUssAvailability returned no USS availability for ID: %s", id)
		}
		result = &restapi.UssAvailabilityStatusResponse{
			Status:  *ussa.ToRest(),
			Version: ussa.Version.String(),
		}
		return nil
	}
	err = a.Store.Transact(ctx, action)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			// In case of older DB versions where availability table doesn't exist
			result = GetDefaultAvailabilityResponse(id)
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
	}

	// Return response to client
	return restapi.SetUssAvailabilityResponseSet{Response200: result}
}
