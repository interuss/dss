import math
from typing import Tuple
import s2sphere


EARTH_CIRCUMFERENCE_KM = 40075
EARTH_RADIUS_M = 40075 * 1000 / (2 * math.pi)
EARTH_AREA_M2 = 4 * math.pi * math.pow(EARTH_RADIUS_M, 2)


def make_latlng_rect(spec: str) -> s2sphere.LatLngRect:
  coords = spec.split(',')
  if len(coords) != 4:
    raise ValueError('Expected lat,lng,lat,lng; found %d coordinates instead' % len(coords))
  lat1 = _validate_lat(coords[0])
  lng1 = _validate_lng(coords[1])
  lat2 = _validate_lat(coords[2])
  lng2 = _validate_lng(coords[3])
  p1 = s2sphere.LatLng.from_degrees(lat1, lng1)
  p2 = s2sphere.LatLng.from_degrees(lat2, lng2)
  return s2sphere.LatLngRect.from_point_pair(p1, p2)


def _validate_lat(lat: str) -> float:
  lat = float(lat)
  if lat < -90 or lat > 90:
    raise ValueError('Latitude must be in [-90, 90] range')
  return lat


def _validate_lng(lng: str) -> float:
  lng = float(lng)
  if lng < -180 or lng > 180:
    raise ValueError('Longitude must be in [-180, 180] range')
  return lng


def flatten(reference: s2sphere.LatLng, point: s2sphere.LatLng) -> Tuple[float, float]:
  """Locally flatten a lat-lng point to (dx, dy) in meters from reference."""
  return (
    (point.lng().degrees - reference.lng().degrees) * EARTH_CIRCUMFERENCE_KM * math.cos(reference.lat().radians) * 1000 / 360,
    (point.lat().degrees - reference.lat().degrees) * EARTH_CIRCUMFERENCE_KM * 1000 / 360
  )


def unflatten(reference: s2sphere.LatLng, point: Tuple[float, float]) -> s2sphere.LatLng:
  """Locally unflatten a (dx, dy) point to an absolute lat-lng point."""
  return s2sphere.LatLng.from_degrees(
    reference.lat().degrees + point[1] * 360 / (EARTH_CIRCUMFERENCE_KM * 1000 * math.cos(reference.lat().radians)),
    reference.lng().degrees + point[0] * 360 / (EARTH_CIRCUMFERENCE_KM * 1000)
  )


def area_of_latlngrect(rect: s2sphere.LatLngRect) -> float:
  """Compute the approximate surface area within a lat-lng rectangle."""
  return EARTH_AREA_M2 * rect.area() / (4 * math.pi)
