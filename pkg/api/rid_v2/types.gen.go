// This file is auto-generated; do not change as any changes will be overwritten
package rid_v2

type Time struct {
	// RFC3339-formatted time/date string.  The time zone must be 'Z'.
	Value string `json:"value"`

	Format string `json:"format"`
}

type Radius struct {
	// Distance from the centerpoint of a circular area, along the WGS84 ellipsoid.
	Value float32 `json:"value"`

	// FIXM-compatible units.  Only meters ("M") are acceptable for UTM.
	Units string `json:"units"`
}

// A circular area on the surface of the earth.
type Circle struct {
	Center *LatLngPoint `json:"center"`

	Radius *Radius `json:"radius"`
}

// A three-dimensional geographic volume consisting of a vertically-extruded shape. Exactly one outline must be specified.
type Volume3D struct {
	// A circular geographic shape on the surface of the earth.
	OutlineCircle *Circle `json:"outline_circle"`

	// A polygonal geographic shape on the surface of the earth.
	OutlinePolygon *Polygon `json:"outline_polygon"`

	// Minimum bounding altitude of this volume. Must be less than altitude_upper, if specified.
	AltitudeLower *Altitude `json:"altitude_lower"`

	// Maximum bounding altitude of this volume. Must be greater than altitude_lower, if specified.
	AltitudeUpper *Altitude `json:"altitude_upper"`
}

// Contiguous block of geographic spacetime.
type Volume4D struct {
	Volume Volume3D `json:"volume"`

	// Beginning time of this volume. Must be before time_end.
	TimeStart *Time `json:"time_start"`

	// End time of this volume. Must be after time_start.
	TimeEnd *Time `json:"time_end"`
}

// Response to DSS request for the subscription with the given id.
type GetSubscriptionResponse struct {
	Subscription Subscription `json:"subscription"`
}

// Response to DSS query for subscriptions in a particular area.
type SearchSubscriptionsResponse struct {
	// Subscriptions that overlap the specified area.
	Subscriptions *[]Subscription `json:"subscriptions"`
}

// Valid http or https URL.
type URL string

// Tracks the notifications sent for a subscription so the subscriber can detect missed notifications more easily.
type SubscriptionNotificationIndex int32

// State of Subscription which is causing a notification to be sent.
type SubscriptionState struct {
	SubscriptionId SubscriptionUUID `json:"subscription_id"`

	NotificationIndex *SubscriptionNotificationIndex `json:"notification_index"`
}

// UUID v4.
type UUIDv4 string

// A version string used to reference an object at a particular point in time. Any updates to an object must contain the corresponding version to maintain idempotent updates.
type Version string

// Universally-unique identifier for an Entity communicated through the DSS.  Formatted as UUIDv4.
type EntityUUID UUIDv4

// Universally-unique identifier for a Subscription communicated through the DSS.  Formatted as UUIDv4.
type SubscriptionUUID UUIDv4

// Data provided when an off-nominal condition was encountered.
type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message"`
}

// Response for a successful request to delete an Subscription.
type DeleteSubscriptionResponse struct {
	// The Subscription which was deleted.
	Subscription Subscription `json:"subscription"`
}

// Plain-string representation of a geographic polygon consisting of at least three geographic
// points describing a closed polygon on the earth.  Each point consists of latitude,longitude
// in degrees.  Points are also comma-delimited, so this parameter will look like
// `lat1,lng1,lat2,lng2,lat3,lng3,...`  Latitude values must fall in the range [-90, 90] and
// longitude values must fall in the range [-180, 180].
//
// All of the requirements and clarifications for Polygon apply to GeoPolygonString as well.
type GeoPolygonString string

// Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid.  Invalid, No Value, or Unknown is 0 degrees (both Lat/Lon).
type Latitude float64

// Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid.  Invalid, No Value, or Unknown is 0 degrees (both Lat/Lon).
type Longitude float64

// Point on the earth's surface.
type LatLngPoint struct {
	Lng Longitude `json:"lng"`

	Lat Latitude `json:"lat"`
}

