// This file is auto-generated; do not change as any changes will be overwritten
package rid

// A three-dimensional geographic volume consisting of a vertically-extruded polygon.
type Volume3D struct {
	// Projection of this volume onto the earth's surface.
	Footprint GeoPolygon `json:"footprint"`

	// Minimum bounding altitude of this volume.
	AltitudeLo *Altitude `json:"altitude_lo"`

	// Maximum bounding altitude of this volume.
	AltitudeHi *Altitude `json:"altitude_hi"`
}

// Contiguous block of geographic spacetime.
type Volume4D struct {
	// Constant spatial extent of this volume.
	SpatialVolume Volume3D `json:"spatial_volume"`

	// Beginning time of this volume.  RFC 3339 format, per OpenAPI specification.
	TimeStart *string `json:"time_start"`

	// End time of this volume.  RFC 3339 format, per OpenAPI specification.
	TimeEnd *string `json:"time_end"`
}

// Response to DSS request for the subscription with the given id.
type GetSubscriptionResponse struct {
	Subscription Subscription `json:"subscription"`
}

// Response to DSS query for subscriptions in a particular area.
type SearchSubscriptionsResponse struct {
	// Subscriptions that overlap the specified area.
	Subscriptions []Subscription `json:"subscriptions"`
}

// Response to remote ID provider query for flight information in an area of interest.
type GetFlightsResponse struct {
	// The remote ID service provider's timestamp for when this information was current.  RFC 3339 format, per OpenAPI specification.
	Timestamp string `json:"timestamp"`

	// A list of all flights that have been within the requested area within the remote ID retention period.  This includes flights that are not currently within the requested area, but were within the requested area within the remote ID retention period.
	Flights []RIDFlight `json:"flights"`
}

// Valid http or https URL.
type URL string

// Tracks the notifications sent for a subscription so the subscriber can detect missed notifications more easily.
type SubscriptionNotificationIndex int32

