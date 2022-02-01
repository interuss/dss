import datetime
from typing import Tuple
import uuid

import flask

from monitoring.monitorlib import rid
from monitoring.monitorlib.mutate import rid as mutate
from monitoring.monitorlib.rid_automated_testing import injection_api
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss import config, resources
from . import database
from .database import db


# Time after the last position report during which the created ISA will still
# exist.  This value must be at least 60 seconds per NET0610.
RECENT_POSITIONS_BUFFER = datetime.timedelta(seconds=60.2)


@webapp.route('/ridsp/injection/tests/<test_id>', methods=['PUT'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def create_test(test_id: str) -> Tuple[str, int]:
  """Implements test creation in RID automated testing injection API."""

  try:
    json = flask.request.json
    if json is None:
      raise ValueError('Request did not contain a JSON payload')
    req_body: injection_api.CreateTestParameters = ImplicitDict.parse(json, injection_api.CreateTestParameters)
    record = database.TestRecord(version=str(uuid.uuid4()), flights=req_body.requested_flights)
  except ValueError as e:
    msg = 'Create test {} unable to parse JSON: {}'.format(test_id, e)
    return msg, 400

  # Create ISA in DSS
  (t0, t1) = req_body.get_span()
  t1 += RECENT_POSITIONS_BUFFER
  rect = req_body.get_rect()
  flights_url = '{}/mock/ridsp/v1/uss/flights'.format(webapp.config.get(config.KEY_BASE_URL))
  mutated_isa = mutate.put_isa(resources.utm_client, rect, t0, t1, flights_url, record.version)
  if not mutated_isa.dss_response.success:
    response = rid.ErrorResponse(message='Unable to create ISA in DSS')
    response['errors'] = mutated_isa.dss_response.errors
    return flask.jsonify(response), 412
  record.isa_version = mutated_isa.dss_response.isa.version
  for (url, notification) in mutated_isa.notifications.items():
    code = notification.response.status_code
    if code != 204 and code != 200:
      pass #TODO: Log notification failures (maybe also log incorrect 200s)

  db.tests[test_id] = record
  return flask.jsonify(injection_api.ChangeTestResponse(version=record.version, injected_flights=record.flights))


@webapp.route('/ridsp/injection/tests/<test_id>', methods=['DELETE'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def delete_test(test_id: str) -> Tuple[str, int]:
  """Implements test deletion in RID automated testing injection API."""

  if test_id not in db.tests:
    return 'Test "{}" not found'.format(test_id), 404

  record = db.tests[test_id]

  # Delete ISA from DSS
  deleted_isa = mutate.delete_isa(resources.utm_client, record.version, record.isa_version)
  if not deleted_isa.dss_response.success:
    response = rid.ErrorResponse(message='Unable to delete ISA from DSS')
    response['errors'] = deleted_isa.dss_response.errors
    return flask.jsonify(response), 412
  for (url, notification) in deleted_isa.notifications.items():
    code = notification.response.status_code
    if code != 204 and code != 200:
      pass #TODO: Log notification failures (maybe also log incorrect 200s)

  del db.tests[test_id]
  return flask.jsonify(injection_api.ChangeTestResponse(version=record.version, injected_flights=record.flights))