type Altitude struct {
	// The numeric value of the altitude. Note that min and max values are added as a sanity check. As use cases evolve and more options are made available in terms of units of measure or reference systems, these bounds may be re-evaluated. Invalid, No Value, or Unknown is –1000 m.
	Value float64 `json:"value"`

	// A code indicating the reference for a vertical distance. See AIXM 5.1 and FIXM 4.2.0. Currently, UTM only allows WGS84 with no immediate plans to allow other options. FIXM and AIXM allow for 'SFC' which is equivalent to AGL.
	Reference string `json:"reference"`

	// The reference quantities used to express the value of altitude. See FIXM 4.2. Currently, UTM only allows meters with no immediate plans to allow other options.
	Units string `json:"units"`
}

// An enclosed area on the earth. The bounding edges of this polygon are defined to be the shortest paths between connected vertices.  This means, for instance, that the edge between two points both defined at a particular latitude is not generally contained at that latitude. The winding order must be interpreted as the order which produces the smaller area. The path between two vertices is defined to be the shortest possible path between those vertices. Edges may not cross. Vertices may not be duplicated.  In particular, the final polygon vertex must not be identical to the first vertex.
type Polygon struct {
	Vertices []LatLngPoint `json:"vertices"`
}

// Response to a request to create or update a reference to an Identification Service Area in the DSS.
type PutIdentificationServiceAreaResponse struct {
	// DSS subscribers that this client now has the obligation to notify of the Identification Service Area changes just made.  This client must call POST for each provided URL according to the `/uss/identification_service_areas/{id}` path API.
	Subscribers *[]SubscriberToNotify `json:"subscribers"`

	// Resulting service area stored in DSS.
	ServiceArea IdentificationServiceArea `json:"service_area"`
}

// Response to DSS query for Identification Service Areas in an area of interest.
type SearchIdentificationServiceAreasResponse struct {
	// Identification Service Areas in the area of interest.
	ServiceAreas *[]IdentificationServiceArea `json:"service_areas"`
}

// Subscriber to notify of a creation/change/deletion of a change in the airspace.  This is provided by the DSS to a client changing the airspace, and it is the responsibility of the client changing the airspace (they will receive a set of these notification requests) to send a notification to each specified `url`.
type SubscriberToNotify struct {
	// Subscription(s) prompting this notification.
	Subscriptions []SubscriptionState `json:"subscriptions"`

	// The endpoint that the client mutating the airspace should provide the update to.  API depends on the DSS action taken that triggered this notification request.
	Url URL `json:"url"`
}

// Response for a request to delete an Identification Service Area.
type DeleteIdentificationServiceAreaResponse struct {
	// Identification Service Area that was just deleted.
	ServiceArea IdentificationServiceArea `json:"service_area"`

	// DSS subscribers that this client now has the obligation to notify of the Identification Service Area just deleted.  This client must call POST for each provided URL according to the `/uss/identification_service_areas` path API.
	Subscribers *[]SubscriberToNotify `json:"subscribers"`
}

// Response for a request to create or update a subscription.
type PutSubscriptionResponse struct {
	// Identification Service Areas in or near the subscription area at the time of creation/update, if `identification_service_area_url` callback was specified.
	ServiceAreas *[]IdentificationServiceArea `json:"service_areas"`

	// Result of the operation on the subscription.
	Subscription Subscription `json:"subscription"`
}

// Response to DSS request for the identification service area with the given ID.
type GetIdentificationServiceAreaResponse struct {
	ServiceArea IdentificationServiceArea `json:"service_area"`
}

// Parameters for a request to create an Identification Service Area in the DSS.
type CreateIdentificationServiceAreaParameters struct {
	// The bounding spacetime extents of this Identification Service Area.  End time must be specified.  If start time is not specified, it will be set to the current time.  Start times in the past should be rejected by the DSS, except that it may adjust very recent start times to the current time.
	//
	// These extents should not reveal any sensitive information about the flight or flights within them.  This means, for instance, that extents should not tightly-wrap a flight path, nor should they generally be centered around the takeoff point of a single flight.
	Extents Volume4D `json:"extents"`

	UssBaseUrl FlightsUSSBaseURL `json:"uss_base_url"`
}

