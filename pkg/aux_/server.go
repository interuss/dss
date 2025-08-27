package aux

import (
	"context"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/auxv1"
	auxstore "github.com/interuss/dss/pkg/aux_/store"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/interuss/dss/pkg/version"
	"github.com/interuss/stacktrace"
)

// Server implements auxv1.Implementation.
type Server struct {
	Store    auxstore.Store
	Locality string
}

func setAuthError(ctx context.Context, authErr error, resp401, resp403 **restapi.ErrorResponse, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}
	case dsserr.PermissionDenied:
		*resp403 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}
	default:

		if authErr == nil {
			authErr = stacktrace.NewError("Unknown error")
		}

		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}

// GetVersion returns information about the version of the server.
func (a *Server) GetVersion(context.Context, *restapi.GetVersionRequest) restapi.GetVersionResponseSet {
	return restapi.GetVersionResponseSet{Response200: &restapi.VersionResponse{
		Version: version.Current().String()}}
}

// ValidateOauth will exercise validating the Oauth token
func (a *Server) ValidateOauth(ctx context.Context, req *restapi.ValidateOauthRequest) restapi.ValidateOauthResponseSet {

	if req.Auth.Error != nil {
		resp := restapi.ValidateOauthResponseSet{}
		switch stacktrace.GetCode(req.Auth.Error) {
		case dsserr.Unauthenticated:
			resp.Response401 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(req.Auth.Error, "Authentication failed"))}
		case dsserr.PermissionDenied:
			resp.Response403 = &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(req.Auth.Error, "Authorization failed"))}
		default:
			resp.Response500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(req.Auth.Error, "Could not perform authorization"))}
		}
		return resp
	}

	if req.Auth.ClientID == nil {
		return restapi.ValidateOauthResponseSet{Response403: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if req.Owner != nil && *req.Owner != *req.Auth.ClientID {
		return restapi.ValidateOauthResponseSet{Response403: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Owner mismatch, required: %s, but oauth token has %s", *req.Owner, *req.Auth.ClientID))}}
	}
	return restapi.ValidateOauthResponseSet{Response200: &api.EmptyResponseBody{}}
}
