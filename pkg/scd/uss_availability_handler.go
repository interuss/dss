package scd

import (
	"context"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetUssAvailability(ctx context.Context, request *scdpb.GetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	id := request.GetUssId()
	if id == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "UssId not provided")
	}
	response := &scdpb.UssAvailabilityStatusResponse{
		Status: &scdpb.UssAvailabilityStatus{
			Availability: "Unknown",
			Uss:          id},
	}
	return response, nil
}

func (a *Server) SetUssAvailability(ctx context.Context, request *scdpb.SetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	id := request.GetUssId()
	if id == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "UssId not provided")
	}
	result := &scdpb.UssAvailabilityStatusResponse{
		Status: &scdpb.UssAvailabilityStatus{
			Availability: "Unknown",
			Uss:          id},
	}
	return result, nil
}
