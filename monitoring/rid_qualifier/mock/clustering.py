import random
from typing import List

import s2sphere

from monitoring.monitorlib import geo
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier.mock import api


class Point(object):
  x: float
  y: float

  def __init__(self, x: float, y: float):
    self.x = x
    self.y = y


class Cluster(ImplicitDict):
  x_min: float
  x_max: float
  y_min: float
  y_max: float
  points: List[Point]

  def randomize(self):
    u_min = min(p.x for p in self.points)
    v_min = min(p.y for p in self.points)
    u_max = max(p.x for p in self.points)
    v_max = max(p.y for p in self.points)
    return Cluster(
      x_min=self.x_min + (u_min - self.x_min) * random.random(),
      y_min=self.y_min + (v_min - self.y_min) * random.random(),
      x_max=u_max + (self.x_max - u_max) * random.random(),
      y_max=v_max + (self.y_max - v_max) * random.random(),
      points=self.points)


def make_clusters(flights: List[api.Flight], view_min: s2sphere.LatLng, view_max: s2sphere.LatLng) -> List[api.Cluster]:
  if not flights:
    return []

  # Make the initial cluster
  points: List[Point] = [
    Point(*geo.flatten(view_min, s2sphere.LatLng.from_degrees(flight.most_recent_position.lat, flight.most_recent_position.lng)))
    for flight in flights]
  x_max, y_max = geo.flatten(view_min, view_max)
  clusters: List[Cluster] = [Cluster(x_min=0, y_min=0, x_max=x_max, y_max=y_max, points=points)]

  # TODO: subdivide cluster into many clusters

  result: List[api.Cluster] = []
  for cluster in clusters:
    cluster = cluster.randomize()
    p1 = geo.unflatten(view_min, (cluster.x_min, cluster.y_min))
    p2 = geo.unflatten(view_min, (cluster.x_max, cluster.y_max))
    result.append(api.Cluster(
      corners=[api.Position(lat=p1.lat().degrees, lng=p1.lng().degrees), api.Position(lat=p2.lat().degrees, lng=p2.lng().degrees)],
      area_sqm=geo.area_of_latlngrect(s2sphere.LatLngRect(p1, p2)),
      number_of_flights=len(cluster.points)
    ))

  return result
