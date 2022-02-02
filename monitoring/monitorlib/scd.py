from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple, Literal
from .typing import ImplicitDict, StringBasedDateTime

import s2sphere


TIME_FORMAT_CODE = 'RFC3339'
DATE_FORMAT = '%Y-%m-%dT%H:%M:%S.%fZ'
EARTH_CIRCUMFERENCE_M = 40.075e6

API_0_3_5 = '0.3.5'
API_0_3_17 = '0.3.17'
# In Both
SCOPE_SC = 'utm.strategic_coordination'
SCOPE_CM = 'utm.constraint_management'

# In 0.3.5
SCOPE_CI = 'utm.constraint_consumption'

# In 0.3.17
SCOPE_CP = 'utm.constraint_processing'
SCOPE_CM_SA = 'utm.conformance_monitoring_sa'
SCOPE_AA = 'utm.availability_arbitration'

NO_OVN_PHRASES = {'', 'Available from USS'}


def make_vol4(
    t0: Optional[datetime] = None,
    t1: Optional[datetime] = None,
    alt0: Optional[float] = None,
    alt1: Optional[float] = None,
    circle: Dict = None,
    polygon: Dict = None) -> Dict:
  vol3 = dict()
  if circle is not None:
    vol3['outline_circle'] = circle
  if polygon is not None:
    vol3['outline_polygon'] = polygon
  if alt0 is not None:
    vol3['altitude_lower'] = make_altitude(alt0)
  if alt1 is not None:
    vol3['altitude_upper'] = make_altitude(alt1)
  vol4 = {'volume': vol3}
  if t0 is not None:
    vol4['time_start'] = make_time(t0)
  if t1 is not None:
    vol4['time_end'] = make_time(t1)
  return vol4


def make_time(t: datetime) -> Dict:
  return {
    'value': t.isoformat() + 'Z',
    'format': 'RFC3339'
  }


def make_altitude(alt: float) -> Dict:
  return {
    'value': alt,
    'reference': 'W84',
    'units': 'M'
  }


def make_circle(lat: float, lng: float, radius: float) -> Dict:
  return {
    "center": {
      "lat": lat,
      "lng": lng,
    },
    "radius": {
      "value": radius,
      "units": "M"
    }
  }


def make_polygon(coords: List[Tuple[float, float]]=None, latlngrect: s2sphere.LatLngRect=None) -> Dict:
  if coords is not None:
    return {
      "vertices": [ {'lat': lat, 'lng': lng} for (lat, lng) in coords]
    }

  return {
    "vertices": [
      {'lat': latlngrect.lat_lo().degrees, 'lng': latlngrect.lng_lo().degrees},
      {'lat': latlngrect.lat_lo().degrees, 'lng': latlngrect.lng_hi().degrees},
      {'lat': latlngrect.lat_hi().degrees, 'lng': latlngrect.lng_hi().degrees},
      {'lat': latlngrect.lat_hi().degrees, 'lng': latlngrect.lng_lo().degrees},
    ]
  }


def latitude_degrees(distance_meters: float) -> float:
  return 360 * distance_meters / EARTH_CIRCUMFERENCE_M


def parse_time(time: Dict) -> datetime:
  t_str = time['value']
  if t_str[-1] == 'Z':
    t_str = t_str[0:-1]
  return datetime.fromisoformat(t_str)


def start_of(vol4s: List[Dict]) -> datetime:
  return min([parse_time(vol4['time_start']) for vol4 in vol4s])


def offset_time(vol4s: List[Dict], dt: timedelta) -> List[Dict]:
  for vol4 in vol4s:
    vol4['time_start'] = make_time(parse_time(vol4['time_start']) + dt)
    vol4['time_end'] = make_time(parse_time(vol4['time_end']) + dt)
  return vol4s


class Subscription(dict):
  @property
  def valid(self) -> bool:
    if self.version is None:
      return False
    return True

  @property
  def version(self) -> Optional[int]:
    return self.get('version', None)


################################################################################
#################### Start of ASTM-standard definitions    #####################
#################### interfaces/astm-utm/Protocol/utm.yaml #####################
################################################################################

class LatLngPoint(ImplicitDict):
    '''A class to hold information about a location as Latitude / Longitude pair '''
    lat: float
    lng: float

class Radius(ImplicitDict):
    ''' A class to hold the radius of a circle for the outline_circle object '''
    value: float
    units: str

class Polygon(ImplicitDict):
    ''' A class to hold the polygon object, used in the outline_polygon of the Volume3D object '''
    vertices: List[LatLngPoint] # A minimum of three LatLngPoints are required

class Circle(ImplicitDict):
    ''' A class the details of a circle object used in the outline_circle object '''
    center: LatLngPoint 
    radius: Radius

class Altitude(ImplicitDict):
    ''' A class to hold altitude information '''
    value:float
    reference:Literal['W84']
    units: str 

class Time(ImplicitDict):
    ''' A class to hold Time details '''
    value: StringBasedDateTime 
    format:Literal['RFC3339'] 

class Volume3D(ImplicitDict):
    '''A class to hold Volume3D objects '''
    outline_circle: Optional[Circle]
    outline_polygon: Optional[Polygon]
    altitude_lower: Altitude
    altitude_upper: Altitude

class Volume4D(ImplicitDict):
    '''A class to hold Volume4D objects '''
    volume: Volume3D
    time_start: Time
    time_end: Time

################################################################################
#################### End of ASTM-standard definitions    #####################
#################### interfaces/astm-utm/Protocol/utm.yaml #####################
################################################################################