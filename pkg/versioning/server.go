package versioning

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	versioning "github.com/interuss/dss/pkg/api/versioningv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/version"
	"github.com/interuss/stacktrace"
)

type Server struct {
}

func (s *Server) GetVersion(ctx context.Context, req *versioning.GetVersionRequest) versioning.GetVersionResponseSet {
	// This should take care of unauthenticated requests as well as
	// any request without the proper scope.
	if req.Auth.Error != nil {
		resp := versioning.GetVersionResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

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

func setAuthError(ctx context.Context, authErr error, resp401, resp403 **api.EmptyResponseBody, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &api.EmptyResponseBody{}
	case dsserr.PermissionDenied:
		*resp403 = &api.EmptyResponseBody{}
	default:

		if authErr == nil {
			authErr = stacktrace.NewError("Unknown error")
		}
		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}
