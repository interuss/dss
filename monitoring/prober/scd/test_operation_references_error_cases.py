"""Operation References corner cases error tests:
"""

import json
import uuid


def test_op_ref_area_too_large(scd_session):
  with open('./scd/resources/op_ref_area_too_large.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


def test_op_ref_start_end_times_past(scd_session):
  with open('./scd/resources/op_ref_start_end_times_past.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


def test_op_ref_incorrect_units(scd_session):
  with open('./scd/resources/op_ref_incorrect_units.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 500, resp.content


def test_op_ref_incorrect_altitude_ref(scd_session):
  with open('./scd/resources/op_ref_incorrect_altitude_ref.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 500, resp.content


def test_op_uss_base_url_non_tls(scd_session, op1_uuid):
  with open('./scd/resources/op_uss_base_url_non_tls.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_bad_keys(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_keys.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_bad_subscription_id(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_subscription.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_bad_subscription_id_random(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_subscription.json', 'r') as f:
    req = json.load(f)
    req['subscription_id'] = uuid.uuid4().hex
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 500, resp.content


def test_op_new_and_existing_subscription(scd_session, op1_uuid):
  with open('./scd/resources/op_new_and_existing_subscription.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 500, resp.content


def test_op_end_time_past(scd_session, op1_uuid):
  with open('./scd/resources/op_end_time_past.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_already_exists(scd_session, op1_uuid):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content


def test_op_404_version1(scd_session, op1_uuid):
  with open('./scd/resources/op_404_version1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 404, resp.content


def test_op_bad_state_version0(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_state_version0.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_bad_lat_lon_range(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_lat_lon_range.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_area_too_large_put(scd_session, op1_uuid):
  with open('./scd/resources/op_area_too_large_put.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_bad_time_format(scd_session, op1_uuid):
  with open('./scd/resources/op_bad_time_format.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


def test_op_repeated_requests(scd_session, op1_uuid):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 200, resp.content

  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 409, resp.content

  # Delete operation
  resp = scd_session.delete('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content
