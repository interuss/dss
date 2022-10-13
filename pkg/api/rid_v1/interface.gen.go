// This file is auto-generated; do not change as any changes will be overwritten
package rid_v1

import (
	"context"

	"github.com/interuss/dss/pkg/api"
)

var (
	SearchIdentificationServiceAreasSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
	GetIdentificationServiceAreaSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
	CreateIdentificationServiceAreaSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.write.identification_service_areas"}},
		},
	}
	UpdateIdentificationServiceAreaSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.write.identification_service_areas"}},
		},
	}
	DeleteIdentificationServiceAreaSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.write.identification_service_areas"}},
		},
	}
	SearchSubscriptionsSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
	GetSubscriptionSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas", "dss.write.identification_service_areas"}},
		},
	}
	CreateSubscriptionSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
	UpdateSubscriptionSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
	DeleteSubscriptionSecurity = map[string]api.SecurityScheme{
		"AuthFromAuthorizationAuthority": []api.AuthorizationOption{
			{RequiredScopes: []string{"dss.read.identification_service_areas"}},
		},
	}
)

type SearchIdentificationServiceAreasRequest struct {
	// The area in which to search for Identification Service Areas.  Some Identification Service Areas near this area but wholly outside it may also be returned.
	Area *GeoPolygonString

	// If specified, indicates non-interest in any Identification Service Areas that end before this time.  RFC 3339 format, per OpenAPI specification.
	EarliestTime *string

	// If specified, indicates non-interest in any Identification Service Areas that start after this time.  RFC 3339 format, per OpenAPI specification.
	LatestTime *string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type SearchIdentificationServiceAreasResponseSet struct {
	// Identification Service Areas were successfully retrieved.
	Response200 *SearchIdentificationServiceAreasResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested area was too large.
	Response413 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetIdentificationServiceAreaRequest struct {
	// EntityUUID of the Identification Service Area.
	Id EntityUUID

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetIdentificationServiceAreaResponseSet struct {
	// Full information of the Identification Service Area was retrieved successfully.
	Response200 *GetIdentificationServiceAreaResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested Entity could not be found.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateIdentificationServiceAreaRequest struct {
	// EntityUUID of the Identification Service Area.
	Id EntityUUID

	// The data contained in the body of this request, if it parsed correctly
	Body *CreateIdentificationServiceAreaParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type CreateIdentificationServiceAreaResponseSet struct {
	// An existing Identification Service Area was created successfully in the DSS.
	Response200 *PutIdentificationServiceAreaResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the Identification Service Area in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * An Identification Service Area with the specified ID already exists and is owned by a different client.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// The area of the Identification Service Area is too large.
	Response413 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type UpdateIdentificationServiceAreaRequest struct {
	// EntityUUID of the Identification Service Area.
	Id EntityUUID

	// Version string used to reference an Identification Service Area at a particular point in time. Any updates to an existing Identification Service Area must contain the corresponding version to maintain idempotent updates.
	Version string

	// The data contained in the body of this request, if it parsed correctly
	Body *UpdateIdentificationServiceAreaParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type UpdateIdentificationServiceAreaResponseSet struct {
	// An existing Identification Service Area was updated successfully in the DSS.
	Response200 *PutIdentificationServiceAreaResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the Identification Service Area in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * The specified Identification Service Area's current version does not match the provided version.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// The area of the Identification Service Area is too large.
	Response413 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteIdentificationServiceAreaRequest struct {
	// EntityUUID of the Identification Service Area.
	Id EntityUUID

	// Version string used to reference an Identification Service Area at a particular point in time. Any updates to an existing Identification Service Area must contain the corresponding version to maintain idempotent updates.
	Version string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type DeleteIdentificationServiceAreaResponseSet struct {
	// Identification Service Area was successfully deleted from DSS.
	Response200 *DeleteIdentificationServiceAreaResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	// * The Identification Service Area does not belong to the client requesting deletion.
	Response403 *ErrorResponse

	// Entity could not be deleted because it could not be found.
	Response404 *ErrorResponse

	// * The specified Identification Service Area's current version does not match the provided version.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type SearchSubscriptionsRequest struct {
	// The area in which to search for Subscriptions.  Some Subscriptions near this area but wholly outside it may also be returned.
	Area *GeoPolygonString

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type SearchSubscriptionsResponseSet struct {
	// Subscriptions were retrieved successfully.
	Response200 *SearchSubscriptionsResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested area was too large.
	Response413 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetSubscriptionRequest struct {
	// SubscriptionUUID of the subscription of interest.
	Id SubscriptionUUID

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetSubscriptionResponseSet struct {
	// Subscription information was retrieved successfully.
	Response200 *GetSubscriptionResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// A Subscription with the specified ID was not found.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateSubscriptionRequest struct {
	// SubscriptionUUID of the subscription of interest.
	Id SubscriptionUUID

	// The data contained in the body of this request, if it parsed correctly
	Body *CreateSubscriptionParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type CreateSubscriptionResponseSet struct {
	// A new Subscription was created successfully.
	Response200 *PutSubscriptionResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the Subscription in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	// * An EntityType was specified in `types_filter` to which the scopes included in the access token do not provide access.
	Response403 *ErrorResponse

	// * A Subscription with the specified ID already exists and is owned by a different client.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// Client already has too many Subscriptions in the area where a new Subscription was requested.  To correct this problem, the client may query GET /subscriptions to see which Subscriptions are counting against their limit.  This problem should not generally be encountered because the Subscription limit should be above what any consumer that reasonably aggregates their Subscriptions should request.  But, a Subscription limit is necessary to bound performance requirements for DSS instances and would likely be hit by, e.g., a large remote ID display provider that created a Subscription for each of their display client users' views.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type UpdateSubscriptionRequest struct {
	// SubscriptionUUID of the subscription of interest.
	Id SubscriptionUUID

	// Version string used to reference a Subscription at a particular point in time. Any updates to an existing Subscription must contain the corresponding version to maintain idempotent updates.
	Version string

	// The data contained in the body of this request, if it parsed correctly
	Body *UpdateSubscriptionParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type UpdateSubscriptionResponseSet struct {
	// An existing Subscription was updated successfully.
	Response200 *PutSubscriptionResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the Subscription in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	// * An EntityType was specified in `types_filter` to which the scopes included in the access token do not provide access.
	Response403 *ErrorResponse

	// * The specified Subscriptions's current version does not match the provided version.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// Client already has too many Subscriptions in the area where a new Subscription was requested.  To correct this problem, the client may query GET /subscriptions to see which Subscriptions are counting against their limit.  This problem should not generally be encountered because the Subscription limit should be above what any consumer that reasonably aggregates their Subscriptions should request.  But, a Subscription limit is necessary to bound performance requirements for DSS instances and would likely be hit by, e.g., a large remote ID display provider that created a Subscription for each of their display client users' views.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteSubscriptionRequest struct {
	// SubscriptionUUID of the subscription of interest.
	Id SubscriptionUUID

	// Version string used to reference a Subscription at a particular point in time. Any updates to an existing Subscription must contain the corresponding version to maintain idempotent updates.
	Version string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type DeleteSubscriptionResponseSet struct {
	// Subscription was deleted successfully.
	Response200 *DeleteSubscriptionResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	// * The Entity does not belong to the client requesting deletion.
	Response403 *ErrorResponse

	// Subscription could not be deleted because it could not be found.
	Response404 *ErrorResponse

	// * The specified Subscriptions's current version does not match the provided version.
	// * Despite repeated attempts, the DSS was unable to update the DAR because of other simultaneous changes.
	Response409 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// /dss/identification_service_areas
	// ---
	// Retrieve all Identification Service Areas in the DAR for a given area during the given time.  Note that some Identification Service Areas returned may lie entirely outside the requested area.
	SearchIdentificationServiceAreas(ctx context.Context, req *SearchIdentificationServiceAreasRequest) SearchIdentificationServiceAreasResponseSet

	// /dss/identification_service_areas/{id}
	// ---
	// Retrieve full information of an Identification Service Area owned by the client.
	GetIdentificationServiceArea(ctx context.Context, req *GetIdentificationServiceAreaRequest) GetIdentificationServiceAreaResponseSet

	// /dss/identification_service_areas/{id}
	// ---
	// Create a new Identification Service Area.  This call will fail if an Identification Service Area with the same ID already exists.
	//
	// The DSS assumes the USS has already added the appropriate retention period to operation end time in `time_end` field before storing it.
	CreateIdentificationServiceArea(ctx context.Context, req *CreateIdentificationServiceAreaRequest) CreateIdentificationServiceAreaResponseSet

	// /dss/identification_service_areas/{id}/{version}
	// ---
	// Update an Identification Service Area.  The full content of the existing Identification Service Area will be replaced with the provided information as only the most recent version is retained.
	//
	// The DSS assumes the USS has already added the appropriate retention period to operation end time in `time_end` field before storing it.  Updating `time_start` is not allowed if it is before the current time.
	UpdateIdentificationServiceArea(ctx context.Context, req *UpdateIdentificationServiceAreaRequest) UpdateIdentificationServiceAreaResponseSet

	// /dss/identification_service_areas/{id}/{version}
	// ---
	// Delete an Identification Service Area.  USSs should not delete Identification Service Areas before the end of the last managed flight plus the retention period.
	DeleteIdentificationServiceArea(ctx context.Context, req *DeleteIdentificationServiceAreaRequest) DeleteIdentificationServiceAreaResponseSet

	// /dss/subscriptions
	// ---
	// Retrieve subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	//
	// Only Subscriptions belonging to the caller are returned.  This endpoint would be used if a USS lost track of Subscriptions they had created and/or wanted to resolve an error indicating that they had too many existing Subscriptions in an area.
	SearchSubscriptions(ctx context.Context, req *SearchSubscriptionsRequest) SearchSubscriptionsResponseSet

	// /dss/subscriptions/{id}
	// ---
	// Verify the existence/valdity and state of a particular subscription.
	GetSubscription(ctx context.Context, req *GetSubscriptionRequest) GetSubscriptionResponseSet

	// /dss/subscriptions/{id}
	// ---
	// Create a subscription.  This call will fail if a Subscription with the same ID already exists.
	//
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) CreateSubscriptionResponseSet

	// /dss/subscriptions/{id}/{version}
	// ---
	// Update a Subscription.  The full content of the existing Subscription will be replaced with the provided information as only the most recent version is retained.
	//
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	UpdateSubscription(ctx context.Context, req *UpdateSubscriptionRequest) UpdateSubscriptionResponseSet

	// /dss/subscriptions/{id}/{version}
	// ---
	// Delete a subscription.
	DeleteSubscription(ctx context.Context, req *DeleteSubscriptionRequest) DeleteSubscriptionResponseSet
}