// Parameters for a request to update an Identification Service Area in the DSS.
type UpdateIdentificationServiceAreaParameters struct {
	// The bounding spacetime extents of this Identification Service Area.  End time must be specified.  If start time is not specified, it will remain unchanged.  Start times in the past should be rejected by the DSS unless they are unchanged from the Identification Service Area's current start time.
	//
	// These extents should not reveal any sensitive information about the flight or flights within them.  This means, for instance, that extents should not tightly-wrap a flight path, nor should they generally be centered around the takeoff point of a single flight.
	Extents Volume4D `json:"extents"`

	UssBaseUrl FlightsUSSBaseURL `json:"uss_base_url"`
}

// Parameters for a request to create a subscription in the DSS.
type CreateSubscriptionParameters struct {
	// The spacetime extents of the volume to subscribe to.
	//
	// This subscription will automatically be deleted after its end time if it has not been refreshed by then.  If end time is not specified, the value will be chosen automatically by the DSS.
	//
	// Note that some Entities triggering notifications may lie entirely outside the requested area.
	Extents Volume4D `json:"extents"`

	UssBaseUrl SubscriptionUSSBaseURL `json:"uss_base_url"`
}

// Parameters for a request to update a subscription in the DSS.
type UpdateSubscriptionParameters struct {
	// The spacetime extents of the volume to subscribe to.
	//
	// This subscription will automatically be deleted after its end time if it has not been refreshed by then.  If end time is not specified, the value will be chosen automatically by the DSS.
	//
	// Note that some Entities triggering notifications may lie entirely outside the requested area.
	Extents Volume4D `json:"extents"`

	UssBaseUrl SubscriptionUSSBaseURL `json:"uss_base_url"`
}

// Specification of a geographic area that a client is interested in on an ongoing basis (e.g., "planning area").  Internal to the DSS.
type Subscription struct {
	// Unique identifier for this subscription.
	Id SubscriptionUUID `json:"id"`

	UssBaseUrl SubscriptionUSSBaseURL `json:"uss_base_url"`

	// Assigned by the DSS based on creating client’s ID (via access token).  Used for restricting mutation and deletion operations to owner.
	Owner string `json:"owner"`

	NotificationIndex *SubscriptionNotificationIndex `json:"notification_index"`

	// If set, this subscription will be automatically removed after this time.
	TimeEnd *Time `json:"time_end"`

	// If set, this Subscription will not generate any notifications before this time.
	TimeStart *Time `json:"time_start"`

	Version Version `json:"version"`
}

// An Identification Service Area (area in which remote ID services are being provided).  The DSS reports only these declarations and clients must exchange flight information peer-to-peer.
type IdentificationServiceArea struct {
	UssBaseUrl FlightsUSSBaseURL `json:"uss_base_url"`

	// Assigned by the DSS based on creating client’s ID (via access token).  Used for restricting mutation and deletion operations to owner.
	Owner string `json:"owner"`

	// Beginning time of service.
	TimeStart Time `json:"time_start"`

	// End time of service.
	TimeEnd Time `json:"time_end"`

	Version Version `json:"version"`

	// Unique identifier for this Identification Service Area.
	Id EntityUUID `json:"id"`
}

// The base URL of a USS implementation of part or all of the remote ID USS-USS API. Per the USS-USS API, the full URL
// of a particular resource can be constructed by appending, e.g., `/uss/{resource}/{id}` to this base URL.
// Accordingly, this URL may not have a trailing '/' character.
type USSBaseURL string

// Base URL for the USS's implementation of the USS API, including the POST identification_service_areas endpoint where notifications for this Subscription will be received.
type SubscriptionUSSBaseURL USSBaseURL

// Base URL for the USS's implementation of the USS API, including the GET flights and GET flights/{id}/details endpoints where flight information can be obtained.
type FlightsUSSBaseURL USSBaseURL
