// This file is auto-generated; do not change as any changes will be overwritten
package scd_v1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	QueryOperationalIntentReferencesSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	GetOperationalIntentReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	CreateOperationalIntentReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.strategic_coordination", "utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	UpdateOperationalIntentReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.strategic_coordination", "utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	DeleteOperationalIntentReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	QueryConstraintReferencesSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
			{RequiredScopes: []string{"utm.constraint_processing"}},
		},
	}
	GetConstraintReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
			{RequiredScopes: []string{"utm.constraint_processing"}},
		},
	}
	CreateConstraintReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
		},
	}
	UpdateConstraintReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
		},
	}
	DeleteConstraintReferenceSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
		},
	}
	QuerySubscriptionsSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
		},
	}
	GetSubscriptionSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
		},
	}
	CreateSubscriptionSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
		},
	}
	UpdateSubscriptionSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
		},
	}
	DeleteSubscriptionSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
		},
	}
	MakeDssReportSecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.constraint_management"}},
			{RequiredScopes: []string{"utm.constraint_processing"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
			{RequiredScopes: []string{"utm.availability_arbitration"}},
		},
	}
	GetUssAvailabilitySecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.availability_arbitration"}},
			{RequiredScopes: []string{"utm.strategic_coordination"}},
			{RequiredScopes: []string{"utm.conformance_monitoring_sa"}},
		},
	}
	SetUssAvailabilitySecurity = map[string]api.SecurityScheme{
		"Authority": []api.AuthorizationOption{
			{RequiredScopes: []string{"utm.availability_arbitration"}},
		},
	}
)

