from typing import List, NamedTuple
from shapely.geometry import Polygon
import shapely.geometry
from datetime import datetime
from monitoring.monitorlib.rid_automated_testing import injection_api
from monitoring.monitorlib.rid import RIDAircraftState, RIDFlightDetails
from monitoring.monitorlib.typing import ImplicitDict, StringBasedTimeDelta


class InjectionTargetConfiguration(ImplicitDict):
    ''' This object defines the data required for a uss '''
    name: str
    injection_base_url: str


class ObserverConfiguration(ImplicitDict):
    name: str
    observation_base_url: str


class EvaluationConfiguration(ImplicitDict):
    min_polling_interval: StringBasedTimeDelta = StringBasedTimeDelta('5s')
    """Do not repeat system observations with intervals smaller than this."""

    max_propagation_latency: StringBasedTimeDelta = StringBasedTimeDelta('10s')
    """Allow up to this much time for data to propagate through the system."""

    min_query_diagonal: float = 100
    """Do not make queries with diagonals smaller than this many meters."""

    repeat_query_rect_period: int = 3
    """If set to a value above zero, reuse the most recent query rectangle/view every this many queries."""



class RIDQualifierTestConfiguration(ImplicitDict):
    injection_targets: List[InjectionTargetConfiguration]
    """Set of Service Providers into which data should be injected"""

    observers: List[ObserverConfiguration]
    """Set of Display Providers through with the system should be observed"""

    flight_start_delay: StringBasedTimeDelta = StringBasedTimeDelta('15s')
    """Amount of time between starting the test and commencement of flights"""

    evaluation: EvaluationConfiguration = EvaluationConfiguration()
    """Settings to control behavior when evaluating observed system data"""


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


class FlightDetails(ImplicitDict):
    ''' This object stores the metadata associated with generated flight, this data is shared as information in the remote id call '''
    rid_details: RIDFlightDetails
    operator_name: str
    aircraft_type: str  # Generic type of aircraft https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1711


class FullFlightRecord(ImplicitDict):
    reference_time: str
    states: List[RIDAircraftState]
    flight_details: FlightDetails


class InjectedFlight(ImplicitDict):
    uss: InjectionTargetConfiguration
    flight: injection_api.TestFlight
