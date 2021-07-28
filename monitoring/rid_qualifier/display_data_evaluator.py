import datetime
import json
import time
from typing import List, Optional, Tuple

import arrow
import s2sphere

from monitoring.monitorlib import fetch, geo, rid
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier.reports import Findings
from monitoring.rid_qualifier.utils import EvaluationConfiguration, InjectedFlight
from monitoring.monitorlib.rid_automated_testing import observation_api
from monitoring.monitorlib.rid_automated_testing.injection_api import TestFlight


class RIDSystemObserver(object):
  def __init__(self, name: str, session: DSSTestSession):
    self.session = session
    self.name = name

  def observe_system(self, rect: s2sphere.LatLngRect) -> Tuple[Optional[observation_api.GetDisplayDataResponse], fetch.Query]:
    initiated_at = datetime.datetime.utcnow()
    resp = self.session.get('/display_data?view={},{},{},{}'.format(
      rect.lo().lat().degrees, rect.lo().lng().degrees,
      rect.hi().lat().degrees, rect.hi().lng().degrees), scope=rid.SCOPE_READ)
    try:
      result = (ImplicitDict.parse(resp.json(),
                                   observation_api.GetDisplayDataResponse)
                if resp.status_code == 200 else None)
    except ValueError as e:
      result = None
    return (result, fetch.describe_query(resp, initiated_at))

  def observe_flight_details(self, flight_id: str) -> Tuple[Optional[observation_api.GetDetailsResponse], fetch.Query]:
    initiated_at = datetime.datetime.utcnow()
    resp = self.session.get('/display_data/{}'.format(flight_id))
    try:
      result = ImplicitDict.parse(resp.json(), observation_api.GetDetailsResponse) if resp.status_code == 200 else None
    except ValueError:
      result = None
    return (result, fetch.describe_query(resp, initiated_at))


def _get_query_rect(injected_flights: List[TestFlight], t: datetime.datetime,
                    config: EvaluationConfiguration) -> s2sphere.LatLngRect:
  data_exists = False
  lat_min = 90
  lng_min = 360
  lat_max = -90
  lng_max = -360

  # Find the bounds of all relevant points
  t_min = t - rid.NetMaxNearRealTimeDataPeriod - config.max_propagation_latency.timedelta
  t_max = t
  for injected_flight in injected_flights:
    for telemetry in injected_flight.telemetry:
      t = arrow.get(telemetry.timestamp).datetime
      if t_min <= t <= t_max:
        data_exists = True
        lat_min = min(lat_min, telemetry.position.lat)
        lat_max = max(lat_max, telemetry.position.lat)
        lng_min = min(lng_min, telemetry.position.lng)
        lng_max = max(lng_max, telemetry.position.lng)

  # If there is no flight data yet, look at the center of where the data will be
  if not data_exists:
    lat = 0
    lng = 0
    n = 0
    for injected_flight in injected_flights:
      for telemetry in injected_flight.telemetry:
        lat += telemetry.position.lat
        lng += telemetry.position.lng
        n += 1
    lat_min = lat_max = lat / n
    lng_min = lng_max = lng / n

  # Expand view size to meet minimum, if necessary
  OVERSHOOT = 1.01
  while True:
    c1 = s2sphere.LatLng.from_degrees(lat_min, lng_min)
    c2 = s2sphere.LatLng.from_degrees(lat_max, lng_max)
    diagonal = c1.get_distance(c2).degrees * geo.EARTH_CIRCUMFERENCE_KM * 1000 / 360
    if diagonal >= config.min_query_diagonal:
      break
    if lat_min == lat_max and lng_min == lng_max:
      lat_min -= 1e-5
      lat_max += 1e-5
      lng_min -= 1e-5
      lng_max += 1e-5
      continue
    lat_center = 0.5 * (lat_min + lat_max)
    lat_span = (lat_max - lat_min) * config.min_query_diagonal / diagonal * OVERSHOOT
    lat_min = lat_center - 0.5 * lat_span
    lat_max = lat_center + 0.5 * lat_span
    lng_center = 0.5 * (lng_min + lng_max)
    lng_span = (lng_max - lng_min) * config.min_query_diagonal / diagonal * OVERSHOOT
    lng_min = lng_center - 0.5 * lng_span
    lng_max = lng_center + 0.5 * lng_span

  p1 = s2sphere.LatLng.from_degrees(lat_min, lng_min)
  p2 = s2sphere.LatLng.from_degrees(lat_max, lng_max)
  return s2sphere.LatLngRect.from_point_pair(p1, p2)


