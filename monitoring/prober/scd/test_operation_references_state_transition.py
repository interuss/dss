"""Operation References state transition tests:
"""

import json

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


OP_TYPE = register_resource_type(8, 'Operational intent')


def test_set_test_owner_ids(test_owner):
  global OP_ID
  OP_ID = utils.encode_owner(test_owner, '00000067-cb83-4880-a7e7-1fee85000000')


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_ensure_clean_workspace_v5(ids, scd_api, scd_session):
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_ensure_clean_workspace_v15(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_accepted_v5(ids, scd_api, scd_session):
  # Accepted for the first time
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_accepted_v15(ids, scd_api, scd_session):
  # Accepted for the first time
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_activated_v5(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Accepted to Activated
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Activated'
    req['old_version'] = 1
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_activated_v15(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operational_intent_reference', None)
  assert existing_op is not None

  # Accepted to Activated
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Activated'
    req['old_version'] = 1
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_accepted_bad1_v5(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  # Activated to Accepted with bad version number 0
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_accepted_bad1_v15(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operational_intent_reference', None)
  assert existing_op is not None

  # Activated to Accepted with bad version number 0
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['key'] = [existing_op["ovn"]]
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_state_transition_v5(ids, scd_api, scd_session):
  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content

  # Create operation with Closed state
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Ended'
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_state_transition_v15(ids, scd_api, scd_session):
  # Delete operation
  resp = scd_session.delete('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content

  # Create operation with Closed state
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Ended'
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content
