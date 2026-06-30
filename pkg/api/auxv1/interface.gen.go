// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	DssReadIdentificationServiceAreasScope  = api.RequiredScope("dss.read.identification_service_areas")
	DssWriteIdentificationServiceAreasScope = api.RequiredScope("dss.write.identification_service_areas")
	InterussPoolStatusHeartbeatWriteScope   = api.RequiredScope("interuss.pool_status.heartbeat.write")
	InterussPoolStatusReadScope             = api.RequiredScope("interuss.pool_status.read")
	GetVersionSecurity                      = []api.AuthorizationOption{}
	ValidateOauthSecurity                   = []api.AuthorizationOption{
		{
			"Auth": {DssReadIdentificationServiceAreasScope},
		},
		{
			"Auth": {DssWriteIdentificationServiceAreasScope},
		},
	}
	GetPoolSecurity = []api.AuthorizationOption{
		{
			"Auth": {InterussPoolStatusReadScope},
		},
	}
	GetDSSInstancesSecurity = []api.AuthorizationOption{
		{
			"Auth": {InterussPoolStatusReadScope},
		},
	}
	PutDSSInstancesHeartbeatSecurity = []api.AuthorizationOption{
		{
			"Auth": {InterussPoolStatusHeartbeatWriteScope},
		},
	}
	GetAcceptedCAsSecurity   = []api.AuthorizationOption{}
	GetInstanceCAsSecurity   = []api.AuthorizationOption{}
	GetGlobalOptionsSecurity = []api.AuthorizationOption{
		{
			"Auth": {InterussPoolStatusReadScope},
		},
	}
)

type GetVersionRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetVersionRequestType = "GetVersion"

func (req *GetVersionRequest) RequestType() string { return GetVersionRequestType }

func (req *GetVersionRequest) IsReadOnly() bool { return true }

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

const ValidateOauthRequestType = "ValidateOauth"

func (req *ValidateOauthRequest) RequestType() string { return ValidateOauthRequestType }

func (req *ValidateOauthRequest) IsReadOnly() bool { return false }

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

type GetPoolRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetPoolRequestType = "GetPool"

func (req *GetPoolRequest) RequestType() string { return GetPoolRequestType }

func (req *GetPoolRequest) IsReadOnly() bool { return true }

type GetPoolResponseSet struct {
	// The information is successfully returned.
	Response200 *PoolResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetDSSInstancesRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetDSSInstancesRequestType = "GetDSSInstances"

func (req *GetDSSInstancesRequest) RequestType() string { return GetDSSInstancesRequestType }

func (req *GetDSSInstancesRequest) IsReadOnly() bool { return true }

type GetDSSInstancesResponseSet struct {
	// The known DSS instances participating in the pool are successfully returned.
	Response200 *DSSInstancesResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type PutDSSInstancesHeartbeatRequest struct {
	// The source of the timestamp
	Source *string

	// Override the timestamp value of the heartbeat. If not set, will use the current time. RFC 3339 format.
	Timestamp *string

	// Set the time before the next heartbeat is expected. RFC 3339 format.
	NextHeartbeatExpectedBefore *string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const PutDSSInstancesHeartbeatRequestType = "PutDSSInstancesHeartbeat"

func (req *PutDSSInstancesHeartbeatRequest) RequestType() string {
	return PutDSSInstancesHeartbeatRequestType
}

func (req *PutDSSInstancesHeartbeatRequest) IsReadOnly() bool { return false }

type PutDSSInstancesHeartbeatResponseSet struct {
	// The heartbeat have been recorded. The known DSS instances participating in the pool are successfully returned.
	Response201 *DSSInstancesResponse

	// The request was not properly formed or the parameters are invalid
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetAcceptedCAsRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetAcceptedCAsRequestType = "GetAcceptedCAs"

func (req *GetAcceptedCAsRequest) RequestType() string { return GetAcceptedCAsRequestType }

func (req *GetAcceptedCAsRequest) IsReadOnly() bool { return true }

type GetAcceptedCAsResponseSet struct {
	// The information is successfully returned.
	Response200 *CAsResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetInstanceCAsRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetInstanceCAsRequestType = "GetInstanceCAs"

func (req *GetInstanceCAsRequest) RequestType() string { return GetInstanceCAsRequestType }

func (req *GetInstanceCAsRequest) IsReadOnly() bool { return true }

type GetInstanceCAsResponseSet struct {
	// The information is successfully returned.
	Response200 *CAsResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetGlobalOptionsRequest struct {
	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetGlobalOptionsRequestType = "GetGlobalOptions"

func (req *GetGlobalOptionsRequest) RequestType() string { return GetGlobalOptionsRequestType }

func (req *GetGlobalOptionsRequest) IsReadOnly() bool { return true }

type GetGlobalOptionsResponseSet struct {
	// The information is successfully returned.
	Response200 *GlobalOptionsResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The server has not implemented this operation.
	Response501 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Queries the version of the DSS.
	GetVersion(ctx context.Context, req *GetVersionRequest) GetVersionResponseSet

	// Validate Oauth token against the DSS.
	ValidateOauth(ctx context.Context, req *ValidateOauthRequest) ValidateOauthResponseSet

	// Queries the current information about the pool of DSS instances constituting the DSS Airspace Representation.
	GetPool(ctx context.Context, req *GetPoolRequest) GetPoolResponseSet

	// Queries the current information for DSS instances participating in the pool.
	GetDSSInstances(ctx context.Context, req *GetDSSInstancesRequest) GetDSSInstancesResponseSet

	// Record a new heartbeat from the DSS instance
	PutDSSInstancesHeartbeat(ctx context.Context, req *PutDSSInstancesHeartbeatRequest) PutDSSInstancesHeartbeatResponseSet

	// Current certificates of certificate authorities (CAs) that this DSS instance accepts as legitimate signers of node certificates for the pool of DSS instances constituting the DSS Airspace Representation.
	GetAcceptedCAs(ctx context.Context, req *GetAcceptedCAsRequest) GetAcceptedCAsResponseSet

	// Current certificates of certificate authorities (CAs) that signed the node certificates for this DSS instance. May return more that one certificate (e.g. for rotations).  Other DSS instances in the pool should accept node certificates signed by these CAs.
	GetInstanceCAs(ctx context.Context, req *GetInstanceCAsRequest) GetInstanceCAsResponseSet

	// Return the value of important behaviour options that need to be set to a common value. May be used to ensure all participants in a pool are using the same value.
	GetGlobalOptions(ctx context.Context, req *GetGlobalOptionsRequest) GetGlobalOptionsResponseSet
}
