package aux

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/stacktrace"
)

func (a *Server) GetScdLockMode(ctx context.Context, req *restapi.GetScdLockModeRequest) restapi.GetScdLockModeResponseSet {

	resp := restapi.GetScdLockModeResponseSet{}

	if req.Auth.Error != nil {
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	return restapi.GetScdLockModeResponseSet{Response200: &restapi.SCDLockModeResponse{GlobalLock: &a.ScdGlobalLock}}
}
