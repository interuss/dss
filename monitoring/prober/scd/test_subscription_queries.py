"""Strategic conflict detection Subscription query tests:

  - add a few Subscriptions spaced in time and footprints
  - query with various combinations of arguments
"""

import datetime

from . import common


LAT0 = 23
LNG0 = 56

# This value should be large enough to ensure areas separated by this distance
# will lie in separate grid cells.
FOOTPRINT_SPACING_M = 10000


def _make_sub1_req():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = LAT0 - common.latitude_degrees(FOOTPRINT_SPACING_M)
  return {
    "extents": common.make_vol4(time_start, time_end, 0, 300, common.make_circle(lat, LNG0, 100)),
    "old_version": 0,
    "uss_base_url": "https://example.com/foo",
    "notify_for_operations": True,
    "notify_for_constraints": False
  }


def _make_sub2_req():
  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=2)
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    "extents": common.make_vol4(time_start, time_end, 350, 650, common.make_circle(LAT0, LNG0, 100)),
    "old_version": 0,
    "uss_base_url": "https://example.com/foo",
    "notify_for_operations": True,
    "notify_for_constraints": False
  }


def _make_sub3_req():
  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=4)
  time_end = time_start + datetime.timedelta(minutes=60)
  lat = LAT0 + common.latitude_degrees(FOOTPRINT_SPACING_M)
  return {
    "extents": common.make_vol4(time_start, time_end, 700, 1000, common.make_circle(lat, LNG0, 100)),
    "old_version": 0,
    "uss_base_url": "https://example.com/foo",
    "notify_for_operations": True,
    "notify_for_constraints": False
  }


# Preconditions: No named Subscriptions exist
# Mutations: None
def test_subs_do_not_exist_get(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  for sub_uuid in (sub1_uuid, sub2_uuid, sub3_uuid):
    resp = scd_session.get('/subscriptions/{}'.format(sub_uuid))
    assert resp.status_code == 404, resp.content


# Preconditions: No named Subscriptions exist
# Mutations: None
def test_subs_do_not_exist_query(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  resp = scd_session.post('/subscriptions/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
  })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  for sub_uuid in (sub1_uuid, sub2_uuid, sub3_uuid):
    assert sub_uuid not in result_ids


# Preconditions: No named Subscriptions exist
# Mutations: Subscriptions 1, 2, and 3 created
def test_create_subs(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  resp = scd_session.put('/subscriptions/{}'.format(sub1_uuid), json=_make_sub1_req())
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/subscriptions/{}'.format(sub2_uuid), json=_make_sub2_req())
  assert resp.status_code == 200, resp.content

  resp = scd_session.put('/subscriptions/{}'.format(sub3_uuid), json=_make_sub3_req())
  assert resp.status_code == 200, resp.content


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
def test_search_find_all_subs(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  resp = scd_session.post(
      '/subscriptions/query',
      json={
        "area_of_interest": common.make_vol4(None, None, 0, 3000,
                                             common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
      })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  for sub_uuid in (sub1_uuid, sub2_uuid, sub3_uuid):
    assert sub_uuid in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
def test_search_footprint(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  lat = LAT0 - common.latitude_degrees(FOOTPRINT_SPACING_M)
  print(lat)
  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(None, None, 0, 3000,
                                           common.make_circle(lat, LNG0, 50))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid in result_ids
  assert sub2_uuid not in result_ids
  assert sub3_uuid not in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(None, None, 0, 3000,
                                           common.make_circle(LAT0, LNG0, 50))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid not in result_ids
  assert sub2_uuid in result_ids
  assert sub3_uuid not in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
def test_search_time(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=1)

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(time_start, time_end, 0, 3000,
                                           common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid in result_ids
  assert sub2_uuid not in result_ids
  assert sub3_uuid not in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(None, time_end, 0, 3000,
                                           common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid in result_ids
  assert sub2_uuid not in result_ids
  assert sub3_uuid not in result_ids

  time_start = datetime.datetime.utcnow() + datetime.timedelta(hours=4)
  time_end = time_start + datetime.timedelta(minutes=1)

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(time_start, time_end, 0, 3000,
                                           common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid not in result_ids
  assert sub2_uuid not in result_ids
  assert sub3_uuid in result_ids

  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(time_start, None, 0, 3000,
                                           common.make_circle(LAT0, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid not in result_ids
  assert sub2_uuid not in result_ids
  assert sub3_uuid in result_ids


# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: None
def test_search_time_footprint(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(hours=2.5)
  lat = LAT0 + common.latitude_degrees(FOOTPRINT_SPACING_M)
  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": common.make_vol4(time_start, time_end, 0, 3000,
                                           common.make_circle(lat, LNG0, FOOTPRINT_SPACING_M))
    })
  assert resp.status_code == 200, resp.content
  result_ids = [x['id'] for x in resp.json()['subscriptions']]
  assert sub1_uuid not in result_ids
  assert sub2_uuid in result_ids
  assert sub3_uuid not in result_ids



# Preconditions: Subscriptions 1, 2, and 3 created
# Mutations: Subscriptions 1, 2, and 3 deleted
def test_delete_subs(scd_session, sub1_uuid, sub2_uuid, sub3_uuid):
  for sub_uuid in (sub1_uuid, sub2_uuid, sub3_uuid):
    resp = scd_session.delete('/subscriptions/{}'.format(sub_uuid))
    assert resp.status_code == 200, resp.content
