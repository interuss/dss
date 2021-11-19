package scd

import "log"
import "context"
import "github.com/interuss/dss/pkg/api/v1/scdpb"

func (a *Server) GetUssAvailability(ctx context.Context, request *scdpb.GetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	log.Println("implement me")
	return nil, nil
}

func (a *Server) SetUssAvailability(ctx context.Context, request *scdpb.SetUssAvailabilityRequest) (*scdpb.UssAvailabilityStatusResponse, error) {
	log.Println("implement me")
	return nil, nil
}
