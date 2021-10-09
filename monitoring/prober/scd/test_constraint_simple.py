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
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC, SCOPE_CI, SCOPE_CM, SCOPE_CP, SCOPE_CM_SA, SCOPE_AA
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions, register_resource_type

import pytest


BASE_URL = 'https://example.com/uss'
CONSTRAINT_TYPE = register_resource_type(1, 'Single constraint')


def _make_c1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-56, 178, 50))],
    'old_version': 0,
    'uss_base_url': BASE_URL,
  }


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session, scd_session_cm):
  if not scd_session_cm:
    pytest.skip('SCD auth1 not enabled for constraint management')
  resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), scope=SCOPE_CM)
  if resp.status_code == 200:
    resp = scd_session.delete('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), scope=SCOPE_CM)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_constraint_does_not_exist_get(ids, scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), scope=scope)
    assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_constraint_does_not_exist_query(ids, scd_api, scd_session):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()

  if scd_api == scd.API_0_3_5:
    auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, time_now, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert ids(CONSTRAINT_TYPE) not in [constraint['id'] for constraint in resp.json().get('constraint_references', [])]


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_single_extent(ids, scd_api, scd_session):
  req = _make_c1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_start(ids, scd_api, scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_end(ids, scd_api, scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: Constraint SIMPLE_CONSTRAINT created by scd_session user
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_create_constraint(ids, scd_api, scd_session):
  id = ids(CONSTRAINT_TYPE)
  req = _make_c1_request()

  resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_SC)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  constraint = data['constraint_reference']
  assert constraint['id'] == id
  assert constraint['uss_base_url'] == BASE_URL
  assert_datetimes_are_equal(constraint['time_start']['value'], req['extents'][0]['time_start']['value'])
  assert_datetimes_are_equal(constraint['time_end']['value'], req['extents'][0]['time_end']['value'])
  assert constraint['version'] == 1


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_get_constraint_by_id(ids, scd_api, scd_session):
  id = ids(CONSTRAINT_TYPE)

  if scd_api == scd.API_0_3_5:
    auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=scope)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    constraint = data['constraint_reference']
    assert constraint['id'] == id
    assert constraint['uss_base_url'] == BASE_URL
    assert constraint['version'] == 1


# Preconditions: None, though preferably Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_missing_params(scd_api, scd_session):
  resp = scd_session.post('/constraint_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_get_constraint_by_search(ids, scd_api, scd_session):
  if scd_api == scd.API_0_3_5:
    auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert ids(CONSTRAINT_TYPE) in [x['id'] for x in resp.json().get('constraint_references', [])]


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_earliest_time_included(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(CONSTRAINT_TYPE) in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_earliest_time_excluded(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(CONSTRAINT_TYPE) not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_latest_time_included(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(CONSTRAINT_TYPE) in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_latest_time_excluded(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(CONSTRAINT_TYPE) not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint SIMPLE_CONSTRAINT created by scd_session user
# Mutations: Constraint SIMPLE_CONSTRAINT mutated to second version
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_mutate_constraint(ids, scd_api, scd_session):
  id = ids(CONSTRAINT_TYPE)
  # GET current constraint
  if scd_api == scd.API_0_3_5:
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=SCOPE_CI)
  if scd_api == scd.API_0_3_17:
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=SCOPE_CP)
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

  if scd_api == scd.API_0_3_5:
    resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_SC)
    assert resp.status_code == 403, resp.content

    resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_CI)
    assert resp.status_code == 403, resp.content

    resp = scd_session.put('/constraint_references/{}'.format(id), json=req, scope=SCOPE_CM)
    assert resp.status_code == 200, resp.content

  if scd_api == scd.API_0_3_17:
    ovn = existing_constraint["ovn"]

    # FIXME: OVNs are not URL safe
    resp = scd_session.put('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_SC)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.put('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CP)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.put('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CM_SA)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.put('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_AA)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.put('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CM)
    assert resp.status_code == 200, "ovn:{}\nresponse: {}".format(ovn, resp.content)

  data = resp.json()
  constraint = data['constraint_reference']
  assert constraint['id'] == id
  assert constraint['uss_base_url'] == 'https://example.com/uss2'
  assert constraint['version'] == 2


# Preconditions: Constraint SIMPLE_CONSTRAINT mutated to second version
# Mutations: Constraint SIMPLE_CONSTRAINT deleted
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_delete_constraint(ids, scd_api, scd_session):
  id = ids(CONSTRAINT_TYPE)

  if scd_api == scd.API_0_3_5:
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=SCOPE_CI)
  if scd_api == scd.API_0_3_17:
    resp = scd_session.get('/constraint_references/{}'.format(id), scope=SCOPE_CP)
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

  if scd_api == scd.API_0_3_5:
    resp = scd_session.delete('/constraint_references/{}'.format(id), scope=SCOPE_SC)
    assert resp.status_code == 403, resp.content

    resp = scd_session.delete('/constraint_references/{}'.format(id), scope=SCOPE_CI)
    assert resp.status_code == 403, resp.content

    resp = scd_session.delete('/constraint_references/{}'.format(id), scope=SCOPE_CM)
    assert resp.status_code == 200, resp.content
  elif scd_api == scd.API_0_3_17:
    ovn = existing_constraint["ovn"]

    resp = scd_session.delete('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_SC)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.delete('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CP)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.delete('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CM_SA)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.delete('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_AA)
    assert resp.status_code == 403, "ovn:{}\nresponse: {}".format(ovn, resp.content)

    resp = scd_session.delete('/constraint_references/{}/{}'.format(id, ovn), json=req, scope=SCOPE_CM)
    assert resp.status_code == 200, "ovn:{}\nresponse: {}".format(ovn, resp.content)


# Preconditions: Constraint SIMPLE_CONSTRAINT deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_deleted_constraint_by_id(ids, scd_api, scd_session):
  resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: Constraint SIMPLE_CONSTRAINT deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_deleted_constraint_by_search(ids, scd_api, scd_session):
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(CONSTRAINT_TYPE) not in [x['id'] for x in resp.json()['constraint_references']]
