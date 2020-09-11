from typing import Dict, List

import s2sphere


MAX_SUB_PER_AREA = 10

MAX_SUB_TIME_HRS = 24

DATE_FORMAT = '%Y-%m-%dT%H:%M:%SZ'

SCOPE_READ = 'dss.read.identification_service_areas'
SCOPE_WRITE = 'dss.write.identification_service_areas'


def geo_polygon_string(vertices: List[Dict[str, float]]) -> str:
  return ','.join('{},{}'.format(v['lat'], v['lng']) for v in vertices)


def vertices_from_latlng_rect(rect: s2sphere.LatLngRect) -> List[Dict[str, float]]:
  return [
    {'lat': rect.lat_lo().degrees, 'lng': rect.lng_lo().degrees},
    {'lat': rect.lat_lo().degrees, 'lng': rect.lng_hi().degrees},
    {'lat': rect.lat_hi().degrees, 'lng': rect.lng_hi().degrees},
    {'lat': rect.lat_hi().degrees, 'lng': rect.lng_lo().degrees},
  ]
