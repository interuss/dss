"""Operation References state transition tests:
"""

import json

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type
from monitoring.prober.scd import actions


OP_TYPE = register_resource_type(8, 'Operational intent')


@for_api_versions(scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session):
    actions.delete_operation_if_exists(ids(OP_TYPE), scd_session, scd_api)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_accepted(ids, scd_api, scd_session):
  # Accepted for the first time
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 201, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_activated(ids, scd_api, scd_session):
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
  resp = scd_session.put('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), existing_op["ovn"]), json=req)
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_accepted_bad1(ids, scd_api, scd_session):
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


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_state_transition(ids, scd_api, scd_session):

  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  ovn = resp.json().get('operational_intent_reference', {}).get('ovn', None)
  # Delete operation
  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn))
  assert resp.status_code == 200, resp.content

  # Create operation with Closed state
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
    req['state'] = 'Ended'
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
def test_final_cleanup(ids, scd_api, scd_session):
    test_ensure_clean_workspace(ids, scd_api, scd_session)
