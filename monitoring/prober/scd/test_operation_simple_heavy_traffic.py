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

from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


BASE_URL = 'https://example.com/uss'
OP_TYPES = [register_resource_type(10 + i, 'Operational intent {}'.format(i)) for i in range(100)]

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


@for_api_versions(scd.API_0_3_5)
def test_ensure_clean_workspace_v5(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
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
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_get_v5(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_get_v15(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 404, resp.content



# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_query_v5(ids, scd_api, scd_session):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)

  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_ops_do_not_exist_query_v15(ids, scd_api, scd_session):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)

  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


# Preconditions: None
# Mutations: Operations with ids in OP_IDS created by scd_session user
@for_api_versions(scd.API_0_3_5)
def test_create_ops_v5(ids, scd_api, scd_session):
  assert len(ovn_map) == 0

  for idx, op_id in enumerate(map(ids, OP_TYPES)):
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

  assert len(ovn_map) == len(OP_TYPES)


@for_api_versions(scd.API_0_3_17)
def test_create_ops_v15(ids, scd_api, scd_session):
  assert len(ovn_map) == 0

  for idx, op_id in enumerate(map(ids, OP_TYPES)):
    req = _make_op_request(idx)
    req['key'] = list(ovn_map.values())

    resp = scd_session.put(
      '/operational_intent_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operational_intent_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert op['uss_availability'] == "Unknown"
    assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
    assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
    assert op['version'] == 1
    assert op['ovn']
    assert 'subscription_id' in op

    ovn_map[op_id] = op['ovn']

  assert len(ovn_map) == len(OP_TYPES)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
def test_get_ops_by_ids_v5(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operation_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert op['version'] == 1
    assert 'state' not in op


@for_api_versions(scd.API_0_3_17)
def test_get_ops_by_ids_v15(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operational_intent_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert op['version'] == 1


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_v5(ids, scd_api, scd_session):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_v15(ids, scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
  print(found_ids)
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_included_v5(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_included_v15(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_excluded_v5(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_earliest_time_excluded_v15(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_included_v5(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_included_v15(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
  assert len(_intersection(map(ids, OP_TYPES), found_ids)) == len(OP_TYPES)



# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_excluded_v5(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_ops_by_search_latest_time_excluded_v15(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 12000))
  })
  assert resp.status_code == 200, resp.content
  found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
  assert not _intersection(map(ids, OP_TYPES), found_ids)


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: Operations with ids in OP_IDS mutated to second version
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_mutate_ops_v5(ids, scd_api, scd_session):
  for idx, op_id in enumerate(map(ids, OP_TYPES)):
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


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_mutate_ops_v17(ids, scd_api, scd_session):
  for idx, op_id in enumerate(map(ids, OP_TYPES)):
    # GET current op
    resp = scd_session.get('/operational_intent_references/{}'.format(op_id))
    assert resp.status_code == 200, resp.content
    existing_op = resp.json().get('operational_intent_reference', None)
    assert existing_op is not None

    req = _make_op_request(idx)

    # QUERY ops in the area and get their ovns
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': req['extents'][0]
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_references', [])]
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

    resp = scd_session.put('/operational_intent_references/{}/{}'.format(op_id, existing_op['ovn']), json=req, scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    op = data['operational_intent_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == 'https://example.com/uss2'
    assert op['uss_availability'] == "Unknown"
    assert op['version'] != existing_op['version']
    assert op['subscription_id'] == existing_op['subscription_id']

    ovn_map[op_id] = op['ovn']


# Preconditions: Operations with ids in OP_IDS mutated to second version
# Mutations: Operations with ids in OP_IDS deleted
@for_api_versions(scd.API_0_3_5)
def test_delete_op_v5(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_delete_op_v15(ids, scd_api, scd_session):
  for op_id in map(ids, OP_TYPES):
    resp = scd_session.delete('/operational_intent_references/{}/{}'.format(op_id, ovn_map[op_id]))
    assert resp.status_code == 200, resp.content


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_ids_v5(ids, scd_api, scd_session):
    for op_id in map(ids, OP_TYPES):
      resp = scd_session.get('/operation_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_ids_v15(ids, scd_api, scd_session):
    for op_id in map(ids, OP_TYPES):
      resp = scd_session.get('/operational_intent_references/{}'.format(op_id))
      assert resp.status_code == 404, resp.content


# Preconditions: Operations with ids in OP_IDS deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_search_v5(ids, scd_api, scd_session):
    resp = scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    assert not _intersection(map(ids, OP_TYPES), found_ids)


@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_ops_by_search_v15(ids, scd_api, scd_session):
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 12000))
    })
    assert resp.status_code == 200, resp.content
    found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    assert not _intersection(map(ids, OP_TYPES), found_ids)
