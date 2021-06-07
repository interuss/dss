from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

import s2sphere


TIME_FORMAT_CODE = 'RFC3339'
DATE_FORMAT = '%Y-%m-%dT%H:%M:%S.%fZ'
EARTH_CIRCUMFERENCE_M = 40.075e6
SCOPE_SC = 'utm.strategic_coordination'
SCOPE_CM = 'utm.constraint_management'
SCOPE_CI = 'utm.constraint_consumption'


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
