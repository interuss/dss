// This file is auto-generated; do not change as any changes will be overwritten
package dummyoauth

type TokenResponse struct {
	// JWT that may be used as a Bearer token to authorize operations on an appropriately-configured DSS instance
	AccessToken string `json:"access_token"`
}

type BadRequestResponse struct {
	// Human-readable message describing problem with request
	Message *string `json:"message"`
}
