"""Basic Operation tests with hundreds of operations created SEQUENTIALLY in the SAME area:

  - make sure operations do not exist with get or query
  - create 100 operations sequentially, with each covers non-overlapping area that are close to others
  - get by IDs
  - search with earliest_time and latest_time
  - mutate
  - delete
  - confirm deletion by get and query
"""

import datetime
import json

from . import common
from .common import SCOPE_SC
from monitoring.monitorlib.infrastructure import default_scope


def _load_op_ids():
  with open('./scd/resources/op_ids_100_1.json', 'r') as f:
    return json.load(f)


BASE_URL = 'https://example.com/uss'
OP_IDS = _load_op_ids()

ovn_map = {}


# Generate request with volumes that cover a circle area that initially centered at (-56, 178)
# The circle's center lat shifts 0.001 degree (111 meters) per sequential idx change
def _make_op_request(idx):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = -56 - 0.001 * idx
  return {
    'extents': [common.make_vol4(time_start, time_end, 0, 120, common.make_circle(lat, 178, 50))],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': BASE_URL,
    'new_subscription': {
      'uss_base_url': BASE_URL,
      'notify_for_constraints': False
    }
  }


def _intersection(list1, list2):
  return list(set(list1) & set(list2))


def test_ensure_clean_workspace(scd_session):
  for op_id in OP_IDS:
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_get(scd_session):
  for op_id in OP_IDS:
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_query(scd_session):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)

  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(time_now, end_time, 0, 5000, common.make_circle(-56, 178, 12000))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(OP_IDS, found_ids)


# Preconditions: None
# Mutations: Operations with ids in OP_IDS created by scd_session user
def test_create_ops(scd_session):
  assert len(ovn_map) == 0

  for idx, op_id in enumerate(OP_IDS):
    req = _make_op_request(idx)
    req['key'] = list(ovn_map.values())

    resp = scd_session.put('/operation_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operation_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert common.iso8601_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
    assert common.iso8601_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
    assert op['version'] == 1
    assert op['ovn']
    assert 'subscription_id' in op
    assert 'state' not in op

    ovn_map[op_id] = op['ovn']

  assert len(ovn_map) == len(OP_IDS)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
def test_get_ops_by_ids(scd_session):
  for op_id in OP_IDS:
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operation_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert op['version'] == 1
    assert 'state' not in op


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_ops_by_search(scd_session):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_included(scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_excluded(scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(OP_IDS, found_ids)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_included(scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_excluded(scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(OP_IDS, found_ids)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: Operations with ids in OP_IDS mutated to second version
@default_scope(SCOPE_SC)
def test_mutate_ops(scd_session):
  for idx, op_id in enumerate(OP_IDS):
    # GET current op
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 200, resp.content
    existing_op = resp.json().get('operation_reference', None)
    assert existing_op is not None

    req = _make_op_request(idx)

    # QUERY ops in the area and get their ovns
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': req['extents'][0]
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    ovns = [ovn_map[id] for id in found_ids]

    # UPDATE operation
    req = {
      'key': ovns,
      'extents': req['extents'],
      'old_version': existing_op['version'],
      'state': 'Activated',
      'uss_base_url': 'https://example.com/uss2',
      'subscription_id': existing_op['subscription_id']
    }

    resp = scd_session.put('/operation_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operation_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == 'https://example.com/uss2'
    assert op['version'] == 2
    assert op['subscription_id'] == existing_op['subscription_id']
    assert 'state' not in op

    ovn_map[op_id] = op['ovn']


# Preconditions: Operations with ids in OP_IDS mutated to second version
# Mutations: Operations with ids in OP_IDS deleted
def test_delete_op(scd_session):
  for op_id in OP_IDS:
    resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_ids(scd_session):
  for op_id in OP_IDS:
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 404, resp.content


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_search(scd_session):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(OP_IDS, found_ids)
