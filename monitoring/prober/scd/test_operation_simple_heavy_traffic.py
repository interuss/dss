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

from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions


def _load_op_ids():
  with open('./scd/resources/op_ids_heavy_traffic_sequential.json', 'r') as f:
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
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(lat, 178, 50))],
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


@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
def test_ensure_clean_workspace(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
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
  elif scd_api == scd.API_0_3_15:
    for op_id in OP_IDS:
      resp = scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      if resp.status_code == 200:
        resp = scd_session.delete('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
        assert resp.status_code == 200, resp.content
      elif resp.status_code == 404:
        # As expected.
        pass
      else:
        assert False, resp.content
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_get(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    for op_id in OP_IDS:
      resp = scd_session.get('/operation_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content
  elif scd_api == scd.API_0_3_15:
    for op_id in OP_IDS:
      resp = scd_session.get('/operation_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_query(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    time_now = datetime.datetime.utcnow()
    end_time = time_now + datetime.timedelta(hours=1)

    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    }, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert not _intersection(OP_IDS, found_ids)
  elif scd_api == scd.API_0_3_15:
    time_now = datetime.datetime.utcnow()
    end_time = time_now + datetime.timedelta(hours=1)

    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    }, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert not _intersection(OP_IDS, found_ids)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: None
# Mutations: Operations with ids in OP_IDS created by scd_session user
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
def test_create_ops(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
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
      assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
      assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
      assert op['version'] == 1
      assert op['ovn']
      assert 'subscription_id' in op
      assert 'state' not in op

      ovn_map[op_id] = op['ovn']

    assert len(ovn_map) == len(OP_IDS)
  elif scd_api == scd.API_0_3_15:
    assert len(ovn_map) == 0

    for idx, op_id in enumerate(OP_IDS):
      req = _make_op_request(idx)
      req['key'] = list(ovn_map.values())

      resp = scd_session.put(
        '/operational_intent_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content

      data = resp.json()
      op = data['operational_intent_reference']
      assert op['id'] == op_id
      assert op['uss_base_url'] == BASE_URL
      assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
      assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
      assert op['version'] == 1
      assert op['ovn']
      assert 'subscription_id' in op
      assert 'state' not in op

      ovn_map[op_id] = op['ovn']

    assert len(ovn_map) == len(OP_IDS)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
def test_get_ops_by_ids(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    for op_id in OP_IDS:
      resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content

      data = resp.json()
      op = data['operation_reference']
      assert op['id'] == op_id
      assert op['uss_base_url'] == BASE_URL
      assert op['version'] == 1
      assert 'state' not in op
  elif scd_api == scd.API_0_3_15:
    for op_id in OP_IDS:
      resp = scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content

      data = resp.json()
      op = data['operation_reference']
      assert op['id'] == op_id
      assert op['uss_base_url'] == BASE_URL
      assert op['version'] == 1
      assert 'state' not in op
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_ops_by_search(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  elif scd_api == scd.API_0_3_15:
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_included(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  elif scd_api == scd.API_0_3_15:
    earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_excluded(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert not _intersection(OP_IDS, found_ids)
  elif scd_api == scd.API_0_3_15:
    earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert not _intersection(OP_IDS, found_ids)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_included(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  elif scd_api == scd.API_0_3_15:
    latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
    assert len(_intersection(OP_IDS, found_ids)) == len(OP_IDS)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_excluded(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert not _intersection(OP_IDS, found_ids)
  elif scd_api == scd.API_0_3_15:
    latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
    assert not _intersection(OP_IDS, found_ids)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: Operations with ids in OP_IDS mutated to second version
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_mutate_ops(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
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
  elif scd_api == scd.API_0_3_15:
    for idx, op_id in enumerate(OP_IDS):
      # GET current op
      resp = scd_session.get('/operational_intent_references/{}'.format(op_id))
      assert resp.status_code == 200, resp.content
      existing_op = resp.json().get('operation_reference', None)
      assert existing_op is not None

      req = _make_op_request(idx)

      # QUERY ops in the area and get their ovns
      resp = scd_session.post('/operational_intent_references/query', json={
        'area_of_interest': req['extents'][0]
      })
      assert resp.status_code == 200, resp.content
      found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
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

      resp = scd_session.put('/operational_intent_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content

      data = resp.json()
      op = data['operation_reference']
      assert op['id'] == op_id
      assert op['uss_base_url'] == 'https://example.com/uss2'
      assert op['version'] == 2
      assert op['subscription_id'] == existing_op['subscription_id']
      assert 'state' not in op

      ovn_map[op_id] = op['ovn']
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS mutated to second version
# Mutations: Operations with ids in OP_IDS deleted
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
def test_delete_op(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    for op_id in OP_IDS:
      resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
  elif scd_api == scd.API_0_3_15:
    for op_id in OP_IDS:
      resp = scd_session.delete('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_ids(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    for op_id in OP_IDS:
      resp = scd_session.get('/operation_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content
  elif scd_api == scd.API_0_3_15:
    for op_id in OP_IDS:
      resp = scd_session.get('/operational_intent_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_search(scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert not _intersection(OP_IDS, found_ids)
  elif scd_api == scd.API_0_3_15:
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert not _intersection(OP_IDS, found_ids)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))
