// This file is auto-generated; do not change as any changes will be overwritten
package main

type EmptyResponseBody struct{}

type InternalServerErrorBody struct {
	ErrorMessage string `json:"error_message"`
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
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

	// Internal server error
	Response500 *InternalServerErrorBody
}

type Implementation interface {
	// Query all operational intent references in the specified area/volume/time from the DSS.
	// ---
	// Note that this endpoint does not produce any mutations in the DSS despite using the HTTP POST verb.  The HTTP GET verb is traditionally used for operations like this one, but requiring or using a request body for HTTP GET requests is non-standard and not supported by some architectures.  POST is used here instead of GET to ensure robust support for the use of a request body.
	QueryOperationalIntentReferences(body QueryOperationalIntentReferenceParameters, bodyParseError *error) QueryOperationalIntentReferencesResponseSet

	// Create the specified operational intent reference in the DSS.
	CreateOperationalIntentReference(entityid EntityID, body PutOperationalIntentReferenceParameters, bodyParseError *error) CreateOperationalIntentReferenceResponseSet

	// Retrieve the specified operational intent reference from the DSS.
	GetOperationalIntentReference(entityid EntityID) GetOperationalIntentReferenceResponseSet

	// Remove the specified operational intent reference from the DSS.
	DeleteOperationalIntentReference(entityid EntityID, ovn EntityOVN) DeleteOperationalIntentReferenceResponseSet

	// Update the specified operational intent reference in the DSS.
	UpdateOperationalIntentReference(entityid EntityID, ovn EntityOVN, body PutOperationalIntentReferenceParameters, bodyParseError *error) UpdateOperationalIntentReferenceResponseSet

	// Query all constraint references in the specified area/volume from the DSS.
	// ---
	// Note that this endpoint does not produce any mutations in the DSS despite using the HTTP POST verb.  The HTTP GET verb is traditionally used for operations like this one, but requiring or using a request body for HTTP GET requests is non-standard and not supported by some architectures.  POST is used here instead of GET to ensure robust support for the use of a request body.
	QueryConstraintReferences(body QueryConstraintReferenceParameters, bodyParseError *error) QueryConstraintReferencesResponseSet

	// Create the specified constraint reference in the DSS.
	CreateConstraintReference(entityid EntityID, body PutConstraintReferenceParameters, bodyParseError *error) CreateConstraintReferenceResponseSet

	// Retrieve the specified constraint reference from the DSS.
	GetConstraintReference(entityid EntityID) GetConstraintReferenceResponseSet

	// Delete the specified constraint reference from the DSS.
	DeleteConstraintReference(entityid EntityID, ovn EntityOVN) DeleteConstraintReferenceResponseSet

	// Update the specified constraint reference in the DSS.
	UpdateConstraintReference(entityid EntityID, ovn EntityOVN, body PutConstraintReferenceParameters, bodyParseError *error) UpdateConstraintReferenceResponseSet

	// Query all subscriptions in the specified area/volume from the DSS.
	// ---
	// Query subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	// Note that this parameter is a JSON object (in the 'request-body'). Note that either or both of the 'altitude' and 'time' values may be omitted from this parameter.
	// Only subscriptions belonging to the caller are returned.  This endpoint would be used if a USS lost track of subscriptions they had created and/or wanted to resolve an error indicating that they had too many existing subscriptions in an area.
	QuerySubscriptions(body QuerySubscriptionParameters, bodyParseError *error) QuerySubscriptionsResponseSet

	// Create the specified subscription in the DSS.
	// ---
	// Create a subscription.
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	CreateSubscription(subscriptionid SubscriptionID, body PutSubscriptionParameters, bodyParseError *error) CreateSubscriptionResponseSet

	// Retrieve the specified subscription from the DSS.
	// ---
	// Retrieve a specific subscription.
	GetSubscription(subscriptionid SubscriptionID) GetSubscriptionResponseSet

	// Remove the specified subscription from the DSS.
	// ---
	// The standard requires each operational intent to have a subscription that cover the 4D volume of the operational intent.  If a USS attempts to delete a subscription upon which an operational intent depends, the deletion will be rejected by the DSS as a bad request.
	DeleteSubscription(subscriptionid SubscriptionID, version string) DeleteSubscriptionResponseSet

	// Update the specified subscription in the DSS.
	// ---
	// Update a subscription.
	// Subscription notifications are only triggered by (and contain full information of) changes to, creation of, or deletion of, Entities referenced by or stored in the DSS; they do not involve any data transfer (such as remote ID telemetry updates) apart from Entity information.
	// The standard requires each operational intent to have a subscription that cover the 4D volume of the operational intent.  If a USS attempts to update a subscription upon which an operational intent depends, and this update would cause the operational intent to lose subscription coverage, the update will be rejected by the DSS as a bad request.
	UpdateSubscription(subscriptionid SubscriptionID, version string, body PutSubscriptionParameters, bodyParseError *error) UpdateSubscriptionResponseSet

	// Report information about communication issues to a DSS.
	// ---
	// Report issues to a DSS. Data sent to this endpoint is archived.
	MakeDssReport(body ErrorReport, bodyParseError *error) MakeDssReportResponseSet

	// Set availability status of a USS.
	SetUssAvailability(uss_id string, body SetUssAvailabilityStatusParameters, bodyParseError *error) SetUssAvailabilityResponseSet

	// Get availability status of a USS.
	GetUssAvailability(uss_id string) GetUssAvailabilityResponseSet
}
