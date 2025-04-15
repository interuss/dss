package aux

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetDSSInstances(ctx context.Context, req *restapi.GetDSSInstancesRequest) restapi.GetDSSInstancesResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetDSSInstancesResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// Pool participant storage is not yet implemented.
	return restapi.GetDSSInstancesResponseSet{Response200: &restapi.DSSInstancesResponse{DssInstances: &[]restapi.DSSInstance{}}}
}
