"""Basic Operation tests with hundreds of NON-OVERLAPPING operations created CONCURRENTLY.
   The core actions are performed in parallel while others like cleanup, assert response, etc are intended to remain
   sequential.

  - make sure operations do not exist with get or query
  - create 100 operations concurrently, with has non-overlapping volume4d in 2ds, altitude ranges and time windows.
  - get by IDs concurrently
  - search by areas concurrently
  - mutate operations concurrently
  - delete operations concurrently
  - confirm deletion by get and query
"""


import asyncio
import datetime
import pytest
from concurrent.futures.thread import ThreadPoolExecutor

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


# To avoid `loop is already running` error.
# https://stackoverflow.com/questions/46827007/runtimeerror-this-event-loop-is-already-running-in-python
import nest_asyncio
nest_asyncio.apply()

# This test is implemented to fire requests concurrently, given there are several concurrent related issues:
# - https://github.com/interuss/dss/issues/417
# - https://github.com/interuss/dss/issues/418
# - https://github.com/interuss/dss/issues/419
# - https://github.com/interuss/dss/issues/420
# - https://github.com/interuss/dss/issues/421
# We intended to keep the thread count to be 1 to enforce sequential execution till the above issues are resolved.
# By then, just update the 'THREAD_COUNT' to a reasonable thread pool size and everything should still work without
# need to touch anything else.
THREAD_COUNT = 1
BASE_URL = 'https://example.com/uss'
OP_TYPES = [register_resource_type(110 + i, 'Operational intent {}'.format(i)) for i in range(100)]
GROUP_SIZE = len(OP_TYPES) // 3

ovn_map = {}


def _calculate_lat(idx):
  return -56 - 0.1 * idx


def _make_op_request_with_extents(extents):
  return {
    'extents': [extents],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': BASE_URL,
    'new_subscription': {
      'uss_base_url': BASE_URL,
      'notify_for_constraints': False
    }
  }


# Generate request with volumes that cover a circle area that initially centered at (-56, 178)
# The circle's center lat shifts 0.1 degree (11.1 km) per sequential idx change
# The altitude and time window won't change with idx
def _make_op_request_differ_in_2d(idx):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = _calculate_lat(idx)

  vol4 = scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(lat, 178, 50))
  return _make_op_request_with_extents(vol4)


# Generate request with volumes that cover the circle area that centered at (-56, 178)
# The altitude starts with [0, 19] and increases 20 per sequential idx change
# The 2D area and time window won't change with idx
def _make_op_request_differ_in_altitude(idx):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  delta = 20
  alt0 = delta * idx
  alt1 = alt0 + delta - 1

  vol4 = scd.make_vol4(time_start, time_end, alt0, alt1, scd.make_circle(-56, 178, 50))
  return _make_op_request_with_extents(vol4)


# Generate request with volumes that cover the circle area that centered at (-56, 178), with altitude 0 to 120
# The operation lasts 9 mins and the time window is one after one per sequential idx change
# The 2D area and altitude won't change with idx
def _make_op_request_differ_in_time(idx):
  delta = 10
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20) + datetime.timedelta(minutes=delta * idx)
  time_end = time_start + datetime.timedelta(minutes=delta - 1)

  vol4 = scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-56, 178, 50))
  return _make_op_request_with_extents(vol4)


# Generate request with non-overlapping operations in volume4d.
# 1/3 operations will be generated with different 2d areas, altitude ranges and time windows respectively
def _make_op_request(idx):
  if idx < GROUP_SIZE:
    return _make_op_request_differ_in_2d(idx)
  elif idx < GROUP_SIZE * 2:
    return _make_op_request_differ_in_altitude(idx)
  else:
    return _make_op_request_differ_in_time(idx)


def _intersection(list1, list2):
  return list(set(list1) & set(list2))


def _put_operation(req, op_id, scd_session, scd_api):
  if scd_api == scd.API_0_3_5:
    return scd_session.put('/operation_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
  elif scd_api == scd.API_0_3_15:
    return scd_session.put('/operational_intent_references/{}'.format(op_id), json=req, scope=SCOPE_SC)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


def _get_operation(op_id, scd_session, scd_api):
  if scd_api == scd.API_0_3_5:
    return scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
  elif scd_api == scd.API_0_3_15:
    return scd_session.get('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


def _query_operation(idx, scd_session, scd_api):
  lat = _calculate_lat(idx)
  if scd_api == scd.API_0_3_5:
    return scd_session.post('/operation_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(lat, 178, 12000))
    }, scope=SCOPE_SC)
  elif scd_api == scd.API_0_3_15:
    return scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(lat, 178, 12000))
    }, scope=SCOPE_SC)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))


