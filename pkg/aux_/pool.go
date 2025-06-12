package aux

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/auxv1"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/interuss/stacktrace"
)

func (a *Server) GetPool(ctx context.Context, req *restapi.GetPoolRequest) restapi.GetPoolResponseSet {
	resp := restapi.GetPoolResponseSet{}
	if req.Auth.Error != nil {
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	repo, err := a.Store.Interact(ctx)
	if err != nil {
		resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Unable to interact with the store"))}
		return resp
	}

	darID, err := repo.GetDSSAirspaceRepresentationID(ctx)

	if err == nil {
		resp.Response200 = &restapi.PoolResponse{DarId: &darID}
	} else {
		switch stacktrace.GetCode(err) {
		case dsserr.NotImplemented:
			resp.Response501 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Operation not implemented"))}
		default:
			resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not retrieve DAR information"))}
		}
	}
	return resp
}
