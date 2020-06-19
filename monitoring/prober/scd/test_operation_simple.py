"""Basic Operation tests:

  - make sure the Operation doesn't exist with get or query
  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime

from . import common


def _make_op1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [common.make_vol4(time_start, time_end, 0, 120, common.make_circle(-56, 178, 50))],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': 'https://example.com/dss',
    'new_subscription': {
      'uss_base_url': 'https://example.com/dss',
      'notify_for_constraints': False
    }
  }


# Preconditions: None
# Mutations: None
def test_op_does_not_exist_get(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid), scope=common.SCOPE_SC)
  assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
def test_op_does_not_exist_query(scd_session, op1_uuid):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(time_now, time_now, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [op['id'] for op in resp.json().get('operation_references', [])]


# Preconditions: None
# Mutations: None
def test_create_op_single_extent(scd_session, op1_uuid):
  req = _make_op1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
def test_create_op_missing_time_start(scd_session, op1_uuid):
  req = _make_op1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
def test_create_op_missing_time_end(scd_session, op1_uuid):
  req = _make_op1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: Operation op1_uuid created by scd_session user
def test_create_op(scd_session, op1_uuid):
  req = _make_op1_request()
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == op1_uuid
  assert op['uss_base_url'] == 'https://example.com/dss'
  assert op['time_start']['value'] == req['extents'][0]['time_start']['value']
  assert op['time_end']['value'] == req['extents'][0]['time_end']['value']
  assert op['version'] == 1
  assert 'subscription_id' in op
  assert 'state' not in op


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_id(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == op1_uuid
  assert op['uss_base_url'] == 'https://example.com/dss'
  assert op['version'] == 1
  assert 'state' not in op


# Preconditions: None, though preferably Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search_missing_params(scd_session):
  resp = scd_session.post('/operation_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search(scd_session, op1_uuid):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json().get('operation_references', [])]


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search_earliest_time_included(scd_session, op1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search_earliest_time_excluded(scd_session, op1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search_latest_time_included(scd_session, op1_uuid):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: None
def test_get_op_by_search_latest_time_excluded(scd_session, op1_uuid):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]


# Preconditions: Operation op1_uuid created by scd_session user
# Mutations: Operation op1_uuid mutated to second version
def test_mutate_op(scd_session, op1_uuid):
  # GET current op
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operation_reference', None)
  assert existing_op is not None

  req = _make_op1_request()
  resp = scd_session.put(
    '/operation_references/{}'.format(op1_uuid),
    json={
      'key': [existing_op["ovn"]],
      'extents': req['extents'],
      'old_version': existing_op['version'],
      'state': 'Activated',
      'uss_base_url': 'https://example.com/dss2',
      'subscription_id': existing_op['subscription_id']
    })
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == op1_uuid
  assert op['uss_base_url'] == 'https://example.com/dss2'
  assert op['version'] == 2
  assert op['subscription_id'] == existing_op['subscription_id']
  assert 'state' not in op


# Preconditions: Operation op1_uuid mutated to second version
# Mutations: Operation op1_uuid deleted
def test_delete_op(scd_session, op1_uuid):
  resp = scd_session.delete('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content


# Preconditions: Operation op1_uuid deleted
# Mutations: None
def test_get_deleted_op_by_id(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 404, resp.content


# Preconditions: Operation op1_uuid deleted
# Mutations: None
def test_get_deleted_op_by_search(scd_session, op1_uuid):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]