def evaluate_system(
    injected_flights: List[InjectedFlight], observers: List[RIDSystemObserver],
    config: EvaluationConfiguration, findings: Findings) -> None:
  """Evaluate a system by polling system state and comparing to expectations.

  This routine periodically polls each of the specified observers for the system
  state and checks that each system state matches expectations based on the
  provided injected flights, updating the provided report findings.
  """

  # Compute the end of all injected data
  t_end = arrow.utcnow()
  for injected_flight in injected_flights:
    for telemetry in injected_flight.flight.telemetry:
      t = arrow.get(telemetry.timestamp)
      t_end = max(t_end, t)
  t_end += rid.NetMaxNearRealTimeDataPeriod + config.max_propagation_latency.timedelta

  if arrow.utcnow() > t_end:
    raise RuntimeError(
      'Cannot evaluate system: injected test flights ended at {}, which is before now ({})'.format(
        t_end, datetime.datetime.utcnow()))

  
  query_counter = 0  
  last_rect = None

  t_next = arrow.utcnow()

  while arrow.utcnow() < t_end:  
    # Evaluate the system at an instant in time    
  
    t_now = arrow.utcnow().datetime      
    if ((query_counter != 0) and (config.reduce_query_variation)): 
      if (query_counter % 3 == 0):# every third request keep the rectangle as the last one   
        rect = last_rect
    else:      
      rect = _get_query_rect(
        [injected_flight.flight for injected_flight in injected_flights],
        t_now, config)
      last_rect = rect
    _evaluate_system_instantaneously(
      injected_flights, observers, config, findings, rect)
    print('After observation at {}, {}'.format(arrow.utcnow(), findings))
    print(json.dumps(findings.issues, indent=2))

    # Wait until minimum polling interval elapses
    while t_next < arrow.utcnow():
      t_next += config.min_polling_interval.timedelta
    if t_next > t_end:
      break
    delay = t_next - arrow.utcnow()
    if delay.total_seconds() > 0:
      time.sleep(delay.total_seconds())
    query_counter += 1


def _evaluate_system_instantaneously(
    injected_flights: List[InjectedFlight], observers: List[RIDSystemObserver],
    config: EvaluationConfiguration, findings: Findings, rect: s2sphere.LatLngRect) -> None:
  for observer in observers:
    # Conduct an observation, then log and evaluate it
    (observation, query) = observer.observe_system(rect)
    findings.add_observation_query(query)
    _evaluate_observation(injected_flights, observer, rect, observation, query,
                          config, findings)

    #TODO: If bounding rect is smaller than cluster threshold, expand slightly above cluster threshold and re-observe
    #TODO: If bounding rect is smaller than area-too-large threshold, expand slightly above area-too-large threshold and re-observe

def _evaluate_observation(
    injected_flights: List[InjectedFlight], observer: RIDSystemObserver,
    rect: s2sphere.LatLngRect,
    observation: Optional[observation_api.GetDisplayDataResponse],
    query: fetch.Query, config: EvaluationConfiguration,
    findings: Findings) -> None:
  diagonal = rect.lo().get_distance(rect.hi()).degrees * geo.EARTH_CIRCUMFERENCE_KM / 360
  if diagonal > rid.NetMaxDisplayAreaDiagonal:
    _evaluate_area_to_large_observation(observer, diagonal, query, findings)
  elif diagonal > rid.NetDetailsMaxDisplayAreaDiagonal:
    _evaluate_clusters_observation(findings)
  else:
    _evaluate_normal_observation(injected_flights, observer, rect,
                                 observation, query, config, findings)


def _evaluate_normal_observation(
    injected_flights: List[InjectedFlight], observer: RIDSystemObserver,
    rect: s2sphere.LatLngRect,
    observation: Optional[observation_api.GetDisplayDataResponse],
    query: fetch.Query, config: EvaluationConfiguration,
    findings: Findings) -> None:
  if observation is None:
    findings.add_observation_failure(observer.name, rect, query)
    return

  for expected_flight in injected_flights:
    t_initiated = query.request.timestamp
    t_response = query.response.reported
    timestamps = [arrow.get(t.timestamp) for t in expected_flight.flight.telemetry]
    t_min = min(timestamps).datetime
    t_max = max(timestamps).datetime

    flight_id = expected_flight.flight.details_responses[0].details.id # TODO: Choose appropriate details rather than first
    matching_flights = [observed_flight
                        for observed_flight in observation.flights
                        if observed_flight.id == flight_id]
    if len(matching_flights) > 1:
      findings.add_duplicate_flights(observer.name, flight_id, len(matching_flights), expected_flight.uss.name, query)

    if t_response < t_min:
      # This flight should definitely not have been observed (it starts in the future)
      if matching_flights:
        findings.add_premature_flight(observer.name, flight_id, t_min, t_response, expected_flight.uss.name, query)
        continue
    elif t_response > t_max + rid.NetMaxNearRealTimeDataPeriod + config.max_propagation_latency.timedelta:
      # This flight should not have been observed (it was too far in the past)
      if matching_flights:
        findings.add_lingering_flight(observer.name, flight_id, t_max, t_initiated, expected_flight.uss.name, query)
        continue
    elif t_min + config.max_propagation_latency.timedelta < t_initiated < t_max + rid.NetMaxNearRealTimeDataPeriod:
      # This flight should definitely have been observed
      if not matching_flights:
        findings.add_missing_flight(observer.name, expected_flight, rect, expected_flight.uss.name, query)
        continue
    elif t_initiated > t_min:
      # If this flight was not observed, there may be propagation latency
      pass #TODO: findings propagation latency

    for matching_flight in matching_flights:
      pass #TODO: Check position, altitude, flight details, etc


def _evaluate_area_to_large_observation(
  observer: RIDSystemObserver, diagonal: float,
  query: fetch.Query, findings: Findings) -> None:
  if query.status_code != 413:
    findings.add_area_too_large_not_indicated(observer.name, diagonal, query)


def _evaluate_clusters_observation(findings: Findings) -> None:
  #TODO: Check cluster sizing, aircraft counts, etc
  pass
