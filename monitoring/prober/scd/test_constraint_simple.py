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
from monitoring.monitorlib.scd import SCOPE_SC, SCOPE_CI, SCOPE_CM, SCOPE_CP , SCOPE_CM_SA, SCOPE_AA
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober import utils
from monitoring.prober.infrastructure import for_api_versions


BASE_URL = 'https://example.com/uss'
CONSTRAINT_ID = ''



def test_set_test_owner_ids(test_owner):
  global CONSTRAINT_ID
  CONSTRAINT_ID = utils.encode_owner(test_owner, '0000006f-5a03-482a-a7e1-23c29c000000')

get_test_owner = lambda: test_set_test_owner_ids(test_owner)

def _make_c1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-56, 178, 50))],
    'old_version': 0,
    'uss_base_url': BASE_URL,
  }


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_ensure_clean_workspace(scd_api, scd_session):
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
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_constraint_does_not_exist_get(scd_api, scd_session):
  auths = None
  if scd_api == scd.API_0_3_5:
      auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
      auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=scope)
    assert resp.status_code == 404, "ConstraintID: {}\n{}".format(CONSTRAINT_ID, resp.content)


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_constraint_does_not_exist_query(scd_api, scd_session):
  if scd_session is None:
    return

  time_now = datetime.datetime.utcnow()
  auths = None
  if scd_api == scd.API_0_3_5:
      auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
      auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, time_now, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert CONSTRAINT_ID not in [constraint['id'] for constraint in resp.json().get('constraint_references', [])]


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_single_extent(scd_api, scd_session):
  req = _make_c1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_start(scd_api, scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint_missing_time_end(scd_api, scd_session):
  req = _make_c1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: Constraint CONSTRAINT_ID created by scd_session user
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_create_constraint(scd_api, scd_session):
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
  assert_datetimes_are_equal(constraint['time_start']['value'], req['extents'][0]['time_start']['value'])
  assert_datetimes_are_equal(constraint['time_end']['value'], req['extents'][0]['time_end']['value'])
  assert constraint['version'] == 1


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_get_constraint_by_id(scd_api, scd_session):
  auths = None
  if scd_api == scd.API_0_3_5:
      auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
      auths = (SCOPE_CM, SCOPE_CP)

  for scope in auths:
    resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=scope)
    assert resp.status_code == 200, resp.content

    data = resp.json()
    constraint = data['constraint_reference']
    assert constraint['id'] == CONSTRAINT_ID
    assert constraint['uss_base_url'] == BASE_URL
    assert constraint['version'] == 1


# Preconditions: None, though preferably Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_missing_params(scd_api, scd_session):
  resp = scd_session.post('/constraint_references/query')
  assert resp.status_code == 400, resp.content


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_get_constraint_by_search(scd_api, scd_session):
  auths = None
  if scd_api == scd.API_0_3_5:
      auths = (SCOPE_SC, SCOPE_CI, SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
      auths = (SCOPE_CM, SCOPE_CP)
  for scope in auths:
    resp = scd_session.post('/constraint_references/query', json={
      'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=scope)
    assert resp.status_code == 200, resp.content
    assert CONSTRAINT_ID in [x['id'] for x in resp.json().get('constraint_references', [])]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_earliest_time_included(scd_api, scd_session):

  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_earliest_time_excluded(scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_latest_time_included(scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_constraint_by_search_latest_time_excluded(scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]


# Preconditions: Constraint CONSTRAINT_ID created by scd_session user
# Mutations: Constraint CONSTRAINT_ID mutated to second version
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_mutate_constraint(scd_api, scd_session):
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
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_delete_constraint(scd_api, scd_session):
  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_SC)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CI)
  assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content


# Preconditions: Constraint CONSTRAINT_ID deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_deleted_constraint_by_id(scd_api, scd_session):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: Constraint CONSTRAINT_ID deleted
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_get_deleted_constraint_by_search(scd_api, scd_session):
  resp = scd_session.post('/constraint_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert CONSTRAINT_ID not in [x['id'] for x in resp.json()['constraint_references']]
