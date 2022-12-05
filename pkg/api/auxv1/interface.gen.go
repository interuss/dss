// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	GetVersionSecurity    = []api.AuthorizationOption{}
	ValidateOauthSecurity = []api.AuthorizationOption{
		{
			"Auth": {"dss.read.identification_service_areas"},
		},
		{
			"Auth": {"dss.write.identification_service_areas"},
		},
	}
)

type GetVersionRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetVersionResponseSet struct {
	// The version of the DSS is successfully returned.
	Response200 *VersionResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type ValidateOauthRequest struct {
	// Validate the owner claim matches the provided owner.
	Owner *string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type ValidateOauthResponseSet struct {
	// The provided token was validated.
	Response200 *api.EmptyResponseBody

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Queries the version of the DSS.
	GetVersion(ctx context.Context, req *GetVersionRequest) GetVersionResponseSet

	// Validate Oauth token against the DSS.
	ValidateOauth(ctx context.Context, req *ValidateOauthRequest) ValidateOauthResponseSet
}
