from typing import Tuple
import uuid

import flask

from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier.mock import database
from monitoring.rid_qualifier.mock.database import db
from . import api, webapp


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
