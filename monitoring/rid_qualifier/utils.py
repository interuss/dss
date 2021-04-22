from typing import List, NamedTuple
from shapely.geometry import Polygon
import shapely.geometry
from datetime import datetime, timedelta


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


class AircraftPosition(NamedTuple):
    ''' A object to hold AircraftPosition details for Remote ID purposes, it mataches the RIDAircraftPosition  per the RID standard, for more information see https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1091'''

    lat : float 
    lng : float 
    alt : float
    accuracy_h : str
    accuracy_v : str
    extrapolated : bool
    pressure_altitude : float

class AircraftHeight(NamedTuple):
    ''' A object to hold relative altitude for the purposes of Remote ID. For more information see: https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1142 '''

    distance: float
    reference: str

class AircraftState(NamedTuple):
    ''' A object to hold Aircraft state details for remote ID purposes. For more information see the published standard API specification at https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1604 '''
    
    timestamp: datetime 
    operational_status: str 
    position: AircraftPosition # See the definition above 
    height: AircraftHeight # See the definition above 
    track: float 
    speed: float 
    speed_accuracy: str 
    vertical_speed: float 

class RIDFlight(NamedTuple):
    ''' A object to store details of a remoteID flight ''' 
    id: str # ID of the flight for Remote ID purposes, e.g. uss1.JA6kHYCcByQ-6AfU, we for this simulation we use just numeric : https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L943
    aircraft_type: str  # Generic type of aircraft https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1711
    current_state: AircraftState # See above for definition


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

class Operator(NamedTuple):
    ''' A object to hold details of a operator while querying Remote ID for testing purposes, contains details of operator assigned to the USS '''
    id: str
    location: OperatorLocation
    operation_description: str
    serial_number: str
    registration_number: str


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
    telemetry: List[AircraftState]
    details_responses : List[TestFlightDetails]    

