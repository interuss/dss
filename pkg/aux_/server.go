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
	Store         auxstore.Store
	Locality      string
	ScdGlobalLock bool
}

// GetVersion returns information about the version of the server.
func (a *Server) GetVersion(context.Context, *restapi.GetVersionRequest) restapi.GetVersionResponseSet {
	return restapi.GetVersionResponseSet{Response200: &restapi.VersionResponse{
		Version: version.Current().String()}}
}

// ValidateOauth will exercise validating the Oauth token
func (a *Server) ValidateOauth(ctx context.Context, req *restapi.ValidateOauthRequest) restapi.ValidateOauthResponseSet {

	if req.Auth.ClientID == nil {
		return restapi.ValidateOauthResponseSet{Response403: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if req.Owner != nil && *req.Owner != *req.Auth.ClientID {
		return restapi.ValidateOauthResponseSet{Response403: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Owner mismatch, required: %s, but oauth token has %s", *req.Owner, *req.Auth.ClientID))}}
	}
	return restapi.ValidateOauthResponseSet{Response200: &api.EmptyResponseBody{}}
}
