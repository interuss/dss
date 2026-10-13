// This file is auto-generated; do not change as any changes will be overwritten
package dummyoauth

type TokenResponse struct {
	// JWT that may be used as a Bearer token to authorize operations on an appropriately-configured DSS instance.  Its `kid` header identifies the key in the JSON Web Key Set at `/.well-known/jwks.json` that verifies it.
	AccessToken string `json:"access_token"`
}

type JWKSResponse struct {
	// Public keys that verify access tokens issued by this provider.
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	// Key type.
	Kty string `json:"kty"`

	// Intended use of the key.
	Use string `json:"use"`

	// Signing algorithm used with this key.
	Alg string `json:"alg"`

	// SHA-256 JWK thumbprint of this key.
	Kid string `json:"kid"`

	// RSA modulus.
	N string `json:"n"`

	// RSA public exponent.
	E string `json:"e"`
}

type BadRequestResponse struct {
	// Human-readable message describing problem with request
	Message *string `json:"message,omitempty"`
}
