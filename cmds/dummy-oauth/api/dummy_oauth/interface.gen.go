// This file is auto-generated; do not change as any changes will be overwritten
package dummy_oauth

import (
	"context"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
)

var (
	GetTokenSecurity = map[string]api.SecurityScheme{}
)

type GetTokenRequest struct {
	// Fully-qualified domain name where the service for which this access token will be used is hosted.  The `aud` claim will be populated with this value.
	IntendedAudience *string

	// Scope or scopes that should be granted in the access token.  Multiple scopes can be delimited by spaces (%20) in a single value.  The `scope` claim will be populated with all requested scopes.
	Scope *string

	// Identity of the issuer of the token.  The `iss` claim will be populated with this value.
	Issuer *string

	// Unix timestamp (seconds since epoch) of the time this access token should expire.  If not specified, defaults to an hour from time of token creation.
	Expire *int64

	// Identity of client/subscriber requesting access token.  The `sub` claim will be populated with this value.
	Sub *string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetTokenResponseSet struct {
	// The requested token was generated successfully
	Response200 *TokenResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Generate an access token
	GetToken(ctx context.Context, req *GetTokenRequest) GetTokenResponseSet
}