type QueryOperationalIntentReferencesRequest struct {
	// The data contained in the body of this request, if it parsed correctly
	Body *QueryOperationalIntentReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type QueryOperationalIntentReferencesResponseSet struct {
	// Operational intents were successfully retrieved.
	Response200 *QueryOperationalIntentReferenceResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested area was too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetOperationalIntentReferenceRequest struct {
	// EntityID of the operational intent.
	Entityid EntityID

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetOperationalIntentReferenceResponseSet struct {
	// Operational intent reference was retrieved successfully.
	Response200 *GetOperationalIntentReferenceResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested Entity could not be found.
	Response404 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateOperationalIntentReferenceRequest struct {
	// EntityID of the operational intent.
	Entityid EntityID

	// The data contained in the body of this request, if it parsed correctly
	Body *PutOperationalIntentReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type CreateOperationalIntentReferenceResponseSet struct {
	// An operational intent reference was created successfully in the DSS.
	Response201 *ChangeOperationalIntentReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the operational intent in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * The provided key did not prove knowledge of all current and relevant airspace Entities
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *AirspaceConflictResponse

	// The client attempted to transition the operational intent to Accepted or Activated while marked as Down in the DSS.
	Response412 *ErrorResponse

	// The area of the operational intent is too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type UpdateOperationalIntentReferenceRequest struct {
	// EntityID of the operational intent.
	Entityid EntityID

	// Opaque version number of the existing operational intent reference.
	Ovn EntityOVN

	// The data contained in the body of this request, if it parsed correctly
	Body *PutOperationalIntentReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type UpdateOperationalIntentReferenceResponseSet struct {
	// An operational intent reference was updated successfully in the DSS.
	Response200 *ChangeOperationalIntentReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the operational intent in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * The provided key did not prove knowledge of all current and relevant airspace Entities
	// * The provided `ovn` does not match the current version of the existing operational intent.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *AirspaceConflictResponse

	// The client attempted to transition the operational intent to Accepted or Activated while marked as Down in the DSS.
	Response412 *ErrorResponse

	// The area of the operational intent is too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteOperationalIntentReferenceRequest struct {
	// EntityID of the operational intent.
	Entityid EntityID

	// Opaque version number of the existing operational intent reference.
	Ovn EntityOVN

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type DeleteOperationalIntentReferenceResponseSet struct {
	// The specified operational intent was successfully removed from the DSS.
	Response200 *ChangeOperationalIntentReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the operational intent in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested Entity could not be found.
	Response404 *ErrorResponse

	// * The provided `ovn` does not match the current version of the existing operational intent.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The client attempted to delete the operational intent while marked as Down in the DSS.
	Response412 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type QueryConstraintReferencesRequest struct {
	// The data contained in the body of this request, if it parsed correctly
	Body *QueryConstraintReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type QueryConstraintReferencesResponseSet struct {
	// Constraint references were successfully retrieved.
	Response200 *QueryConstraintReferencesResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested area was too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetConstraintReferenceRequest struct {
	// EntityID of the constraint.
	Entityid EntityID

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetConstraintReferenceResponseSet struct {
	// Constraint reference was retrieved successfully.
	Response200 *GetConstraintReferenceResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested Entity could not be found.
	Response404 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateConstraintReferenceRequest struct {
	// EntityID of the constraint.
	Entityid EntityID

	// The data contained in the body of this request, if it parsed correctly
	Body *PutConstraintReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type CreateConstraintReferenceResponseSet struct {
	// A constraint reference was created successfully in the DSS.
	Response201 *ChangeConstraintReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the constraint in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * A constraint with the provided ID already exists.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The area of the constraint is too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type UpdateConstraintReferenceRequest struct {
	// EntityID of the constraint.
	Entityid EntityID

	// Opaque version number of the existing operational intent reference.
	Ovn EntityOVN

	// The data contained in the body of this request, if it parsed correctly
	Body *PutConstraintReferenceParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type UpdateConstraintReferenceResponseSet struct {
	// A constraint reference was updated successfully in the DSS.
	Response200 *ChangeConstraintReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the constraint in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// * The provided `ovn` does not match the current version of the existing constraint.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The area of the constraint is too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteConstraintReferenceRequest struct {
	// EntityID of the constraint.
	Entityid EntityID

	// Opaque version number of the existing operational intent reference.
	Ovn EntityOVN

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type DeleteConstraintReferenceResponseSet struct {
	// The constraint was successfully removed from the DSS.
	Response200 *ChangeConstraintReferenceResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the constraint in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested Entity could not be found.
	Response404 *ErrorResponse

	// * The provided `ovn` does not match the current version of the existing constraint.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type QuerySubscriptionsRequest struct {
	// The data contained in the body of this request, if it parsed correctly
	Body *QuerySubscriptionParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type QuerySubscriptionsResponseSet struct {
	// Subscriptions were retrieved successfully.
	Response200 *QuerySubscriptionsResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// The requested area was too large.
	Response413 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetSubscriptionRequest struct {
	// SubscriptionID of the subscription of interest.
	Subscriptionid SubscriptionID

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

	// A subscription with the specified ID was not found.
	Response404 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateSubscriptionRequest struct {
	// SubscriptionID of the subscription of interest.
	Subscriptionid SubscriptionID

	// The data contained in the body of this request, if it parsed correctly
	Body *PutSubscriptionParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type CreateSubscriptionResponseSet struct {
	// A new subscription was created successfully.
	Response200 *PutSubscriptionResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the subscription in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	// * Client attempted to request notifications for an Entity type to which the scopes included in the access token do not provide access.
	Response403 *ErrorResponse

	// * A subscription with the specified ID already exists.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The client may have issued too many requests within a small period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type UpdateSubscriptionRequest struct {
	// SubscriptionID of the subscription of interest.
	Subscriptionid SubscriptionID

	// Version of the subscription to be modified.
	Version string

	// The data contained in the body of this request, if it parsed correctly
	Body *PutSubscriptionParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type UpdateSubscriptionResponseSet struct {
	// A subscription was updated successfully.
	Response200 *PutSubscriptionResponse

	// * One or more input parameters were missing or invalid.
	// * The request attempted to mutate the subscription in a disallowed way.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	// * Client attempted to request notifications for an Entity type to which the scopes included in the access token do not provide access.
	Response403 *ErrorResponse

	// * A subscription with the specified ID already exists and is managed by a different client.
	// * The provided `version` does not match the current subscription.
	// * Despite repeated attempts, the DSS was unable to complete the update because of other simultaneous changes.
	Response409 *ErrorResponse

	// The client may have issued too many requests within a small period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteSubscriptionRequest struct {
	// SubscriptionID of the subscription of interest.
	Subscriptionid SubscriptionID

	// Version of the subscription to be modified.
	Version string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type DeleteSubscriptionResponseSet struct {
	// Subscription was successfully removed from DSS.
	Response200 *DeleteSubscriptionResponse

	// One or more input parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// The access token was decoded successfully but did not include a scope appropriate to this endpoint.
	Response403 *ErrorResponse

	// A subscription with the specified ID was not found.
	Response404 *ErrorResponse

	// * A subscription with the specified ID is managed by a different client.
	// * The provided `version` does not match the current subscription.
	// * Despite repeated attempts, the DSS was unable to complete the deletion because of other simultaneous changes.
	Response409 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type MakeDssReportRequest struct {
	// The data contained in the body of this request, if it parsed correctly
	Body *ErrorReport

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type MakeDssReportResponseSet struct {
	// A new Report was created successfully (and archived).
	Response201 *ErrorReport

	// * One or more parameters were missing or invalid.
	// * The report could not be parsed, or contains unrecognized data.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	Response403 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetUssAvailabilityRequest struct {
	// Client ID (matching their `sub` in access tokens) of the USS to which this availability applies.
	UssId string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type GetUssAvailabilityResponseSet struct {
	// Availability status of specified USS was successfully retrieved.
	Response200 *UssAvailabilityStatusResponse

	// * One or more parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	Response403 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type SetUssAvailabilityRequest struct {
	// Client ID (matching their `sub` in access tokens) of the USS to which this availability applies.
	UssId string

	// The data contained in the body of this request, if it parsed correctly
	Body *SetUssAvailabilityStatusParameters

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}
type SetUssAvailabilityResponseSet struct {
	// Availability status of specified USS was successfully updated.
	Response200 *UssAvailabilityStatusResponse

	// * One or more parameters were missing or invalid.
	Response400 *ErrorResponse

	// Bearer access token was not provided in Authorization header, token could not be decoded, or token was invalid.
	Response401 *ErrorResponse

	// * The access token was decoded successfully but did not include a scope appropriate to this endpoint or the request.
	Response403 *ErrorResponse

	// The client issued too many requests in a short period of time.
	Response429 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Query all operational intent references in the specified area/volume/time from the DSS.
	// ---
	// Note that this endpoint does not produce any mutations in the DSS despite using the HTTP POST verb.  The HTTP GET verb is traditionally used for operations like this one, but requiring or using a request body for HTTP GET requests is non-standard and not supported by some architectures.  POST is used here instead of GET to ensure robust support for the use of a request body.
	QueryOperationalIntentReferences(ctx context.Context, req *QueryOperationalIntentReferencesRequest) QueryOperationalIntentReferencesResponseSet

	// Retrieve the specified operational intent reference from the DSS.
	GetOperationalIntentReference(ctx context.Context, req *GetOperationalIntentReferenceRequest) GetOperationalIntentReferenceResponseSet

	// Create the specified operational intent reference in the DSS.
	CreateOperationalIntentReference(ctx context.Context, req *CreateOperationalIntentReferenceRequest) CreateOperationalIntentReferenceResponseSet

	// Update the specified operational intent reference in the DSS.
	UpdateOperationalIntentReference(ctx context.Context, req *UpdateOperationalIntentReferenceRequest) UpdateOperationalIntentReferenceResponseSet

	// Remove the specified operational intent reference from the DSS.
	DeleteOperationalIntentReference(ctx context.Context, req *DeleteOperationalIntentReferenceRequest) DeleteOperationalIntentReferenceResponseSet

	// Query all constraint references in the specified area/volume from the DSS.
	// ---
	// Note that this endpoint does not produce any mutations in the DSS despite using the HTTP POST verb.  The HTTP GET verb is traditionally used for operations like this one, but requiring or using a request body for HTTP GET requests is non-standard and not supported by some architectures.  POST is used here instead of GET to ensure robust support for the use of a request body.
	QueryConstraintReferences(ctx context.Context, req *QueryConstraintReferencesRequest) QueryConstraintReferencesResponseSet

	// Retrieve the specified constraint reference from the DSS.
	GetConstraintReference(ctx context.Context, req *GetConstraintReferenceRequest) GetConstraintReferenceResponseSet

	// Create the specified constraint reference in the DSS.
	CreateConstraintReference(ctx context.Context, req *CreateConstraintReferenceRequest) CreateConstraintReferenceResponseSet

	// Update the specified constraint reference in the DSS.
	UpdateConstraintReference(ctx context.Context, req *UpdateConstraintReferenceRequest) UpdateConstraintReferenceResponseSet

	// Delete the specified constraint reference from the DSS.
	DeleteConstraintReference(ctx context.Context, req *DeleteConstraintReferenceRequest) DeleteConstraintReferenceResponseSet

	// Query all subscriptions in the specified area/volume from the DSS.
	// ---
	// Query subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	// Note that this parameter is a JSON object (in the 'request-body'). Note that either or both of the 'altitude' and 'time' values may be omitted from this parameter.
	// Only subscriptions belonging to the caller are returned.  This endpoint would be used if a USS lost track of subscriptions they had created and/or wanted to resolve an error indicating that they had too many existing subscriptions in an area.
	QuerySubscriptions(ctx context.Context, req *QuerySubscriptionsRequest) QuerySubscriptionsResponseSet

	// Retrieve the specified subscription from the DSS.
	// ---
	// Retrieve a specific subscription.
	GetSubscription(ctx context.Context, req *GetSubscriptionRequest) GetSubscriptionResponseSet

	// Create the specified subscription in the DSS.
	// ---
	// Create a subscription.
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) CreateSubscriptionResponseSet

	// Update the specified subscription in the DSS.
	// ---
	// Update a subscription.
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	// The standard requires each operational intent to have a subscription that cover the 4D volume of the operational intent.  If a USS attempts to update a subscription upon which an operational intent depends, and this update would cause the operational intent to lose subscription coverage, the update will be rejected by the DSS as a bad request.
	UpdateSubscription(ctx context.Context, req *UpdateSubscriptionRequest) UpdateSubscriptionResponseSet

	// Remove the specified subscription from the DSS.
	// ---
	// The standard requires each operational intent to have a subscription that cover the 4D volume of the operational intent.  If a USS attempts to delete a subscription upon which an operational intent depends, the deletion will be rejected by the DSS as a bad request.
	DeleteSubscription(ctx context.Context, req *DeleteSubscriptionRequest) DeleteSubscriptionResponseSet

	// Report information about communication issues to a DSS.
	// ---
	// Report issues to a DSS. Data sent to this endpoint is archived.
	MakeDssReport(ctx context.Context, req *MakeDssReportRequest) MakeDssReportResponseSet

	// Get availability status of a USS.
	GetUssAvailability(ctx context.Context, req *GetUssAvailabilityRequest) GetUssAvailabilityResponseSet

	// Set availability status of a USS.
	SetUssAvailability(ctx context.Context, req *SetUssAvailabilityRequest) SetUssAvailabilityResponseSet
}
