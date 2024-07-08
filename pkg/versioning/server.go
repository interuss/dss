package versioning

import (
	"context"
	"github.com/interuss/dss/pkg/api"
	versioning "github.com/interuss/dss/pkg/api/versioningv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

type Server struct {
	systemIdentity    *versioning.SystemBoundaryIdentifier
	versionIdentifier *versioning.VersionIdentifier
}

func NewServer(systemIdentity string, versionIdentifier string) *Server {
	return &Server{
		systemIdentity:    (*versioning.SystemBoundaryIdentifier)(&systemIdentity),
		versionIdentifier: (*versioning.VersionIdentifier)(&versionIdentifier),
	}
}

func (s *Server) GetVersion(ctx context.Context, req *versioning.GetVersionRequest) versioning.GetVersionResponseSet {
	// This should take care of unauthenticated requests as well as
	// any request without the proper scope.
	if req.Auth.Error != nil {
		resp := versioning.GetVersionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	// TODO initial assumption is that the DSS reports the version for itself _only_:
	//  confirm this is the correct approach. Alternatively we could configure a map of identities to versions
	//  and provide versions for multiple identities.
	if *s.systemIdentity != req.SystemIdentity {
		return versioning.GetVersionResponseSet{
			Response404: &api.EmptyResponseBody{},
		}
	}

	return versioning.GetVersionResponseSet{
		Response200: &versioning.GetVersionResponse{
			SystemIdentity: s.systemIdentity,
			SystemVersion:  s.versionIdentifier,
		},
	}
}

func setAuthError(ctx context.Context, authErr error, resp401, resp403 **api.EmptyResponseBody, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &api.EmptyResponseBody{}
	case dsserr.PermissionDenied:
		*resp403 = &api.EmptyResponseBody{}
	default:
		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}
