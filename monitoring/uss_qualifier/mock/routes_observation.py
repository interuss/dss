import datetime
from typing import List, Optional, Tuple

import flask
import iso8601
import s2sphere

from monitoring.monitorlib import geo, rid
from monitoring.monitorlib.rid_automated_testing import injection_api, observation_api
from monitoring.uss_qualifier.mock import behavior
from monitoring.uss_qualifier.mock.database import db
from . import clustering, webapp


FLIGHT_ACCESSIBLE_DURATION = datetime.timedelta(seconds=65) # after last telemetry


def _make_position_report(
    true_lat: float, true_lng: float, true_alt: float,
    sp_behavior: behavior.ServiceProviderBehavior,
    dp_behavior: behavior.DisplayProviderBehavior) -> observation_api.Position:
  lat = true_lat
  lng = true_lng
  alt = true_alt
  if sp_behavior.switch_latitude_and_longitude_when_reporting:
    temp = lat
    lat = lng
    lng = temp
  return observation_api.Position(lat=lat, lng=lng, alt=alt)


def _make_api_flight(flight: injection_api.TestFlight,
                     sp_behavior: behavior.ServiceProviderBehavior,
                     dp_behavior: behavior.DisplayProviderBehavior,
                     t_earliest: datetime.datetime, t_now: datetime.datetime,
                     lat_min: float, lng_min: float, lat_max: float, lng_max: float) -> observation_api.Flight:
  """Extract the currently-relevant information from a TestFlight.

  :param flight: TestFlight with telemetry for all time
  :param t_earliest: The time before which telemetry should be ignored
  :param t_now: The time after which telemetry should be ignored
  :return: Flight information currently visible in the remote ID system
  """
  paths: List[List[observation_api.Position]] = []
  current_path: List[observation_api.Position] = []
  previous_position: Optional[observation_api.Position] = None
  most_recent_position: Optional[Tuple[datetime.datetime, observation_api.Position]] = None

  for telemetry in flight.telemetry:
    t = iso8601.parse_date(telemetry.timestamp)
    if t < t_earliest:
      # Not relevant; telemetry more than 60s in the past
      continue
    if t > t_now:
      # Not yet relevant; will occur in the future
      continue
    lat = telemetry.position.lat
    lng = telemetry.position.lng
    alt = telemetry.position.alt

    # Mangle data on the Service Provider side
    if sp_behavior.use_agl_instead_of_wgs84_for_altitude:
      if 'height' in telemetry and telemetry.height.reference == 'TakeoffLocation':
        alt = telemetry.height.distance
      else:
        alt -= flight.telemetry[0].position.alt
    if sp_behavior.use_feet_instead_of_meters_for_altitude:
      alt /= 0.3048

    position_report = _make_position_report(lat, lng, alt, sp_behavior, dp_behavior)

    inside_view = lat_min <= lat <= lat_max and lng_min <= lng <= lng_max
    if inside_view:
      # This is a relevant point inside the view
      if not current_path and previous_position:
        # Positions were previously outside the view but this one is in
        current_path.append(previous_position)
      current_path.append(position_report)
      if most_recent_position is None or most_recent_position[0] < t:
        most_recent_position = (t, position_report)
    else:
      # This point is in the relevant time range but outside the view
      if current_path:
        # Positions were previously inside the view but this one is out
        current_path.append(position_report)
        paths.append(current_path)
        current_path = []
    previous_position = position_report
  if current_path:
    paths.append(current_path)

  kwargs = {'id': flight.get_id(t_now)}
  if paths and not dp_behavior.always_omit_recent_paths:
    kwargs['recent_paths'] = [observation_api.Path(positions=p) for p in paths]
  if most_recent_position:
    kwargs['most_recent_position'] = most_recent_position[1]
  return observation_api.Flight(**kwargs)


@webapp.route('/dp/<dp_id>/display_data', methods=['GET'])
def poll_display_data(dp_id: str) -> Tuple[str, int]:
  """Implements display data polling in RID automated testing observation API."""

  # TODO: Validate token signature & scope

  # Retrieve view parameters
  if 'view' not in flask.request.args:
    return 'Missing "view" argument in request', 400

  try:
    coords = [float(v) for v in flask.request.args['view'].split(',')]
  except ValueError as e:
    return '"view" argument not properly formatted: {}'.format(e), 400

  if len(coords) != 4:
    return '"view" argument contains the wrong number of coordinates (expected 4, found {})'.format(len(coords)), 400

  lat_min = min(coords[0], coords[2])
  lat_max = max(coords[0], coords[2])
  lng_min = min(coords[1], coords[3])
  lng_max = max(coords[1], coords[3])

  if (lat_min < -90 or lat_min >= 90 or lat_max <= -90 or lat_max > 90 or
    lng_min < -180 or lng_min >= 360 or lng_max <= -180 or lng_max > 360):
    return '"view" coordinates do not fall within the valid range of -90 <= lat <= 90 and -180 <= lng <= 360', 400

  # Check view size
  view_min = s2sphere.LatLng.from_degrees(lat_min, lng_min)
  view_max = s2sphere.LatLng.from_degrees(lat_max, lng_max)
  diagonal = view_min.get_distance(view_max).degrees * geo.EARTH_CIRCUMFERENCE_KM / 360
  if diagonal > 3.6:
    return flask.jsonify(rid.ErrorResponse(message='Requested diagonal was too large')), 413

  # Get Display Provider behavior
  dp_behavior = db.get_dp(dp_id).behavior

  # Find flights to report
  t_now = datetime.datetime.now(datetime.timezone.utc)
  t_earliest = t_now - datetime.timedelta(seconds=60)
  flights: List[observation_api.Flight] = []
  for sp_id, sp in db.sps.items():
    if sp_id in dp_behavior.do_not_display_flights_from:
      continue
    sp_behavior = db.get_sp(sp_id).behavior
    for test_id, test in sp.tests.items():
      for flight in test.flights:
        flights.append(_make_api_flight(
          flight, sp_behavior, dp_behavior, t_earliest, t_now,
          lat_min, lng_min, lat_max, lng_max))
  flights = [flight for flight in flights if 'most_recent_position' in flight]

  if diagonal <= 1:
    return flask.jsonify(observation_api.GetDisplayDataResponse(flights=flights))
  else:
    return flask.jsonify(observation_api.GetDisplayDataResponse(clusters=clustering.make_clusters(flights, view_min, view_max)))


@webapp.route('/dp/<dp_id>/display_data/<flight_id>', methods=['GET'])
def display_data_details(dp_id: str, flight_id: str) -> Tuple[str, int]:
  """Implements display data details in RID automated testing observation API."""

  # TODO: Validate token signature & scope

  t_now = datetime.datetime.now(datetime.timezone.utc)
  for sp_id, sp in db.sps.items():
    for test_id, test in sp.tests.items():
      for flight in test.flights:
        tf_details = flight.get_details(t_now)
        if tf_details and tf_details.id == flight_id:
          t_max = max(iso8601.parse_date(telemetry.timestamp) for telemetry in flight.telemetry)
          if t_now <= t_max + FLIGHT_ACCESSIBLE_DURATION:
            return flask.jsonify(observation_api.GetDetailsResponse())
          else:
            return 'Flight no longer exists', 404

  return 'Could not find flight with ID of {} at current time'.format(flight_id), 404
