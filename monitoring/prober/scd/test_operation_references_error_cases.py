"""Operation References corner cases error tests:
"""

import datetime
import json
import uuid

import yaml

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


OP_TYPE = register_resource_type(6, 'Primary operational intent')
OP_TYPE2 = register_resource_type(7, 'Conflicting operational intent')


@for_api_versions(scd.API_0_3_5)
def test_ensure_clean_workspace_v5(ids, scd_api, scd_session):
  for op_id in (ids(OP_TYPE), ids(OP_TYPE2)):
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


@for_api_versions(scd.API_0_3_17)
def test_ensure_clean_workspace_v15(ids, scd_api, scd_session):
  for op_id in (ids(OP_TYPE), ids(OP_TYPE2)):
    resp = scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_ref_area_too_large_v5(scd_api, scd_session):
  with open('./scd/resources/op_ref_area_too_large.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_ref_area_too_large_v15(scd_api, scd_session):
  with open('./scd/resources/op_ref_area_too_large_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_ref_start_end_times_past_v5(scd_api, scd_session):
  with open('./scd/resources/op_ref_start_end_times_past.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  # It is ok (and useful) to query for past Operations that may not yet have
  # been explicitly deleted.  This is unlike remote ID where ISAs are
  # auto-removed from the perspective of the client immediately after their end
  # time.
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_ref_start_end_times_past_v15(scd_api, scd_session):
  with open('./scd/resources/op_ref_start_end_times_past_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  # It is ok (and useful) to query for past Operations that may not yet have
  # been explicitly deleted.  This is unlike remote ID where ISAs are
  # auto-removed from the perspective of the client immediately after their end
  # time.
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_ref_incorrect_units(scd_api, scd_session):
  with open('./scd/resources/op_ref_incorrect_units.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_ref_incorrect_units_v15(scd_api, scd_session):
  with open('./scd/resources/op_ref_incorrect_units_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_ref_incorrect_altitude_ref_v5(scd_api, scd_session):
  with open('./scd/resources/op_ref_incorrect_altitude_ref.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content

@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_ref_incorrect_altitude_ref_v15(scd_api, scd_session):
  with open('./scd/resources/op_ref_incorrect_altitude_ref_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_uss_base_url_non_tls_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_uss_base_url_non_tls.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_uss_base_url_non_tls_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_uss_base_url_non_tls_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_subscription_id_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_subscription.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_subscription_id_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_subscription_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_subscription_id_random_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_subscription.json', 'r') as f:
    req = json.load(f)
    req['subscription_id'] = uuid.uuid4().hex
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_subscription_id_random_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_subscription_v15.json', 'r') as f:
    req = json.load(f)
    req['subscription_id'] = uuid.uuid4().hex
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_new_and_existing_subscription_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_new_and_existing_subscription.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_new_and_existing_subscription_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_new_and_existing_subscription_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_end_time_past_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_end_time_past.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_end_time_past_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_end_time_past_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_already_exists_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content

  # Verify deletion
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_already_exists_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operational_intent_reference/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content

  # Verify deletion
  resp = scd_session.get('/operational_intent_reference/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_404_version1_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_404_version1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_404_version1_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_404_version1_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_state_version0_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_state_version0.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_state_version0_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_state_version0_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_lat_lon_range_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_lat_lon_range.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_lat_lon_range_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_lat_lon_range_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_area_too_large_put_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_area_too_large_put.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_area_too_large_put_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_area_too_large_put_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_bad_time_format_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_time_format.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_bad_time_format_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_bad_time_format_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_repeated_requests_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_repeated_requests_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operational_intent_reference/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_invalid_id_v5(scd_api, scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/not_uuid_format', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_invalid_id_v15(scd_api, scd_session):
  with open('./scd/resources/op_request_1_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_reference/not_uuid_format', json=req)
  assert resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_missing_conflicted_operation_v5(ids, scd_api, scd_session):
  # Emplace the initial version of Operation 1
  with open('./scd/resources/op_missing_initial.yaml', 'r') as f:
    req = yaml.full_load(f)
  dt = datetime.datetime.utcnow() - scd.start_of(req['extents'])
  req['extents'] = scd.offset_time(req['extents'], dt)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content
  ovn1a = resp.json()['operation_reference']['ovn']
  sub_id = resp.json()['operation_reference']['subscription_id']

  # Emplace the pre-existing Operation that conflicted in the original observation
  with open('./scd/resources/op_missing_preexisting_unknown.yaml', 'r') as f:
    req = yaml.full_load(f)
  req['extents'] = scd.offset_time(req['extents'], dt)
  req['key'] = [ovn1a]
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE2)), json=req)
  assert resp.status_code == 200, resp.content

  # Attempt to update Operation 1 without OVN for the pre-existing Operation
  with open('./scd/resources/op_missing_update.json', 'r') as f:
    req = json.load(f)
  req['extents'] = scd.offset_time(req['extents'], dt)
  req['key'] = [ovn1a]
  req['subscription_id'] = sub_id
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content
  conflicts = []
  for conflict in resp.json()['entity_conflicts']:
    if conflict.get('operation_reference', None):
      conflicts.append(conflict['operation_reference']['id'])
  assert ids(OP_TYPE2) in conflicts, resp.content

  # Perform an area-based query on the area occupied by Operation 1
  with open('./scd/resources/op_missing_query.json', 'r') as f:
    req = json.load(f)
  req['area_of_interest'] = scd.offset_time([req['area_of_interest']], dt)[0]
  resp = scd_session.post('/operation_references/query', json=req)
  assert  resp.status_code == 200, resp.content
  ops = [op['id'] for op in resp.json()['operation_references']]
  assert ids(OP_TYPE) in ops, resp.content

  # ids(OP_ID2) not expected here because its ceiling is <575m whereas query floor is
  # >591m.
  assert ids(OP_TYPE2) not in ops, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_missing_conflicted_operation_v15(ids, scd_api, scd_session):
  # Emplace the initial version of Operation 1
  with open('./scd/resources/op_missing_initial.yaml', 'r') as f:
    req = yaml.full_load(f)
  dt = datetime.datetime.utcnow() - scd.start_of(req['extents'])
  req['extents'] = scd.offset_time(req['extents'], dt)
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 200, resp.content
  ovn1a = resp.json()['operational_intent_reference']['ovn']
  sub_id = resp.json()['operational_intent_reference']['subscription_id']

  # Emplace the pre-existing Operation that conflicted in the original observation
  with open('./scd/resources/op_missing_preexisting_unknown.yaml', 'r') as f:
    req = yaml.full_load(f)
  req['extents'] = scd.offset_time(req['extents'], dt)
  req['key'] = [ovn1a]
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE2)), json=req)
  assert resp.status_code == 200, resp.content

  # Attempt to update Operation 1 without OVN for the pre-existing Operation
  with open('./scd/resources/op_missing_update.json', 'r') as f:
    req = json.load(f)
  req['extents'] = scd.offset_time(req['extents'], dt)
  req['key'] = [ovn1a]
  req['subscription_id'] = sub_id
  resp = scd_session.put('/operational_intent_reference/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 409, resp.content
  # TODO: entity_conflicts is not there in v15 response. What is the replacement key?
  # conflicts = []
  # for conflict in resp.json()['entity_conflicts']:
  #   if conflict.get('operation_reference', None):
  #     conflicts.append(conflict['operation_reference']['id'])
  # assert ids(OP_ID2) in conflicts, resp.content

  # Perform an area-based query on the area occupied by Operation 1
  with open('./scd/resources/op_missing_query.json', 'r') as f:
    req = json.load(f)
  req['area_of_interest'] = scd.offset_time([req['area_of_interest']], dt)[0]
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  assert  resp.status_code == 200, resp.content
  ops = [op['id'] for op in resp.json()['operational_intent_reference']]
  assert ids(OP_TYPE) in ops, resp.content

  # ids(OP_ID2) not expected here because its ceiling is <575m whereas query floor is
  # >591m.
  assert ids(OP_TYPE2) not in ops, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_big_operation_search_v5(scd_api, scd_session):
  """
  This test reproduces a case where a search resulted in 503 because the
  underlying gRPC backend had crashed.
  """
  with open('./scd/resources/op_big_operation.json', 'r') as f:
    req = json.load(f)
  dt = datetime.datetime.utcnow() - scd.start_of([req['area_of_interest']])
  req['area_of_interest'] = scd.offset_time([req['area_of_interest']], dt)[0]
  resp = scd_session.post('/operation_references/query', json=req)
  assert  resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_big_operation_search_v15(scd_api, scd_session):
  with open('./scd/resources/op_big_operation.json', 'r') as f:
    req = json.load(f)
  dt = datetime.datetime.utcnow() - scd.start_of([req['area_of_interest']])
  req['area_of_interest'] = scd.offset_time([req['area_of_interest']], dt)[0]
  resp = scd_session.post('/operational_intent_reference/query', json=req)
  assert  resp.status_code == 400, resp.content


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_ensure_clean_workspace_v5(ids, scd_api, scd_session):
  for op_id in (ids(OP_TYPE), ids(OP_TYPE2)):
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      # only the owner of the subscription can delete a operation reference.
      resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_ensure_clean_workspace_v15(ids, scd_api, scd_session):
  for op_id in (ids(OP_TYPE), ids(OP_TYPE2)):
    resp = scd_session.get('/operational_intent_reference/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      # only the owner of the subscription can delete a operation reference.
      resp = scd_session.delete('/operational_intent_reference/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content
