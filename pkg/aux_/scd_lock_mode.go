package aux

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
)

func (a *Server) GetScdLockMode(ctx context.Context, req *restapi.GetScdLockModeRequest) restapi.GetScdLockModeResponseSet {
	return restapi.GetScdLockModeResponseSet{Response200: &restapi.SCDLockModeResponse{GlobalLock: &a.ScdGlobalLock}}
}
