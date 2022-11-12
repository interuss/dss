// This file is auto-generated; do not change as any changes will be overwritten
package dummyoauth

import (
	"context"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
)

var (
	GetTokenSecurity  = map[string]api.SecurityScheme{}
	PostFimsTokenSecurity = map[string]api.SecurityScheme{
    }
	GetFimsWellKnownOauthAuthorizationServerSecurity = map[string]api.SecurityScheme{}
	GetFimsWellKnownJwksJSONSecurity                 = map[string]api.SecurityScheme{}
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

	// The request was not properly formed
	Response400 *BadRequestResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type PostFimsTokenRequest struct {
	// The signature as defined in https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures-11#section-4.2
	XUtmMessageSignature *string

	// Defines what data is covered by the accompanying signature. Defined in https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures-11#section-4.1
	XUtmMessageSignatureInput *string

	// Contains information necessary to verify a JWS signature
	XUtmJwsHeader *xUtmMessageSignatureJoseHeader

	// SHA-512 hash of message contents as defined in:https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-digest-headers-10.
	ContentDigest *string

	// The data contained in the body of this request, if it parsed correctly
	Body *TokenRequestForm

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type PostFimsTokenResponseSet struct {
	// OK
	Response200 *HTTPTokenResponse

	// - Request did not conform to the API specification or failed validation.
	Response400 *HTTPErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetFimsWellKnownOauthAuthorizationServerRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetFimsWellKnownOauthAuthorizationServerResponseSet struct {
	// OK
	Response200 *Metadata

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetFimsWellKnownJwksJSONRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetFimsWellKnownJwksJSONResponseSet struct {
	// OK
	Response200 *JSONWebKeySet

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Generate an access token
	GetToken(ctx context.Context, req *GetTokenRequest) GetTokenResponseSet

	// Request an access token.
	// ---
	// The primary endpoint for this authorization server.  Used to request an access token
	// suitable for authorizing data exchanges within the USS Network.
	//
	// Implemented per https://tools.ietf.org/html/rfc6749#section-3.2 .
	//
	PostFimsToken(ctx context.Context, req *PostFimsTokenRequest) PostFimsTokenResponseSet

	// Provides metadata related to use of this authorization server
	// ---
	// Per RFC8414, this endpoint provides metadata related to use of this authorization
	// server. See https://tools.ietf.org/html/rfc8414#section-3 for more details.
	//
	GetFimsWellKnownOauthAuthorizationServer(ctx context.Context, req *GetFimsWellKnownOauthAuthorizationServerRequest) GetFimsWellKnownOauthAuthorizationServerResponseSet

	// Serves the public JWKS of the authorization server
	// ---
	// This endpoint serves the signing key(s) the client uses to validate
	// signatures from the authorization server.
	//
	// The JWK Set MAY also contain the server's encryption key or keys,
	// which are used by clients to encrypt requests to the server.
	//
	// When both signing and encryption keys are made available, a "use"
	// (public key use) parameter value is REQUIRED for all keys in the
	// referenced JWK Set to indicate each key's intended usage.
	//
	// Refer to RFC7517 - https://tools.ietf.org/html/rfc7517
	//
	GetFimsWellKnownJwksJSON(ctx context.Context, req *GetFimsWellKnownJwksJSONRequest) GetFimsWellKnownJwksJSONResponseSet
}
