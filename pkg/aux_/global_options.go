package aux

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
)

func (a *Server) GetGlobalOptions(ctx context.Context, req *restapi.GetGlobalOptionsRequest) restapi.GetGlobalOptionsResponseSet {

	return restapi.GetGlobalOptionsResponseSet{Response200: &restapi.GlobalOptionsResponse{ScdGlobalLock: &a.Options.GlobalLock, ScdHashLock: &a.Options.HashLock, TimeBasedNotificationIndex: &a.Options.TimeBasedNotificationIndex}}
}
