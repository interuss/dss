package versioning

import (
	"context"

	versioning "github.com/interuss/dss/pkg/api/versioningv1"
	"github.com/interuss/dss/pkg/version"
)

type Server struct {
}

func (s *Server) GetVersion(ctx context.Context, req *versioning.GetVersionRequest) versioning.GetVersionResponseSet {
	// The DSS has no notion of particular system identities: whatever the request, we will
	// always return the current version of the DSS binary.
	versionStr := version.Current().String()
	return versioning.GetVersionResponseSet{
		Response200: &versioning.GetVersionResponse{
			SystemIdentity: &req.SystemIdentity,
			SystemVersion:  (*versioning.VersionIdentifier)(&versionStr),
		},
	}
}
