from typing import List, NamedTuple, Optional, Dict
from shapely.geometry import Polygon
import shapely.geometry
from datetime import datetime, timedelta
from monitoring.monitorlib.rid import RIDAircraftState
from monitoring.monitorlib.typing import ImplicitDict


class RIDQualifierUSSConfig(ImplicitDict):
    ''' This object defines the data required for a uss '''

    injection_base_url: str
    injection_suffix: str
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
    timestamp_before: timedelta
    timestamp_after: timedelta
    
class FlightPoint(NamedTuple):
    ''' This object holds basic information about a point on the flight track, it has latitude, longitude and altitude in WGS 1984 datum '''

    lat: float # Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1160
    lng: float # Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1170
    alt: float # meters in WGS 84, normally calculated as height of ground level in WGS84 and altitude above ground level
    


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


class OperatorLocation(NamedTuple):
    ''' A object to hold location of the operator when submitting flight data to USS '''
    lat: float
    lng: float

class RIDFlightDetails(NamedTuple):
    ''' A object to hold RID details of a flight operator that will be reported by the USS as a part of the test ''' 
    operator_id:str
    operation_description: str
    operator_location: OperatorLocation
    serial_number: str
    registration_number: str

class TestFlightDetails(NamedTuple):
    ''' A object to hold the remote ID Details,  and a date time after which the USS should submit the flight details, it matches the TestFlightDetails in the injection interface, for more details see: https://github.com/interuss/dss/blob/master/interfaces/automated-testing/rid/injection.yaml#L158 ''' 
    effective_after: datetime
    details: RIDFlightDetails


class TestFlight(NamedTuple):
    ''' Represents the data necessary to inject a single, complete test flight into a Remote ID Service Provider under test; matches TestFlight in injection interface ''' 

    injection_id: str    
    telemetry: List[RIDAircraftState]
    details_responses : TestFlightDetails   


class TestPayload(ImplicitDict):
    ''' This object defines the detail of a test object, one or more flight tracks may be assigned in a test therefore the requested flights is a list. '''

    test_id: str 
    requested_flights:List[TestFlight]

class DeliverablePayload(ImplicitDict):
    ''' This object defines the payload that needs will be submitted to the Test Inejection URL. The payload is a set of flight tracks, operator details and other associated objects. '''
        
    injection_path: str    
    injection_payload: List[TestPayload]
