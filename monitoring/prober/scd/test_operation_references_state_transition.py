"""Operation References state transition tests:
"""

import json

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.scd import SCOPE_SC

OP_ID = '00000067-cb83-4880-a7e7-1fee85000000'


def test_ensure_clean_workspace(scd_session):
  resp = scd_session.get('/operation_references/{}'.format(OP_ID), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/operation_references/{}'.format(OP_ID), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_SC)
def test_op_accepted(scd_session):
  # Accepted for the first time
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(OP_ID), json=req)
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_SC)
def test_op_activated(scd_session):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(OP_ID))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Accepted to Activated
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Activated'
    req['old_version'] = 1
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(OP_ID), json=req)
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_SC)
def test_op_accepted_bad1(scd_session):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(OP_ID))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Activated to Accepted with bad version number 0
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(OP_ID), json=req)
  assert resp.status_code == 409, resp.content


@default_scope(SCOPE_SC)
def test_op_bad_state_transition(scd_session):
  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(OP_ID))
  assert resp.status_code == 200, resp.content

  # Create operation with Closed state
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Ended'
  resp = scd_session.put('/operation_references/{}'.format(OP_ID), json=req)
  assert resp.status_code == 400, resp.content
