"""Basic Constraint tests:

  - make sure the Constraint doesn't exist with get or query
  - create the Constraint with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from . import common
from .common import SCOPE_SC, SCOPE_CI, SCOPE_CM


BASE_URL = 'https://example.com/uss'
CONSTRAINT_ID = '0000006f-5a03-482a-a7e1-23c29c000000'


def _make_c1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [common.make_vol4(time_start, time_end, 0, 120, common.make_circle(-56, 178, 50))],
    'old_version': 0,
    'uss_base_url': BASE_URL,
  }


def test_ensure_clean_workspace(scd_session):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
  if resp.status_code == 200:
    resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


# Preconditions: None
# Mutations: None
def test_constraint_does_not_exist_get(scd_session):
  for scope in (SCOPE_SC, SCOPE_CI, SCOPE_CM):
    resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=scope)
    assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
def test_constraint_does_not_exist_query(scd_session):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()
  for scope in (SCOPE_SC, SCOPE_CI, SCOPE_CM):
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': common.make_vol4(time_now, time_now, 0, 5000, common.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert CONSTRAINT_ID not in [constraint['id'] for constraint in resp.json().get('constraint_references', [])]


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_CM)
def test_create_constraint_single_extent(scd_session):
  req = _make_c1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_start(scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_end(scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: Constraint CONSTRAINT_ID created by scd_session user
def test_create_constraint(scd_session):
  req = _make_c1_request()

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_SC)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  constraint = data['constraint_reference']
  assert constraint['id'] == CONSTRAINT_ID
  assert constraint['uss_base_url'] == BASE_URL
  assert constraint['time_start']['value'] == req['extents'][0]['time_start']['value']
  assert constraint['time_end']['value'] == req['extents'][0]['time_end']['value']
  assert constraint['version'] == 1


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
def test_get_constraint_by_id(scd_session):
  for scope in (SCOPE_SC, SCOPE_CI, SCOPE_CM):
    resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=scope)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    constraint = data['constraint_reference']
    assert constraint['id'] == CONSTRAINT_ID
    assert constraint['uss_base_url'] == BASE_URL
    assert constraint['version'] == 1


# Preconditions: None, though preferably Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_constraint_by_search_missing_params(scd_session):
  resp = scd_session.post('/constraint_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
def test_get_constraint_by_search(scd_session):
  for scope in (SCOPE_SC, SCOPE_CI, SCOPE_CM):
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert CONSTRAINT_ID in [x['id'] for x in resp.json().get('constraint_references', [])]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_constraint_by_search_earliest_time_included(scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_constraint_by_search_earliest_time_excluded(scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_constraint_by_search_latest_time_included(scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_constraint_by_search_latest_time_excluded(scd_session):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: Constraint CONSTRAINT_ID mutated to second version
def test_mutate_constraint(scd_session):
  # GET current constraint
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CI)
  assert resp.status_code == 200, resp.content
  existing_constraint = resp.json().get('constraint_reference', None)
  assert existing_constraint is not None

  req = _make_c1_request()
  req = {
    'key': [existing_constraint["ovn"]],
    'extents': req['extents'],
    'old_version': existing_constraint['version'],
    'uss_base_url': 'https://example.com/uss2'
  }

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_SC)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  constraint = data['constraint_reference']
  assert constraint['id'] == CONSTRAINT_ID
  assert constraint['uss_base_url'] == 'https://example.com/uss2'
  assert constraint['version'] == 2


# Preconditions: Constraint CONSTRAINT_ID mutated to second version
# Mutations: Constraint CONSTRAINT_ID deleted
def test_delete_constraint(scd_session):
  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_SC)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content


# Preconditions: Constraint CONSTRAINT_ID deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_constraint_by_id(scd_session):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: Constraint CONSTRAINT_ID deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_constraint_by_search(scd_session):
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]

