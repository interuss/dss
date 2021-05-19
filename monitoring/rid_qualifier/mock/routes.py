import datetime
from typing import List, Optional, Tuple
import uuid

import flask
import iso8601
import s2sphere

from monitoring.monitorlib import geo, rid
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier.mock import database
from monitoring.rid_qualifier.mock.database import db
from . import api, clustering, webapp


FLIGHT_ACCESSIBLE_DURATION = datetime.timedelta(seconds=65) # after last telemetry


@webapp.route('/status')
def status():
  return 'RID system mock for rid_qualifier is Ok', 200


@webapp.route('/sp/<sp_id>/tests/<test_id>', methods=['PUT'])
def create_test(sp_id: str, test_id: str) -> Tuple[str, int]:
  """Implements test creation in RID automated testing injection API."""

  # TODO: Validate token signature & scope

  try:
    json = flask.request.json
    if json is None:
      raise ValueError('Request did not contain a JSON payload')
    req_body: api.CreateTestParameters = ImplicitDict.parse(json, api.CreateTestParameters)
    record = database.TestRecord(version=str(uuid.uuid4()), flights=req_body.requested_flights)
    if sp_id not in db.sps:
      db.sps[sp_id] = database.RIDSP()
    db.sps[sp_id].tests[test_id] = record

    return flask.jsonify(api.ChangeTestResponse(version=record.version, injected_flights=record.flights))
  except ValueError as e:
    msg = 'Create test {} for Service Provider {} unable to parse JSON: {}'.format(test_id, sp_id, e)
    return msg, 400


@webapp.route('/sp/<sp_id>/tests/<test_id>', methods=['DELETE'])
def delete_test(sp_id: str, test_id: str) -> Tuple[str, int]:
  """Implements test deletion in RID automated testing injection API."""

  # TODO: Validate token signature & scope

  if sp_id not in db.sps:
    return 'RID Service Provider "{}" not found'.format(sp_id), 404
  if test_id not in db.sps[sp_id].tests:
    return 'Test "{}" not found for RID Service Provider "{}"'.format(test_id, sp_id), 404

  record = db.sps[sp_id].tests[test_id]
  del db.sps[sp_id].tests[test_id]
  return flask.jsonify(api.ChangeTestResponse(version=record.version, injected_flights=record.flights))


def _make_api_flight(flight: api.TestFlight,
                     t_earliest: datetime.datetime, t_now: datetime.datetime,
                     lat_min: float, lng_min: float, lat_max: float, lng_max: float) -> api.Flight:
  """Extract the currently-relevant information from a TestFlight.

  :param flight: TestFlight with telemetry for all time
  :param t_earliest: The time before which telemetry should be ignored
  :param t_now: The time after which telemetry should be ignored
  :return: Flight information currently visible in the remote ID system
  """
  paths: List[List[api.Position]] = []
  current_path: List[api.Position] = []
  previous_position: Optional[api.Position] = None
  most_recent_position: Optional[Tuple[datetime.datetime, api.Position]] = None

  for telemetry in flight.telemetry:
    t = iso8601.parse_date(telemetry.timestamp)
    if t < t_earliest:
      # Not relevant; telemetry more than 60s in the past
      continue
    if t > t_now:
      # Not yet relevant; will occur in the future
      continue

    if (lat_min <= telemetry.position.lat and telemetry.position.lat <= lat_max and
        lng_min <= telemetry.position.lng and telemetry.position.lng <= lng_max):
      # This is a relevant point inside the view
      if not current_path and previous_position:
        # Positions were previously outside the view but this one is in
        current_path.append(previous_position)
      current_path.append(telemetry.position)
      if most_recent_position is None or most_recent_position[0] < t:
        most_recent_position = (t, api.Position(lat=telemetry.position.lat,
                                                lng=telemetry.position.lng,
                                                alt=telemetry.position.alt))
    else:
      # This point is in the relevant time range but outside the view
      if current_path:
        # Positions were previously inside the view but this one is out
        current_path.append(telemetry.position)
        paths.append(current_path)
        current_path = []
    previous_position = api.Position(lat=telemetry.position.lat,
                                     lng=telemetry.position.lng,
                                     alt=telemetry.position.alt)
  if current_path:
    paths.append(current_path)

  kwargs = {'id': flight.get_id(t_now)}
  if paths:
    kwargs['recent_paths'] = [api.Path(positions=p) for p in paths]
  if most_recent_position:
    kwargs['most_recent_position'] = most_recent_position[1]
  return api.Flight(**kwargs)


@webapp.route('/dp/display_data', methods=['GET'])
def poll_display_data() -> Tuple[str, int]:
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

  # Find flights to report
  t_now = datetime.datetime.now(datetime.timezone.utc)
  t_earliest = t_now - datetime.timedelta(seconds=60)
  flights: List[api.Flight] = []
  for sp_id, sp in db.sps.items():
    for test_id, test in sp.tests.items():
      for flight in test.flights:
        flights.append(_make_api_flight(flight, t_earliest, t_now, lat_min, lng_min, lat_max, lng_max))
  flights = [flight for flight in flights if 'recent_paths' in flight]

  if diagonal <= 1:
    return flask.jsonify(api.GetDisplayDataResponse(flights=flights))
  else:
    return flask.jsonify(api.GetDisplayDataResponse(clusters=clustering.make_clusters(flights, view_min, view_max)))


@webapp.route('/dp/display_data/<id>', methods=['GET'])
def display_data_details(id: str) -> Tuple[str, int]:
  """Implements display data details in RID automated testing observation API."""

  # TODO: Validate token signature & scope

  t_now = datetime.datetime.now(datetime.timezone.utc)
  for sp_id, sp in db.sps.items():
    for test_id, test in sp.tests.items():
      for flight in test.flights:
        tf_details = flight.get_details(t_now)
        if tf_details and tf_details.id == id:
          t_max = max(iso8601.parse_date(telemetry.timestamp) for telemetry in flight.telemetry)
          if t_now <= t_max + FLIGHT_ACCESSIBLE_DURATION:
            return flask.jsonify(api.GetDetailsResponse())
          else:
            return 'Flight no longer exists', 404

  return 'Could not find flight with ID of {} at current time'.format(id), 404