def _build_mutate_request(idx, op_id, op_map, scd_session, scd_api):
  # GET current op
  if scd_api == scd.API_0_3_5:
    resp = scd_session.get('/operation_references/{}'.format(op_id))
    assert resp.status_code == 200, resp.content
    existing_op = resp.json().get('operation_reference', None)
    assert existing_op is not None
    op_map[op_id] = existing_op
  elif scd_api == scd.API_0_3_15:
    resp = scd_session.get('/operational_intent_references/{}'.format(op_id))
    assert resp.status_code == 200, resp.content
    existing_op = resp.json().get('operational_intent_reference', None)
    assert existing_op is not None
    op_map[op_id] = existing_op
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))

  req = _make_op_request(idx)
  req = {
    'key': [existing_op["ovn"]],
    'extents': req['extents'],
    'old_version': existing_op['version'],
    'state': 'Activated',
    'uss_base_url': 'https://example.com/uss2',
    'subscription_id': existing_op['subscription_id']
  }
  return req


def _delete_operation(op_id, scd_session, scd_api):
  if scd_api == scd.API_0_3_5:
    return scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
  elif scd_api == scd.API_0_3_15:
    return scd_session.delete('/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
  else:
    raise ValueError('Unsupported SCD API version: {}'.format(scd_api))

def _collect_resp_callback(key, op_resp_map, future):
  op_resp_map[key] = future.result()


@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
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


@for_api_versions(scd.API_0_3_15)
@default_scope(SCOPE_SC)
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


async def _create_ops_concurrent(executor, loop, ids, scd_api, scd_session, op_req_map):
  blocking_tasks = {}
  for idx, op_id in enumerate(map(ids, OP_TYPES)):
      req = _make_op_request(idx)
      op_req_map[op_id] = req
      blocking_tasks.update(
        {loop.run_in_executor(executor, _put_operation, req, op_id, scd_session, scd_api): op_id})
  responses = {}
  for task, op_id in blocking_tasks.items():
    responses[op_id] = await task
  return responses

# Preconditions: None
# Mutations: Operations with ids in OP_IDS created by scd_session user
@pytest.mark.asyncio
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_create_ops_concurrent(ids, scd_api, scd_session):
  assert len(ovn_map) == 0
  op_req_map = {}
  op_resp_map = {}
  # Create operations concurrently
  event_loop = asyncio.get_event_loop()
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    results = event_loop.run_until_complete(
        _create_ops_concurrent(executor, event_loop, ids, scd_api, scd_session, op_req_map)
    )
    for op_id, result in results.items():
      op_resp_map[op_id] = result
  for op_id, resp in op_resp_map.items():
    assert resp.status_code == 200, resp.content
    req = op_req_map[op_id]
    data = resp.json()
    if scd_api == scd.API_0_3_5:
      op = data['operation_reference']
    else:
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
  assert len(ovn_map) == len(OP_TYPES)


async def _get_ops_by_ids_concurrent(executor, loop, ids, scd_api, scd_session):
  blocking_tasks = {}
  for op_id in map(ids, OP_TYPES):
      blocking_tasks.update(
        {loop.run_in_executor(executor, _get_operation, op_id, scd_session, scd_api): op_id})
  responses = {}
  for task, op_id in blocking_tasks.items():
    responses[op_id] = await task
  return responses


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@pytest.mark.asyncio
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
async def test_get_ops_by_ids_concurrent(ids, scd_api, scd_session, event_loop):
  op_resp_map = {}
  # Get opetions concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    results = event_loop.run_until_complete(
        _get_ops_by_ids_concurrent(executor, event_loop, ids, scd_api, scd_session)
    )
    for op_id, result in results.items():
      op_resp_map[op_id] = result

  for op_id, resp in op_resp_map.items():
    assert resp.status_code == 200, resp.content

    data = resp.json()
    if scd_api == scd.API_0_3_5:
      op = data['operation_reference']
    else:
      op = data['operational_intent_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == BASE_URL
    assert op['version'] == 1
    assert 'state' not in op


async def _get_ops_by_search_concurrent(executor, loop, scd_api, scd_session):
  blocking_tasks = {}
  for idx in range(len(OP_TYPES)):
      blocking_tasks.update(
        {loop.run_in_executor(executor, _query_operation, idx, scd_session, scd_api): idx})
  responses = {}
  for task, op_id in blocking_tasks.items():
    responses[op_id] = await task
  return responses


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: None
@pytest.mark.asyncio
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
async def test_get_ops_by_search_concurrent(ids, scd_api, scd_session, event_loop):
  op_resp_map = {}
  total_found_ids = set()
  # Query opetions concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    results = event_loop.run_until_complete(
        _get_ops_by_search_concurrent(executor, event_loop, scd_api, scd_session)
    )
    for op_id, result in results.items():
      op_resp_map[op_id] = result

  for idx, resp in op_resp_map.items():
    assert resp.status_code == 200, resp.content
    if scd_api == scd.API_0_3_5:
      found_ids = [op['id'] for op in resp.json().get('operation_references', [])]
    else:
      found_ids = [op['id'] for op in resp.json().get('operational_intent_reference', [])]
    total_found_ids.update(found_ids)

  assert len(_intersection(map(ids, OP_TYPES), total_found_ids)) == len(OP_TYPES)


