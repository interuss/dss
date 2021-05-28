from typing import Tuple
import uuid

import flask

from monitoring.monitorlib.rid_automated_testing import injection_api
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_ridsp import webapp
from monitoring.mock_ridsp.auth import requires_scope
from . import database
from .database import db


@webapp.route('/injection/tests/<test_id>', methods=['PUT'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def create_test(test_id: str) -> Tuple[str, int]:
  """Implements test creation in RID automated testing injection API."""

  try:
    json = flask.request.json
    if json is None:
      raise ValueError('Request did not contain a JSON payload')
    req_body: injection_api.CreateTestParameters = ImplicitDict.parse(json, injection_api.CreateTestParameters)
    record = database.TestRecord(version=str(uuid.uuid4()), flights=req_body.requested_flights)
    #TODO: Create ISA in DSS
    db.tests[test_id] = record

    return flask.jsonify(injection_api.ChangeTestResponse(version=record.version, injected_flights=record.flights))
  except ValueError as e:
    msg = 'Create test {} unable to parse JSON: {}'.format(test_id, e)
    return msg, 400


@webapp.route('/injection/tests/<test_id>', methods=['DELETE'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def delete_test(test_id: str) -> Tuple[str, int]:
  """Implements test deletion in RID automated testing injection API."""

  if test_id not in db.tests:
    return 'Test "{}" not found'.format(test_id), 404

  record = db.tests[test_id]
  #TODO: Delete ISA in DSS
  del db.tests[test_id]
  return flask.jsonify(injection_api.ChangeTestResponse(version=record.version, injected_flights=record.flights))
