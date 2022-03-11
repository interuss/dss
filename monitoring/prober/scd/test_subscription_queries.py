"""Strategic conflict detection Subscription query tests:

  - add a few Subscriptions spaced in time and footprints
  - query with various combinations of arguments
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type
from monitoring.prober.scd import actions


SUB1_TYPE = register_resource_type(216, 'Subscription 1')
SUB2_TYPE = register_resource_type(217, 'Subscription 2')
SUB3_TYPE = register_resource_type(218, 'Subscription 3')


LAT0 = 23
LNG0 = 56

# This value should be large enough to ensure areas separated by this distance
# will lie in separate grid cells.
FOOTPRINT_SPACING_M = 10000


def _make_sub1_req(scd_api):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = LAT0 - scd.latitude_degrees(FOOTPRINT_SPACING_M)
  req = {
    "extents": scd.make_vol4(None, time_end, 0, 300, scd.make_circle(lat, LNG0, 100)),
    "uss_base_url": "https://example.com/foo",
    "notify_for_constraints": False
  }
  if scd_api == scd.API_0_3_5:
    req.update({"old_version": 0, "notify_for_operations": True})
  elif scd_api == scd.API_0_3_17:
    req.update({"notify_for_operational_intents": True})
  return req


def _make_sub2_req(scd_api):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=2)
  time_end = time_start + datetime.timedelta(minutes=60)
  req = {
    "extents": scd.make_vol4(time_start, time_end, 350, 650, scd.make_circle(LAT0, LNG0, 100)),
    "old_version": 0,
    "uss_base_url": "https://example.com/foo",
    "notify_for_operations": True,
    "notify_for_constraints": False
  }
  if scd_api == scd.API_0_3_5:
    req.update({"old_version": 0, "notify_for_operations": True})
  elif scd_api == scd.API_0_3_17:
    req.update({"notify_for_operational_intents": True})
  return req


def _make_sub3_req(scd_api):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=4)
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = LAT0 + scd.latitude_degrees(FOOTPRINT_SPACING_M)
  req = {
    "extents": scd.make_vol4(time_start, time_end, 700, 1000, scd.make_circle(lat, LNG0, 100)),
    "uss_base_url": "https://example.com/foo",
    "notify_for_constraints": False
  }
  if scd_api == scd.API_0_3_5:
    req.update({"old_version": 0, "notify_for_operations": True})
  elif scd_api == scd.API_0_3_17:
    req.update({"notify_for_operational_intents": True})
  return req


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session):
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
      actions.delete_subscription_if_exists(sub_id, scd_session, scd_api)


# Preconditions: No named Subscriptions exist
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_subs_do_not_exist_get(ids, scd_api, scd_session):
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    resp = scd_session.get('/subscriptions/{}'.format(sub_id))
    assert resp.status_code == 404, resp.content


# Preconditions: No named Subscriptions exist
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_subs_do_not_exist_query(ids, scd_api, scd_session):
  resp = scd_session.post('/subscriptions/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
  })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    assert sub_id not in result_ids


# Preconditions: No named Subscriptions exist
# Mutations: Subscriptions 1, 2, and 3 created
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_create_subs(ids, scd_api, scd_session):
  resp = scd_session.put('/subscriptions/{}'.format(ids(SUB1_TYPE)), json=_make_sub1_req(scd_api))
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/subscriptions/{}'.format(ids(SUB2_TYPE)), json=_make_sub2_req(scd_api))
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/subscriptions/{}'.format(ids(SUB3_TYPE)), json=_make_sub3_req(scd_api))
  assert resp.status_code == 200, resp.content


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_search_find_all_subs(ids, scd_api, scd_session):
  resp = scd_session.post(
      '/subscriptions/query',
      json={
        "area_of_interest": scd.make_vol4(None, None, 0, 3000,
                                          scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
      })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    assert sub_id in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_search_footprint(ids, scd_api, scd_session):
  lat = LAT0 - scd.latitude_degrees(FOOTPRINT_SPACING_M)
  print(lat)
  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(None, None, 0, 3000,
                                        scd.make_circle(lat, LNG0, 50))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) in result_ids
  assert ids(SUB2_TYPE) not in result_ids
  assert ids(SUB3_TYPE) not in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(None, None, 0, 3000,
                                        scd.make_circle(LAT0, LNG0, 50))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) not in result_ids
  assert ids(SUB2_TYPE) in result_ids
  assert ids(SUB3_TYPE) not in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_search_time(ids, scd_api, scd_session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=1)

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(time_start, time_end, 0, 3000,
                                        scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) in result_ids
  assert ids(SUB2_TYPE) not in result_ids
  assert ids(SUB3_TYPE) not in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(None, time_end, 0, 3000,
                                        scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) in result_ids
  assert ids(SUB2_TYPE) not in result_ids
  assert ids(SUB3_TYPE) not in result_ids

  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=4)
  time_end = time_start + datetime.timedelta(minutes=1)

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(time_start, time_end, 0, 3000,
                                        scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) not in result_ids
  assert ids(SUB2_TYPE) not in result_ids
  assert ids(SUB3_TYPE) in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(time_start, None, 0, 3000,
                                        scd.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) not in result_ids
  assert ids(SUB2_TYPE) not in result_ids
  assert ids(SUB3_TYPE) in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_search_time_footprint(ids, scd_api, scd_session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(hours=2.5)
  lat = LAT0 + scd.latitude_degrees(FOOTPRINT_SPACING_M)
  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(time_start, time_end, 0, 3000,
                                        scd.make_circle(lat, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert ids(SUB1_TYPE) not in result_ids
  assert ids(SUB2_TYPE) in result_ids
  assert ids(SUB3_TYPE) not in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: Subscriptions 1, 2, and 3 deleted
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_delete_subs(ids, scd_api, scd_session):
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    if scd_api == scd.API_0_3_5:
      resp = scd_session.delete('/subscriptions/{}'.format(sub_id))
    elif scd_api == scd.API_0_3_17:
      resp = scd_session.get('/subscriptions/{}'.format(sub_id))
      assert resp.status_code == 200
      resp = scd_session.delete('/subscriptions/{}/{}'.format(sub_id, resp.json()['subscription']['version']))
    else:
      raise NotImplementedError('Unsupported API version {}'.format(scd_api))
    assert resp.status_code == 200, resp.content
