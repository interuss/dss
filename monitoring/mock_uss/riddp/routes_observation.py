from typing import Dict, List, Optional, Tuple

import arrow
import flask
import s2sphere

from monitoring.monitorlib import geo, rid
from monitoring.monitorlib.fetch import rid as fetch
from monitoring.monitorlib.rid_automated_testing import observation_api
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_uss import resources, webapp
from monitoring.mock_uss.auth import requires_scope
from . import database
from .database import db


def _make_flight_observation(flight: rid.RIDFlight, view: s2sphere.LatLngRect) -> observation_api.Flight:
  paths: List[List[observation_api.Position]] = []
  current_path: List[observation_api.Position] = []
  previous_position: Optional[observation_api.Position] = None

  lat_min = view.lat_lo().degrees
  lat_max = view.lat_hi().degrees
  lng_min = view.lng_lo().degrees
  lng_max = view.lng_hi().degrees

  # Decompose recent positions into a list of contiguous paths
  for p in flight.recent_positions:
    lat = p.position.lat
    lng = p.position.lng
    position_report = observation_api.Position(
      lat=lat, lng=lng, alt=p.position.alt)

    inside_view = lat_min <= lat <= lat_max and lng_min <= lng <= lng_max
    if inside_view:
      # This point is inside the view
      if not current_path and previous_position:
        # Positions were previously outside the view but this one is in
        current_path.append(previous_position)
      current_path.append(position_report)
    else:
      # This point is outside the view
      if current_path:
        # Positions were previously inside the view but this one is out
        current_path.append(position_report)
        paths.append(current_path)
        current_path = []
    previous_position = position_report
  if current_path:
    paths.append(current_path)

  return observation_api.Flight(
    id=flight.id,
    most_recent_position=observation_api.Position(
      lat=flight.current_state.position.lat,
      lng=flight.current_state.position.lng,
      alt=flight.current_state.position.alt),
    recent_paths=[observation_api.Path(positions=path) for path in paths])


@webapp.route('/riddp/observation/display_data', methods=['GET'])
@requires_scope([rid.SCOPE_READ])
def display_data() -> Tuple[str, int]:
  """Implements retrieval of current display data per automated testing API."""

  if 'view' not in flask.request.args:
    return flask.jsonify(rid.ErrorResponse(message='Missing required "view" parameter')), 400
  try:
    view = geo.make_latlng_rect(flask.request.args['view'])
  except ValueError as e:
    return flask.jsonify(rid.ErrorResponse(message='Error parsing view: {}'.format(e))), 400

  # Determine what kind of response to produce
  diagonal = geo.get_latlngrect_diagonal_km(view)
  if diagonal > rid.NetMaxDisplayAreaDiagonal:
    return flask.jsonify(rid.ErrorResponse(message='Requested diagonal was too large')), 413

  # Get ISAs in the DSS
  t = arrow.utcnow().datetime
  isas_response: fetch.FetchedISAs = fetch.isas(resources.utm_client, view, t, t)
  if not isas_response.success:
    response = rid.ErrorResponse(message='Unable to fetch ISAs from DSS')
    response['errors'] = [isas_response]
    return flask.jsonify(response), 412

  # Fetch flights from each unique flights URL
  validated_flights: List[rid.RIDFlight] = []
  flight_info: Dict[str, database.FlightInfo] = {k: v for k, v in db.flights.items()}

  for flights_url in isas_response.flight_urls:
    flights_response = fetch.flights(resources.utm_client, flights_url, view, True)
    if not flights_response.success:
      response = rid.ErrorResponse(message='Error querying {}'.format(flights_url))
      response['errors'] = [flights_response]
      return flask.jsonify(response), 412
    for flight in flights_response.flights:
      try:
        validated_flight: rid.RIDFlight = ImplicitDict.parse(flight, rid.RIDFlight)
      except ValueError as e:
        response = rid.ErrorResponse(message='Error parsing flight from {}'.format(flights_url))
        response['parse_error'] = str(e)
        response['flights'] = flights_response.flights
        return flask.jsonify(response), 412
      validated_flights.append(validated_flight)
      flight_info[flight.id] = database.FlightInfo(flights_url=flights_url)

  if diagonal <= rid.NetDetailsMaxDisplayAreaDiagonal:
    # Construct detailed flights response
    response = observation_api.GetDisplayDataResponse(flights=[
      _make_flight_observation(f, view) for f in validated_flights])
    db.flights = flight_info
    return flask.jsonify(response)
  else:
    # Construct clusters response
    return 'Cluster response not yet implemented', 500


@webapp.route('/riddp/observation/display_data/<flight_id>', methods=['GET'])
@requires_scope([rid.SCOPE_READ])
def flight_details(flight_id: str) -> Tuple[str, int]:
  """Implements get flight details endpoint per automated testing API."""

  if flight_id not in db.flights:
    return 'Flight "{}" not found'.format(flight_id), 404

  return flask.jsonify(observation_api.GetDetailsResponse())
