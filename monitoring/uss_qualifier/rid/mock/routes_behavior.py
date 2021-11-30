from typing import Tuple

import flask

from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.mock import behavior
from monitoring.uss_qualifier.rid.mock.database import db
from . import webapp


@webapp.route('/dp/<dp_id>/behavior', methods=['PUT'])
def set_dp_behavior(dp_id: str) -> Tuple[str, int]:
  """Set the behavior of the specified virtual Display Provider."""
  try:
    json = flask.request.json
    if json is None:
      raise ValueError('Request did not contain a JSON payload')
    dp_behavior = ImplicitDict.parse(json, behavior.DisplayProviderBehavior)
  except ValueError as e:
    msg = 'Change behavior for Display Provider {} unable to parse JSON: {}'.format(dp_id, e)
    return msg, 400

  dp = db.get_dp(dp_id)
  dp.behavior = dp_behavior

  return flask.jsonify(dp.behavior)


@webapp.route('/sp/<sp_id>/behavior', methods=['PUT'])
def set_sp_behavior(sp_id: str) -> Tuple[str, int]:
  """Set the behavior of the specified virtual Service Provider."""
  try:
    json = flask.request.json
    if json is None:
      raise ValueError('Request did not contain a JSON payload')
    sp_behavior = ImplicitDict.parse(json, behavior.ServiceProviderBehavior)
  except ValueError as e:
    msg = 'Change behavior for Service Provider {} unable to parse JSON: {}'.format(sp_id, e)
    return msg, 400

  sp = db.get_sp(sp_id)
  sp.behavior = sp_behavior

  return flask.jsonify(sp.behavior)