// State of AreaSubscription which is causing a notification to be sent.
type SubscriptionState struct {
	SubscriptionId *SubscriptionUUID `json:"subscription_id"`

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

// ID, unique to a remote ID service provider, which identifies a particular flight for which the remote ID service provider is providing remote ID services.
//
// The following characters are not allowed: \0 \t \r \n # % / : ? @ [ \ ]
type RIDFlightID string

// Additional authentication data.
type RIDAuthData struct {
	// Format of additional authentication data.
	Format string `json:"format"`

	// Authentication data in form specified by `format`.
	Data string `json:"data"`
}

// Response to remote ID provider query for details about a specific flight.
type GetFlightDetailsResponse struct {
	Details RIDFlightDetails `json:"details"`
}

// This is the NACp enumeration from ADS-B, plus 1m for a more complete range for UAs.
//
// `HAUnknown`: Unknown horizontal accuracy
//
// `HA10NMPlus`: > 10NM (18.52km)
//
// `HA10NM`: < 10NM (18.52km)
//
// `HA4NM`: < 4NM (7.408km)
//
// `HA2NM`: < 2NM (3.704km)
//
// `HA1NM`: < 1NM (1852m)
//
// `HA05NM`: < 0.5NM (926m)
//
// `HA03NM`: < 0.3NM (555.6m)
//
// `HA01NM`: < 0.1NM (185.2m)
//
// `HA005NM`: < 0.05NM (92.6m)
//
// `HA30m`: < 30m
//
// `HA10m`: < 10m
//
// `HA3m`: < 3m
//
// `HA1m`: < 1m
type HorizontalAccuracy string

const (
	HorizontalAccuracy_HAUnknown  HorizontalAccuracy = "HAUnknown"
	HorizontalAccuracy_HA10NMPlus HorizontalAccuracy = "HA10NMPlus"
	HorizontalAccuracy_HA10NM     HorizontalAccuracy = "HA10NM"
	HorizontalAccuracy_HA4NM      HorizontalAccuracy = "HA4NM"
	HorizontalAccuracy_HA2NM      HorizontalAccuracy = "HA2NM"
	HorizontalAccuracy_HA1NM      HorizontalAccuracy = "HA1NM"
	HorizontalAccuracy_HA05NM     HorizontalAccuracy = "HA05NM"
	HorizontalAccuracy_HA03NM     HorizontalAccuracy = "HA03NM"
	HorizontalAccuracy_HA01NM     HorizontalAccuracy = "HA01NM"
	HorizontalAccuracy_HA005NM    HorizontalAccuracy = "HA005NM"
	HorizontalAccuracy_HA30m      HorizontalAccuracy = "HA30m"
	HorizontalAccuracy_HA10m      HorizontalAccuracy = "HA10m"
	HorizontalAccuracy_HA3m       HorizontalAccuracy = "HA3m"
	HorizontalAccuracy_HA1m       HorizontalAccuracy = "HA1m"
)

// This is the GVA enumeration from ADS-B, plus some finer values for UAs.
//
// `VAUnknown`: Unknown vertical accuracy
//
// `VA150mPlus`: > 150m
//
// `VA150m`: < 150m
//
// `VA45m`: < 45m
//
// `VA25m`: < 25m
//
// `VA10m`: < 10m
//
// `VA3m`: < 3m
//
// `VA1m`: < 1m
type VerticalAccuracy string

const (
	VerticalAccuracy_VAUnknown  VerticalAccuracy = "VAUnknown"
	VerticalAccuracy_VA150mPlus VerticalAccuracy = "VA150mPlus"
	VerticalAccuracy_VA150m     VerticalAccuracy = "VA150m"
	VerticalAccuracy_VA45m      VerticalAccuracy = "VA45m"
	VerticalAccuracy_VA25m      VerticalAccuracy = "VA25m"
	VerticalAccuracy_VA10m      VerticalAccuracy = "VA10m"
	VerticalAccuracy_VA3m       VerticalAccuracy = "VA3m"
	VerticalAccuracy_VA1m       VerticalAccuracy = "VA1m"
)

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

// This is the same enumeration scale and values from ADS-B NACv.
//
// `SAUnknown`: Unknown speed accuracy
//
// `SA10mpsPlus`: > 10 m/s
//
// `SA10mps`: < 10 m/s
//
// `SA3mps`: < 3 m/s
//
// `SA1mps`: < 1 m/s
//
// `SA03mps`: < 0.3 m/s
type SpeedAccuracy string

const (
	SpeedAccuracy_SAUnknown   SpeedAccuracy = "SAUnknown"
	SpeedAccuracy_SA10mpsPlus SpeedAccuracy = "SA10mpsPlus"
	SpeedAccuracy_SA10mps     SpeedAccuracy = "SA10mps"
	SpeedAccuracy_SA3mps      SpeedAccuracy = "SA3mps"
	SpeedAccuracy_SA1mps      SpeedAccuracy = "SA1mps"
	SpeedAccuracy_SA03mps     SpeedAccuracy = "SA03mps"
)

// Geodetic altitude (NOT altitude above launch, altitude above ground, or EGM96): aircraft distance above the WGS84 ellipsoid as measured along a line that passes through the aircraft and is normal to the surface of the WGS84 ellipsoid.  This value is provided in meters and must have a minimum resolution of 1 meter.
type RIDAircraftPositionAlt float32

// The uncorrected altitude (based on reference standard 29.92 inHg, 1013.25 mb) provides a reference for algorithms that utilize "altitude deltas" between aircraft.  This value is provided in meters and must have a minimum resolution of 1 meter.
type RIDAircraftPositionPressureAltitude float32

// Position of an aircraft as reported for remote ID purposes.
type RIDAircraftPosition struct {
	Lat Latitude `json:"lat"`

	Lng Longitude `json:"lng"`

	// Geodetic altitude (NOT altitude above launch, altitude above ground, or EGM96): aircraft distance above the WGS84 ellipsoid as measured along a line that passes through the aircraft and is normal to the surface of the WGS84 ellipsoid.  This value is provided in meters and must have a minimum resolution of 1 meter.
	Alt RIDAircraftPositionAlt `json:"alt"`

	// Horizontal error that is likely to be present in this reported position.  Required when `extrapolated` field is true and always in the entry for the current state.
	AccuracyH *HorizontalAccuracy `json:"accuracy_h"`

	// Vertical error that is likely to be present in this reported position.  Required when `extrapolated` field is true and always in the entry for the current state.
	AccuracyV *VerticalAccuracy `json:"accuracy_v"`

	// True if this position was generated primarily by computation rather than primarily from a direct instrument measurement.  Assumed false if not specified.
	Extrapolated *bool `json:"extrapolated"`

	// The uncorrected altitude (based on reference standard 29.92 inHg, 1013.25 mb) provides a reference for algorithms that utilize "altitude deltas" between aircraft.  This value is provided in meters and must have a minimum resolution of 1 meter.
	PressureAltitude *RIDAircraftPositionPressureAltitude `json:"pressure_altitude"`
}

// Plain-string representation of a geographic polygon consisting of at least three geographic points describing a closed polygon on the earth.  Each point consists of latitude,longitude in degrees.  Points are also comma-delimited, so this parameter will look like `lat1,lng1,lat2,lng2,lat3,lng3,...`  Latitude values must fall in the range [-90, 90] and longitude values must fall in the range [-180, 180].
//
// All of the requirements and clarifications for GeoPolygon apply to GeoPolygonString as well.
type GeoPolygonString string

// Distance above reference datum.  This value is provided in meters and must have a minimum resolution of 1 meter.
type RIDHeightDistance float32

// A relative altitude for the purposes of remote ID.
type RIDHeight struct {
	// Distance above reference datum.  This value is provided in meters and must have a minimum resolution of 1 meter.
	Distance RIDHeightDistance `json:"distance"`

	// The reference datum above which the height is reported.
	Reference string `json:"reference"`
}

// Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid.
type Latitude float64

// Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid.
type Longitude float64

// Point on the earth's surface.
type LatLngPoint struct {
	Lng Longitude `json:"lng"`

	Lat Latitude `json:"lat"`
}

// An altitude, in meters, above the WGS84 ellipsoid.
type Altitude float32

// Description of a remote ID flight.
type RIDFlight struct {
	Id RIDFlightID `json:"id"`

	// Generic type of aircraft.
	AircraftType RIDAircraftType `json:"aircraft_type"`

	// The most up-to-date state of the aircraft.  Required when the aircraft is currently in the requested area unless `volumes` is specified.
	//
	// If current data is not being received from the UAS by the Service Provider, the lack of change in this field is sufficient to indicate that current data is not being received.
	CurrentState *RIDAircraftState `json:"current_state"`

	// The set of spacetime volumes the aircraft is within.  Required if `current_state` is not specified.  The fields `time_start` and `time_end` are required if `current_state` is not specified.
	Volumes *[]Volume4D `json:"volumes"`

	// If specified as true, this flight is not a physical aircraft; it is just a simulation to test the system.
	Simulated *bool `json:"simulated"`

	// A short collection of recent aircraft movement, specified only when `include_recent_positions` is true.  If `volumes` is not specified and `include_recent_positions` is true, then this field is required.
	//
	// Recent positions provided in this field must conform to requirements in the standard which generally prohibit including positions outside the requested area except transitionally when the aircraft enters or exits the requested area, and which prohibit including positions that not sufficiently recent.
	//
	// Note that a UI should not draw a connective line between two consecutive position reports that both lie outside the requested area.
	RecentPositions *[]RIDRecentAircraftPosition `json:"recent_positions"`
}

// An enclosed area on the earth.
// The bounding edges of this polygon shall be the shortest paths between connected vertices.  This means, for instance, that the edge between two points both defined at a particular latitude is not generally contained at that latitude.
// The winding order shall be interpreted as the order which produces the smaller area.
// The path between two vertices shall be the shortest possible path between those vertices.
// Edges may not cross.
// Vertices may not be duplicated.  In particular, the final polygon vertex shall not be identical to the first vertex.
type GeoPolygon struct {
	Vertices []LatLngPoint `json:"vertices"`
}

// Response to a request to create or update a reference to an Identification Service Area in the DSS.
type PutIdentificationServiceAreaResponse struct {
	// DSS subscribers that this client now has the obligation to notify of the Identification Service Area changes just made.  This client must call POST for each provided URL according to the `/uss/identification_service_areas/{id}` path API.
	Subscribers []SubscriberToNotify `json:"subscribers"`

	// Resulting service area stored in DSS.
	ServiceArea IdentificationServiceArea `json:"service_area"`
}

// Response to DSS query for Identification Service Areas in an area of interest.
type SearchIdentificationServiceAreasResponse struct {
	// Identification Service Areas in the area of interest.
	ServiceAreas []IdentificationServiceArea `json:"service_areas"`
}

// Parameters of a message informing of new full information for an Identification Service Area.  Pushed (by a client, not the DSS) directly to clients with subscriptions when another client makes a change to airspace within a cell with a subscription.
type PutIdentificationServiceAreaNotificationParameters struct {
	// Identification Service Area that the notifying client changed or created.
	//
	// If this field is populated, the Identification Service Area was created or updated.  If this field is not populated, the Identification Service Area was deleted.
	ServiceArea *IdentificationServiceArea `json:"service_area"`

	// Subscription(s) prompting this notification.
	Subscriptions []SubscriptionState `json:"subscriptions"`

	// The new or updated extents of the Identification Service Area.
	//
	// Omitted if Identification Service Area was deleted.
	Extents *Volume4D `json:"extents"`
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
	// Indentification Service Area that was just deleted.
	ServiceArea IdentificationServiceArea `json:"service_area"`

	// DSS subscribers that this client now has the obligation to notify of the Identification Service Area just deleted.  This client must call POST for each provided URL according to the `/uss/identification_service_areas` path API.
	Subscribers []SubscriberToNotify `json:"subscribers"`
}

// The URL at which notifications regarding Identification Service Areas may be delivered.  See the `/uss/identification_service_areas/{id}` path for specification of this endpoint.
type IdentificationServiceAreaURL string

// Endpoints that should be called when an applicable event occurs.  At least one field must be specified.
type SubscriptionCallbacks struct {
	// If specified, other clients will be instructed by the DSS to call this endpoint when an Identification Service Area relevant to this Subscription is created, modified, or deleted.  Must implement PUT and DELETE according to the `/uss/identification_service_areas/{id}` path API.
	IdentificationServiceAreaUrl *IdentificationServiceAreaURL `json:"identification_service_area_url"`
}

// Response for a request to create or update a subscription.
type PutSubscriptionResponse struct {
	// Identification Service Areas in or near the subscription area at the time of creation/update, if `identification_service_area_url` callback was specified.
	ServiceAreas *[]IdentificationServiceArea `json:"service_areas"`

	// Result of the operation on the subscription.
	Subscription Subscription `json:"subscription"`
}

// Indicates operational status of associated aircraft.
type RIDOperationalStatus string

const (
	RIDOperationalStatus_Undeclared RIDOperationalStatus = "Undeclared"
	RIDOperationalStatus_Ground     RIDOperationalStatus = "Ground"
	RIDOperationalStatus_Airborne   RIDOperationalStatus = "Airborne"
)

// Details about a flight reported by a remote ID service provider.  At least one of the registration or serial fields must be filled if required by CAA.
type RIDFlightDetails struct {
	// ID for this flight, matching argument in request.
	Id string `json:"id"`

	// CAA-issued registration/license ID for the remote pilot or operator.
	OperatorId *string `json:"operator_id"`

	// Location of party controlling the aircraft.
	OperatorLocation *LatLngPoint `json:"operator_location"`

	// Free-text field that enables the operator to describe the purpose of a flight, if so desired.
	OperationDescription *string `json:"operation_description"`

	AuthData *RIDAuthData `json:"auth_data"`

	// Can be specified when no registration ID exists and required by law in a region. This is expressed in the ANSI/CTA-2063 Physical Serial Number format.
	SerialNumber *string `json:"serial_number"`

	// If a CAA provides a method of registering UAS, this number is provided by the CAA or its authorized representative.  Required when required by law in a region.
	RegistrationNumber *string `json:"registration_number"`
}

type RIDRecentAircraftPosition struct {
	// Time at which this position applied.  RFC 3339 format, per OpenAPI specification.
	Time string `json:"time"`

	Position RIDAircraftPosition `json:"position"`
}

// Response to DSS request for the identification service area with the given id.
type GetIdentificationServiceAreaResponse struct {
	ServiceArea IdentificationServiceArea `json:"service_area"`
}

// Parameters for a request to create an Identification Service Area in the DSS.
type CreateIdentificationServiceAreaParameters struct {
	// The bounding spacetime extents of this Identification Service Area.  End time must be specified.  If start time is not specified, it will be set to the current time.  Start times in the past should be rejected by the DSS, except that it may adjust very recent start times to the current time.
	//
	// These extents should not reveal any sensitive information about the flight or flights within them.  This means, for instance, that extents should not tightly-wrap a flight path, nor should they generally be centered around the takeoff point of a single flight.
	Extents Volume4D `json:"extents"`

	FlightsUrl RIDFlightsURL `json:"flights_url"`
}

// Parameters for a request to update an Identification Service Area in the DSS.
type UpdateIdentificationServiceAreaParameters struct {
	// The bounding spacetime extents of this Identification Service Area.  End time must be specified.  If start time is not specified, it will remain unchanged.  Start times in the past should be rejected by the DSS unless they are unchanged from the Identification Service Area's current start time.
	//
	// These extents should not reveal any sensitive information about the flight or flights within them.  This means, for instance, that extents should not tightly-wrap a flight path, nor should they generally be centered around the takeoff point of a single flight.
	Extents Volume4D `json:"extents"`

	FlightsUrl RIDFlightsURL `json:"flights_url"`
}

// Parameters for a request to create a subscription in the DSS.
type CreateSubscriptionParameters struct {
	// The spacetime extents of the volume to subscribe to.
	//
	// This subscription will automatically be deleted after its end time if it has not been refreshed by then.  If end time is not specified, the value will be chosen automatically by the DSS.
	//
	// Note that some Entities triggering notifications may lie entirely outside the requested area.
	Extents Volume4D `json:"extents"`

	Callbacks SubscriptionCallbacks `json:"callbacks"`
}

// Parameters for a request to update a subscription in the DSS.
type UpdateSubscriptionParameters struct {
	// The spacetime extents of the volume to subscribe to.
	//
	// This subscription will automatically be deleted after its end time if it has not been refreshed by then.  If end time is not specified, the value will be chosen automatically by the DSS.
	//
	// Note that some Entities triggering notifications may lie entirely outside the requested area.
	Extents Volume4D `json:"extents"`

	Callbacks SubscriptionCallbacks `json:"callbacks"`
}

// Specification of a geographic area that a client is interested in on an ongoing basis (e.g., “planning area”).  Internal to the DSS.
type Subscription struct {
	// Unique identifier for this subscription.
	Id SubscriptionUUID `json:"id"`

	Callbacks SubscriptionCallbacks `json:"callbacks"`

	// Assigned by the DSS based on creating client’s ID (via access token).  Used for restricting mutation and deletion operations to owner.
	Owner string `json:"owner"`

	NotificationIndex SubscriptionNotificationIndex `json:"notification_index"`

	// If set, this subscription will be automatically removed after this time.  RFC 3339 format, per OpenAPI specification.
	TimeEnd *string `json:"time_end"`

	// If set, this Subscription will not generate any notifications before this time.  RFC 3339 format, per OpenAPI specification.
	TimeStart *string `json:"time_start"`

	Version Version `json:"version"`
}

// An Identification Service Area (area in which remote ID services are being provided).  The DSS reports only these declarations and clients must exchange flight information peer-to-peer.
type IdentificationServiceArea struct {
	FlightsUrl RIDFlightsURL `json:"flights_url"`

	// Assigned by the DSS based on creating client’s ID (via access token).  Used for restricting mutation and deletion operations to owner.
	Owner string `json:"owner"`

	// Beginning time of service.  RFC 3339 format, per OpenAPI specification.
	TimeStart string `json:"time_start"`

	// End time of service.  RFC 3339 format, per OpenAPI specification.
	TimeEnd string `json:"time_end"`

	Version Version `json:"version"`

	// Unique identifier for this Identification Service Area.
	Id EntityUUID `json:"id"`
}

// Declaration of timestamp accuracy, which is the largest difference between Timestamp and true time of applicability for any of the following fields: Latitude, Longitude, Geodetic Altitude, Pressure Altitude of Position, Height. to determine time of applicability of the location data provided.  Expressed in seconds, precise to 1/10ths of seconds. The accuracy reflects the 95% uncertainty bound value for the timestamp.
type RIDAircraftStateTimestampAccuracy float32

// Direction of flight expressed as a "True North-based" ground track angle.  This value is provided in degrees East of North with a minimum resolution of 1 degree.
type RIDAircraftStateTrack float32

// Ground speed of flight in meters per second.
type RIDAircraftStateSpeed float32

// Speed up (vertically) WGS84-HAE, m/s.
type RIDAircraftStateVerticalSpeed float32

// Farthest horizontal distance from reported group location at which an aircraft in the group may be located (meters).  This value contains the "Operating Area Radius" data from the common data dictionary when group operation area is specified by point-radius.
type RIDAircraftStateGroupRadius float32

// Maximum altitude (meters WGS84-HAE) of Group Operation.  This value contains the "Operating Area Ceiling" data from the common data dictionary when group operation area is specified by point-radius.
type RIDAircraftStateGroupCeiling float32

// Minimum altitude (meters WGS84-HAE) of Group Operation.  If not specified, ground level shall be assumed.  This value contains the "Operating Area Floor" data from the common data dictionary when group operation area is specified by point-radius.
type RIDAircraftStateGroupFloor float32

// When operating a group (or formation or swarm), number of aircraft in group.  This value contains the "Operating Area Count" data from the common data dictionary when group operation area is specified by point-radius.
type RIDAircraftStateGroupCount int32

// State of an aircraft for the purposes of remote ID.
type RIDAircraftState struct {
	// Time at which this state was valid.  This may be the time coming from the source, such as a GPS, or the time when the system computes the values using an algorithm such as an Extended Kalman Filter (EKF).  Timestamp must be expressed with a minimum resolution of 1/10th of a second.  RFC 3339 format, per OpenAPI specification.
	Timestamp string `json:"timestamp"`

	// Declaration of timestamp accuracy, which is the largest difference between Timestamp and true time of applicability for any of the following fields: Latitude, Longitude, Geodetic Altitude, Pressure Altitude of Position, Height. to determine time of applicability of the location data provided.  Expressed in seconds, precise to 1/10ths of seconds. The accuracy reflects the 95% uncertainty bound value for the timestamp.
	TimestampAccuracy RIDAircraftStateTimestampAccuracy `json:"timestamp_accuracy"`

	OperationalStatus *RIDOperationalStatus `json:"operational_status"`

	Position RIDAircraftPosition `json:"position"`

	// Direction of flight expressed as a "True North-based" ground track angle.  This value is provided in degrees East of North with a minimum resolution of 1 degree.
	Track RIDAircraftStateTrack `json:"track"`

	// Ground speed of flight in meters per second.
	Speed RIDAircraftStateSpeed `json:"speed"`

	// Accuracy of horizontal ground speed.
	SpeedAccuracy SpeedAccuracy `json:"speed_accuracy"`

	// Speed up (vertically) WGS84-HAE, m/s.
	VerticalSpeed RIDAircraftStateVerticalSpeed `json:"vertical_speed"`

	Height *RIDHeight `json:"height"`

	// Farthest horizontal distance from reported group location at which an aircraft in the group may be located (meters).  This value contains the "Operating Area Radius" data from the common data dictionary when group operation area is specified by point-radius.
	GroupRadius *RIDAircraftStateGroupRadius `json:"group_radius"`

	// Maximum altitude (meters WGS84-HAE) of Group Operation.  This value contains the "Operating Area Ceiling" data from the common data dictionary when group operation area is specified by point-radius.
	GroupCeiling *RIDAircraftStateGroupCeiling `json:"group_ceiling"`

	// Minimum altitude (meters WGS84-HAE) of Group Operation.  If not specified, ground level shall be assumed.  This value contains the "Operating Area Floor" data from the common data dictionary when group operation area is specified by point-radius.
	GroupFloor *RIDAircraftStateGroupFloor `json:"group_floor"`

	// When operating a group (or formation or swarm), number of aircraft in group.  This value contains the "Operating Area Count" data from the common data dictionary when group operation area is specified by point-radius.
	GroupCount *RIDAircraftStateGroupCount `json:"group_count"`

	// Time at which a group operation starts.  This value contains the "Operation Area Start" data from the common data dictionary when group operation area is specified by point-radius.
	GroupTimeStart *string `json:"group_time_start"`

	// Time at which a group operation starts.  This value contains the "Operation Area End" data from the common data dictionary when group operation area is specified by point-radius.
	GroupTimeEnd *string `json:"group_time_end"`
}

// Type of aircraft for the purposes of remote ID.
//
// `VTOL` is a fixed wing aircraft that can take off vertically.  `Rotocraft` includes multirotor.
type RIDAircraftType string

const (
	RIDAircraftType_NotDeclared             RIDAircraftType = "NotDeclared"
	RIDAircraftType_Aeroplane               RIDAircraftType = "Aeroplane"
	RIDAircraftType_Rotorcraft              RIDAircraftType = "Rotorcraft"
	RIDAircraftType_Gyroplane               RIDAircraftType = "Gyroplane"
	RIDAircraftType_VTOL                    RIDAircraftType = "VTOL"
	RIDAircraftType_Ornithopter             RIDAircraftType = "Ornithopter"
	RIDAircraftType_Glider                  RIDAircraftType = "Glider"
	RIDAircraftType_Kite                    RIDAircraftType = "Kite"
	RIDAircraftType_FreeBalloon             RIDAircraftType = "FreeBalloon"
	RIDAircraftType_CaptiveBalloon          RIDAircraftType = "CaptiveBalloon"
	RIDAircraftType_Airship                 RIDAircraftType = "Airship"
	RIDAircraftType_FreeFallOrParachute     RIDAircraftType = "FreeFallOrParachute"
	RIDAircraftType_Rocket                  RIDAircraftType = "Rocket"
	RIDAircraftType_TetheredPoweredAircraft RIDAircraftType = "TetheredPoweredAircraft"
	RIDAircraftType_GroundObstacle          RIDAircraftType = "GroundObstacle"
	RIDAircraftType_Other                   RIDAircraftType = "Other"
)

// The URL at which the remote ID flights and their details may be retrieved.  See `/flights` and `/flights/{id}/details` paths for specification of this endpoint.
// This URL is the base flights resource.  If this URL is specified as https://example.com/flights then the flight details for a particular {id} may be obtained at the URL https://example.com/flights/{id}/details.  This URL may not have a trailing / character.
type RIDFlightsURL string
