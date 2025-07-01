package aux

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetDSSInstances(ctx context.Context, req *restapi.GetDSSInstancesRequest) restapi.GetDSSInstancesResponseSet {

	resp := restapi.GetDSSInstancesResponseSet{}

	if req.Auth.Error != nil {
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	repo, err := a.Store.Interact(ctx)
	if err != nil {
		resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to interact with the store"))}
		return resp
	}

	metadata, err := repo.GetDSSMetadata(ctx)

	if err != nil {
		switch stacktrace.GetCode(err) {
		case dsserr.NotImplemented:
			resp.Response501 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Operation not implemented"))}
		default:
			resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not retrieve DAR information"))}
		}

		return resp
	}

	instances := make([]restapi.DSSInstance, len(metadata))

	for index, instanceMetadata := range metadata {

		instances[index] = restapi.DSSInstance{
			Id:             instanceMetadata.Locality,
			PublicEndpoint: &instanceMetadata.PublicEndpoint,
		}

		if instanceMetadata.LatestTimestamp.Source.Valid {

			instances[index].MostRecentHeartbeat = &restapi.Heartbeat{
				Timestamp: instanceMetadata.LatestTimestamp.Timestamp.Format(time.RFC3339Nano),
				Reporter:  &instanceMetadata.LatestTimestamp.Reporter.String,
				Source:    instanceMetadata.LatestTimestamp.Source.String,
			}

			if instanceMetadata.LatestTimestamp.NextHeartbeatExpectedBefore != nil {
				nextExpectedTimestamp := instanceMetadata.LatestTimestamp.NextHeartbeatExpectedBefore.Format(time.RFC3339Nano)
				instances[index].MostRecentHeartbeat.NextHeartbeatExpectedBefore = &nextExpectedTimestamp
			}

		}

	}

	resp.Response200 = &restapi.DSSInstancesResponse{DssInstances: &instances}

	return resp

}

func (a *Server) PutDSSInstancesHeartbeat(ctx context.Context, req *restapi.PutDSSInstancesHeartbeatRequest) restapi.PutDSSInstancesHeartbeatResponseSet {

	resp := restapi.PutDSSInstancesHeartbeatResponseSet{}

	if req.Auth.Error != nil {
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	repo, err := a.Store.Interact(ctx)
	if err != nil {
		resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to interact with the store"))}
		return resp
	}

	heartbeat := models.Heartbeat{
		Source:   *req.Source,
		Reporter: *req.Auth.ClientID,
		Locality: a.Locality,
	}

	if req.Timestamp != nil {
		ts, err := time.Parse(time.RFC3339Nano, *req.Timestamp)
		if err != nil {
			resp.Response400 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to parse timestamp as RFC3339 time"))}
			return resp
		}
		heartbeat.Timestamp = &ts
	}

	if req.NextHeartbeatExpectedBefore != nil {
		ts, err := time.Parse(time.RFC3339Nano, *req.NextHeartbeatExpectedBefore)
		if err != nil {
			resp.Response400 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to parse next heartbeat expected before as RFC3339 time"))}
			return resp
		}
		heartbeat.NextHeartbeatExpectedBefore = &ts
	}

	err = repo.RecordHeartbeat(ctx, heartbeat)

	if err != nil {
		resp.Response400 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to record heartbeat"))}
		return resp
	}

	// Return the same response as the get one

	getResponse := a.GetDSSInstances(ctx, &restapi.GetDSSInstancesRequest{})

	return restapi.PutDSSInstancesHeartbeatResponseSet{
		Response201: getResponse.Response200,
		Response401: getResponse.Response401,
		Response403: getResponse.Response403,
		Response500: getResponse.Response500,
		Response501: getResponse.Response501,
	}
}
