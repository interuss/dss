"""Operation References state transition tests:
"""

import json


def test_op_accepted(scd_session, op1_uuid):
  # Accepted for the first time
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 200, resp.content


def test_op_activated(scd_session, op1_uuid):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Accepted to Activated
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Activated'
    req['old_version'] = 1
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 200, resp.content


def test_op_accepted_bad1(scd_session, op1_uuid):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Activated to Accepted with bad version number 0
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 409, resp.content


def test_op_bad_state_transition(scd_session, op1_uuid):
  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content

  # Create operation with Closed state
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Ended'
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content
