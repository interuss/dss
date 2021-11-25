package scd

import (
	"context"
	"database/sql"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetUssAvailability(ctx context.Context, request *scdpb.GetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	id := request.GetUssId()
	if id == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "UssId not provided")
	}

	var response *scdpb.UssAvailabilityStatusResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get USS availability from Store
		ussa, err := r.GetUssAvailability(ctx, id)
		if err != nil && err != sql.ErrNoRows {
			return stacktrace.Propagate(err, "Could not get USS availability from repo")
		}
		if ussa == nil {
			// Return default availability status "Unknown"
			response = &scdpb.UssAvailabilityStatusResponse{
				Status: &scdpb.UssAvailabilityStatus{
					Availability: "Unknown",
					Uss:          id},
			}
			return nil
		}
		response = &scdpb.UssAvailabilityStatusResponse{
			Status: &scdpb.UssAvailabilityStatus{
				Availability: ussa.Availability.String(),
				Uss:          id},
		}
		return nil
	}

	err := a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}
	return response, nil
}

func (a *Server) SetUssAvailability(ctx context.Context, request *scdpb.SetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	return a.PutUssAvailability(ctx, request.GetUssId(), "", request.GetParams())
}

func (a *Server) PutUssAvailability(ctx context.Context, ussID string, version string, params *scdpb.SetUssAvailabilityStatusParameters) (*scdpb.UssAvailabilityStatusResponse, error) {
	if ussID == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "ussID not provided.")
	}
	// Retrieve USS availability status from request params
	availability := params.GetAvailability()
	if availability == "" {
		// Set availability default to Unknown
		availability = "Unknown"
	}
	ussareq := &scdmodels.UssAvailabilityStatus{
		Uss:          ussID,
		Availability: scdmodels.UssAvailabilityState(availability),
	}

	var result *scdpb.UssAvailabilityStatusResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		ussa, err := r.UpsertUssAvailability(ctx, ussareq)
		if err != nil {
			return stacktrace.Propagate(err, "Could not upsert UssAvailability into repo")
		}
		if ussa == nil {
			return stacktrace.NewError("UssAvailability returned no Uss for ID: %s", ussID)
		}
		result = &scdpb.UssAvailabilityStatusResponse{
			Status: &scdpb.UssAvailabilityStatus{
				Availability: ussa.Availability.String(),
				Uss:          ussID},
		}
		return nil
	}
	err := a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	// Return response to client
	return result, nil
}
