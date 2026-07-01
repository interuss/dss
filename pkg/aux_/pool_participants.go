package aux

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/dss/pkg/aux_/actions"
	"github.com/interuss/dss/pkg/aux_/repos"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/locality"
	"github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetDSSInstances(ctx context.Context, req *restapi.GetDSSInstancesRequest) restapi.GetDSSInstancesResponseSet {

	resp := restapi.GetDSSInstancesResponseSet{}

	response, err := store.TransactWithResult[repos.Repository, *restapi.DSSInstancesResponse](ctx, a.Store, &store.ActionAdapter[repos.Repository, *restapi.GetDSSInstancesRequest]{
		Data: req,
		Run:  actions.GetDSSInstances,
	})

	if err != nil {
		switch stacktrace.GetCode(err) {
		case dsserr.NotImplemented:
			resp.Response501 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Operation not implemented"))}
		default:
			resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not retrieve DAR information"))}
		}

		return resp
	}

	resp.Response200 = response

	return resp

}

func (a *Server) PutDSSInstancesHeartbeat(ctx context.Context, req *restapi.PutDSSInstancesHeartbeatRequest) restapi.PutDSSInstancesHeartbeatResponseSet {
	ctx = locality.WithLocality(ctx, a.Locality)

	_, err := a.Store.Transact(ctx, &store.ActionAdapter[repos.Repository, *restapi.PutDSSInstancesHeartbeatRequest]{
		Data: req,
		Run:  actions.PutDSSInstancesHeartbeat,
	})
	if err != nil {
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.PutDSSInstancesHeartbeatResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Invalid heartbeat"))}}
		default:
			return restapi.PutDSSInstancesHeartbeatResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to record heartbeat"))}}
		}
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
