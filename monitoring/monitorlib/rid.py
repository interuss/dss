from typing import Dict, List, Optional, NamedTuple, Any
import s2sphere
from datetime import datetime

MAX_SUB_PER_AREA = 10

MAX_SUB_TIME_HRS = 24

DATE_FORMAT = '%Y-%m-%dT%H:%M:%SZ'

SCOPE_READ = 'dss.read.identification_service_areas'
SCOPE_WRITE = 'dss.write.identification_service_areas'

# This scope is used only for experimentation during UPP2
UPP2_SCOPE_ENHANCED_DETAILS = 'rid.read.enhanced_details'


def geo_polygon_string(vertices: List[Dict[str, float]]) -> str:
  return ','.join('{},{}'.format(v['lat'], v['lng']) for v in vertices)


def vertices_from_latlng_rect(rect: s2sphere.LatLngRect) -> List[Dict[str, float]]:
  return [
    {'lat': rect.lat_lo().degrees, 'lng': rect.lng_lo().degrees},
    {'lat': rect.lat_lo().degrees, 'lng': rect.lng_hi().degrees},
    {'lat': rect.lat_hi().degrees, 'lng': rect.lng_hi().degrees},
    {'lat': rect.lat_hi().degrees, 'lng': rect.lng_lo().degrees},
  ]


class ISA(dict):
  @property
  def errors(self) -> List[str]:
    errors: List[str] = []
    if 'flights_url' not in self:
      errors.append('flights_url field missing')
    return errors

  @property
  def id(self) -> Optional[str]:
    return self.get('id', None)

  @property
  def owner(self) -> Optional[str]:
    return self.get('owner', None)

  @property
  def flights_url(self) -> Optional[str]:
    return self.get('flights_url', None)


class Flight(dict):
  @property
  def valid(self) -> bool:
    if self.id is None:
      return False
    return True

  @property
  def id(self) -> str:
    return self.get('id', None)


class FlightDetails(dict):
  pass


class Subscription(dict):
  @property
  def valid(self) -> bool:
    if self.version is None:
      return False
    return True

  @property
  def version(self) -> Optional[str]:
    return self.get('version', None)

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