async def _mutate_ops_concurrent(executor, loop, ids, op_req_map, scd_api, scd_session):
  blocking_tasks = {}
  for op_id in map(ids, OP_TYPES):
      req = op_req_map[op_id]
      blocking_tasks.update(
        {loop.run_in_executor(executor, _put_operation, req, op_id, scd_session, scd_api): op_id})
  responses = {}
  for task, op_id in blocking_tasks.items():
    responses[op_id] = await task
  return responses


# Preconditions: Operations with ids in OP_IDS created by scd_session user
# Mutations: Operations with ids in OP_IDS mutated to second version
@pytest.mark.asyncio
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_mutate_ops_concurrent(ids, scd_api, scd_session, event_loop):
  op_req_map = {}
  op_resp_map = {}
  op_map = {}

  # Build mutate requests
  for idx, op_id in enumerate(map(ids, OP_TYPES)):
    op_req_map[op_id] = _build_mutate_request(idx, op_id, op_map, scd_session, scd_api)
  assert len(op_req_map) == len(OP_TYPES)

  # Mutate operations in parallel
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    results = event_loop.run_until_complete(
        _mutate_ops_concurrent(executor, event_loop, ids, op_req_map, scd_api, scd_session)
    )
    for op_id, result in results.items():
      op_resp_map[op_id] = result
  ovn_map.clear()

  for op_id, resp in op_resp_map.items():
    existing_op = op_map[op_id]
    assert existing_op

    assert resp.status_code == 200, resp.content
    data = resp.json()
    if scd_api == scd.API_0_3_5:
      op = data['operation_reference']
    else:
      op = data['operational_intent_reference']
    assert op['id'] == op_id
    assert op['uss_base_url'] == 'https://example.com/uss2'
    assert op['version'] == 2
    assert op['subscription_id'] == existing_op['subscription_id']
    assert 'state' not in op

    ovn_map[op_id] = op['ovn']

  assert len(ovn_map) == len(OP_TYPES)


async def _delete_op_concurrent(executor, loop, ids, scd_api, scd_session):
  blocking_tasks = {}
  for op_id in map(ids, OP_TYPES):
      blocking_tasks.update(
        {loop.run_in_executor(executor, _delete_operation, op_id, scd_session, scd_api): op_id})
  responses = {}
  for task, op_id in blocking_tasks.items():
    responses[op_id] = await task
  return responses


# Preconditions: Operations with ids in OP_IDS mutated to second version
# Mutations: Operations with ids in OP_IDS deleted
@pytest.mark.asyncio
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
async def test_delete_op_concurrent(ids, scd_api, scd_session, event_loop):
  op_resp_map = {}
  # Delete operations concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    results = event_loop.run_until_complete(
        _delete_op_concurrent(executor, event_loop, ids, scd_api, scd_session)
    )
    for op_id, result in results.items():
      op_resp_map[op_id] = result

  assert len(op_resp_map) == len(OP_TYPES)

  for resp in op_resp_map.values():
    assert resp.status_code == 200, resp.content
