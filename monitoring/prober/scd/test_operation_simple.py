"""Basic Operation tests:

  - make sure the Operation doesn't exist with get or query
  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC, SCOPE_CI, SCOPE_CM
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


BASE_URL = 'https://example.com/uss'
OP_TYPE = register_resource_type(9, 'Operational intent')


@for_api_versions(scd.API_0_3_5)
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
def test_ensure_clean_workspace_v17(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


def _make_op1_request():
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-56, 178, 50))],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': BASE_URL,
    'new_subscription': {
      'uss_base_url': BASE_URL,
      'notify_for_constraints': False
    }
  }


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_does_not_exist_get_v5(ids, scd_api, scd_session):
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_does_not_exist_get_v17(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_does_not_exist_query_v5(ids, scd_api, scd_session):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [op['id'] for op in resp.json().get('operation_references', [])]

  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_does_not_exist_query_v17(ids, scd_api, scd_session):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [op['id'] for op in resp.json().get('operational_intent_reference', [])]

  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_create_op_single_extent_v5(ids, scd_api, scd_session):
  req = _make_op1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_create_op_single_extent_v17(ids, scd_api, scd_session):
  req = _make_op1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_create_op_missing_time_start_v5(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_create_op_missing_time_start_v17(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_create_op_missing_time_end_v5(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_create_op_missing_time_end_v17(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: Operation ids(OP_ID) created by scd_session user
@for_api_versions(scd.API_0_3_5)
def test_create_op_v5(ids, scd_api, scd_session):
  req = _make_op1_request()

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
  assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
  assert op['version'] == 1
  assert 'subscription_id' in op
  assert 'state' not in op


# Preconditions: None
# Mutations: Operation ids(OP_ID) created by scd_session user
@for_api_versions(scd.API_0_3_17)
def test_create_op_v17(ids, scd_api, scd_session):
  req = _make_op1_request()

  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == "Unknown"
  assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
  assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
  assert op['version'] == 1
  assert 'subscription_id' in op
  assert op['state'] == 'Accepted'


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
def test_get_op_by_id_v5(ids, scd_api, scd_session):
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['version'] == 1
  assert 'state' not in op


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
def test_get_op_by_id_v17(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == "Unknown"
  assert op['version'] == 1
  assert 'state' in op
  assert op['state'] == 'Accepted',\
          "The response has a state = '{}'"\
          .format(data['operational_intent_reference']['state'])


# Preconditions: None, though preferably Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_missing_params_v5(scd_api, scd_session):
  resp = scd_session.post('/operation_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: None, though preferably Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_missing_params_v17(scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_v5(ids, scd_api, scd_session):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json().get('operation_references', [])]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_v17(ids, scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json().get('operational_intent_references', [])], resp.json()


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_earliest_time_included_v5(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_earliest_time_included_v17(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operational_intent_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_earliest_time_excluded_v5(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_earliest_time_excluded_v17(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_latest_time_included_v5(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_latest_time_included_v17(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operational_intent_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_op_by_search_latest_time_excluded_v5(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_op_by_search_latest_time_excluded_v17(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: Operation ids(OP_ID) mutated to second version
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_mutate_op_v5(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  req = _make_op1_request()
  req = {
    'key': [existing_op["ovn"]],
    'extents': req['extents'],
    'old_version': existing_op['version'],
    'state': 'Activated',
    'uss_base_url': 'https://example.com/uss2',
    'subscription_id': existing_op['subscription_id']
  }

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operation_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == 'https://example.com/uss2'
  assert op['version'] == 2
  assert op['subscription_id'] == existing_op['subscription_id']
  assert 'state' not in op


# Preconditions: Operation ids(OP_ID) created by scd_session user
# Mutations: Operation ids(OP_ID) mutated to second version
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_mutate_op_v17(ids, scd_api, scd_session):
  # GET current op
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operational_intent_reference', None)
  assert existing_op is not None, resp.json()

  req = _make_op1_request()
  req = {
    'key': [existing_op["ovn"]],
    'extents': req['extents'],
    'old_version': existing_op['version'],
    'state': 'Activated',
    'uss_base_url': 'https://example.com/uss2',
    'subscription_id': existing_op['subscription_id']
  }

  resp = scd_session.put(
    '/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put(
    '/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put(
    '/operational_intent_references/{}/{}'.format(ids(OP_TYPE), existing_op["ovn"]), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == 'https://example.com/uss2'
  assert op['version'] == 2
  assert op['subscription_id'] == existing_op['subscription_id']
  # assert 'state' not in op


# Preconditions: Operation ids(OP_ID) mutated to second version
# Mutations: Operation ids(OP_ID) deleted
@for_api_versions(scd.API_0_3_5)
def test_delete_op_v5(ids, scd_api, scd_session):
  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/operation_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content


# Preconditions: Operation ids(OP_ID) mutated to second version
# Mutations: Operation ids(OP_ID) deleted
@for_api_versions(scd.API_0_3_17)
def test_delete_op_v17(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  ovn = resp.json()['operational_intent_reference']['ovn']

  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_CM)
  assert resp.status_code == 403, resp.content


  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content


# Preconditions: Operation ids(OP_ID) deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_deleted_op_by_id_v5(ids, scd_api, scd_session):
  resp = scd_session.get('/operation_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: Operation ids(OP_ID) deleted
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_op_by_id_v17(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: Operation ids(OP_ID) deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_deleted_op_by_search_v5(ids, scd_api, scd_session):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation ids(OP_ID) deleted
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_op_by_search_v17(ids, scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]
