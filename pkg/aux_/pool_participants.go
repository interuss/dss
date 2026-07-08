package aux

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/dss/pkg/aux_/repos"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetDSSInstances(ctx context.Context, req *restapi.GetDSSInstancesRequest) restapi.GetDSSInstancesResponseSet {
	resp := restapi.GetDSSInstancesResponseSet{}

	instancesResponse, err := dssstore.TransactWithResult[repos.Repository, *restapi.DSSInstancesResponse](ctx, a.Store, req)
	if err != nil {
		switch stacktrace.GetCode(err) {
		case dsserr.NotImplemented:
			resp.Response501 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Operation not implemented"))}
		default:
			resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not retrieve DAR information"))}
		}
		return resp
	}

	resp.Response200 = instancesResponse
	return resp
}

func (a *Server) PutDSSInstancesHeartbeat(ctx context.Context, req *restapi.PutDSSInstancesHeartbeatRequest) restapi.PutDSSInstancesHeartbeatResponseSet {
	_, err := a.Store.Transact(ctx, req)
	if err != nil {
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.PutDSSInstancesHeartbeatResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to record heartbeat"))}}
		default:
			return restapi.PutDSSInstancesHeartbeatResponseSet{Response500: &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to record heartbeat"))}}
		}
	}

	// Return the same response as the get one
	getResponse := a.GetDSSInstances(ctx, &restapi.GetDSSInstancesRequest{Auth: req.Auth})

	return restapi.PutDSSInstancesHeartbeatResponseSet{
		Response201: getResponse.Response200,
		Response401: getResponse.Response401,
		Response403: getResponse.Response403,
		Response500: getResponse.Response500,
		Response501: getResponse.Response501,
	}
}
