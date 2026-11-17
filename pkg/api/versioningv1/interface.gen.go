// This file is auto-generated; do not change as any changes will be overwritten
package versioningv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	InterussVersioningReadSystemVersionsScope = api.RequiredScope("interuss.versioning.read_system_versions")
	GetVersionSecurity                        = []api.AuthorizationOption{
		{
			"Authority": {InterussVersioningReadSystemVersionsScope},
		},
	}
)

type GetVersionRequest struct {
	// The system identity/boundary for which a version should be provided, if known.
	SystemIdentity SystemBoundaryIdentifier

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetVersionOperationID = "versioningv1.GetVersion"

func (request *GetVersionRequest) OperationID() string { return GetVersionOperationID }

type GetVersionResponseSet struct {
	// This interface successfully provided the version of the system identity/boundary that was requested.
	Response200 *GetVersionResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *api.ClientErrorBody

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *api.ClientErrorBody

	// The requested system identity/boundary is not known, or the versioning automated testing interface is not available.
	Response404 *api.EmptyResponseBody

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// System version
	// ---
	// Get the requested system version.
	GetVersion(ctx context.Context, req *GetVersionRequest) GetVersionResponseSet
}
