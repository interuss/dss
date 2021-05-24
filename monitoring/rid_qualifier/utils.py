from typing import List, NamedTuple
from shapely.geometry import Polygon
import shapely.geometry
from datetime import datetime
from monitoring.monitorlib.rid import RIDAircraftState
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier import injection_api
from monitoring.rid_qualifier.injection_api import OperatorLocation


class RIDQualifierUSSConfig(ImplicitDict):
    ''' This object defines the data required for a uss '''

    injection_base_url: str
    allocated_flight_track_number: int


class RIDQualifierTestConfiguration(ImplicitDict):
    ''' This is the object that defines the test configuration for a RID Qualifier '''

    locale: str  # The locale here is indicating the geographical location in ISO3166 3-letter country code and also a folder within the test definitions directory. The aircraft_state_replayer reads flight track information from the locale/aircraft_states directory.  The locale directory also contains information about the query_bboxes that the rid display provider will use to query and retrieve the flight information.
    now: str
    test_start_time: str
    auth_spec: str
    usses: List[RIDQualifierUSSConfig]


class QueryBoundingBox(NamedTuple):
    ''' This is the object that stores details of query bounding box '''

    name: str
    shape: Polygon
    timestamp_before: datetime
    timestamp_after: datetime


class FlightPoint(NamedTuple):
    ''' This object holds basic information about a point on the flight track, it has latitude, longitude and altitude in WGS 1984 datum '''

    lat: float  # Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1160
    lng: float  # Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1170
    alt: float  # meters in WGS 84, normally calculated as height of ground level in WGS84 and altitude above ground level
    speed: float # speed in m / s
    bearing: float # forward azimuth for the this and the next point on the track


class GridCellFlight(NamedTuple):
    ''' A object to hold details of a grid location and the track within it '''
    bounds: shapely.geometry.polygon.Polygon
    track: List[FlightPoint]


class RIDSP(NamedTuple):

    ''' This is the object that stores details of a USS, mainly it will hold the injection endpoint and details of the flights allocated to the USS and their submissiion status '''

    test_id: str
    name: str
    flight_id: int
    rid_state_injection_url: str
    rid_state_submission_status: bool


class GeneratedFlightDetails(ImplicitDict):
    ''' This object stores the metadata associated with generated flight, this data is shared as information in the remote id call '''
    serial_number: str
    operation_description: str


class GeneratedOperatorDetails(ImplicitDict):
    ''' This object stores the details of the operator that will be transmitted in the RID output. The registration number and operator id can be customized to meet local regulations, at the moment it generates a random alpha numeric string '''

    operator_id: str
    name: str
    registration_number: str
    location: OperatorLocation


class FlightOperatorDetails(ImplicitDict):
    ''' This object stores the "automatically" generated data about operator and the fight '''
    flight_details: GeneratedFlightDetails
    operator_details: GeneratedOperatorDetails


class RIDFlight(ImplicitDict):
    ''' A object to stored details of a remoteID flight '''
    id: str  # ID of the flight for Remote ID purposes, e.g. uss1.JA6kHYCcByQ-6AfU, we for this simulation we use just numeric : https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L943
    aircraft_type: str  # Generic type of aircraft https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1711
    states : List[RIDAircraftState]  # See above for definition


class TestPayload(ImplicitDict):
    ''' This object defines the detail of a test object, one or more flight tracks may be assigned in a test therefore the requested flights is a list. '''

    test_id: str
    requested_flights:List[injection_api.TestFlight]


class DeliverablePayload(ImplicitDict):
    ''' This object defines the payload that needs will be submitted to the Test Inejection URL. The payload is a set of flight tracks, operator details and other associated objects. '''

    injection_path: str
    injection_payload: TestPayload


#TODO: Remove this class; it doesn't make sense to associate many simultaneous flights with the same single operator and flight details
class FlightTelemetryDetails(ImplicitDict):
    ''' Store the Aircraft states, and other metadata before saving on disk '''
    telemetry_data_list: List[RIDFlight]
    reference_time: str
    operator_details: GeneratedOperatorDetails
    flight_details: GeneratedFlightDetails


class FullFlightRecord(ImplicitDict):
    reference_time: str
    flight_telemetry: RIDFlight
    flight_details: GeneratedFlightDetails
    operator_details: GeneratedOperatorDetails
