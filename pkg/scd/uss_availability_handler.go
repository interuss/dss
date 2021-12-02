package scd

import "context"
import "github.com/interuss/dss/pkg/api/v1/scdpb"
import "github.com/interuss/stacktrace"

func (a *Server) GetUssAvailability(ctx context.Context, request *scdpb.GetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	return nil, stacktrace.NewError("Not implemented")
}

func (a *Server) SetUssAvailability(ctx context.Context, request *scdpb.SetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	return nil, stacktrace.NewError("Not implemented")
}
