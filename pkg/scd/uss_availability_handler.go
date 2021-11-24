package scd

import (
	"context"
	"log"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetUssAvailability(ctx context.Context, request *scdpb.GetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	id := request.GetUssId()
	if id == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", request.GetUssId())
	}

	var response *scdpb.UssAvailabilityStatusResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Get USS availability from Store
		ussa, err := r.GetUssAvailability(ctx, id)
		// log.Println(".. ussa.Availability: ", ussa.Availability)
		if err != nil {
			return stacktrace.Propagate(err, "Could not get USS availability from repo")
		}
		if ussa == nil {
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "USS ID %s not found", id)
		}
		response = &scdpb.UssAvailabilityStatusResponse{
			Status: &scdpb.UssAvailabilityStatus{
				Availability: ussa.Availability.String(),
				Uss: id},
		}
		return nil
	}

	err := a.Store.Transact(ctx, action)
	if err != nil {
		log.Println("")
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}
	return response, nil
}

func (a *Server) SetUssAvailability(ctx context.Context, request *scdpb.SetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	log.Println("implement me")
	return nil, nil
}
