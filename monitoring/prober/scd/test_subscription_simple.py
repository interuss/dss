"""Basic strategic conflict detection Subscription tests:

  - make sure Subscription doesn't exist by ID
  - make sure Subscription doesn't exist by search
  - create the Subscription with a 60 minute expiry
  - get by ID
  - get by searching a circular area
  - delete
  - make sure Subscription can't be found by ID
  - make sure Subscription can't be found by search
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


SUB_TYPE = register_resource_type(220, 'Subscription')


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session):
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/subscriptions/{}'.format(ids(SUB_TYPE)), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


def _make_sub1_req(scd_api):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  req = {
    "extents": scd.make_vol4(time_start, time_end, 0, 1000, scd.make_circle(12, -34, 300)),
    "uss_base_url": "https://example.com/foo",
    "notify_for_constraints": False
  }
  if scd_api == scd.API_0_3_5:
    req.update({"old_version": 0, "notify_for_operations": True})
  elif scd_api == scd.API_0_3_17:
    req.update({"notify_for_operational_intents": True})
  return req


def _check_sub1(data, sub_id, scd_api):
  assert data['subscription']['id'] == sub_id
  assert (('notification_index' not in data['subscription']) or
          (data['subscription']['notification_index'] == 0))
  assert data['subscription']['uss_base_url'] == 'https://example.com/foo'
  assert data['subscription']['time_start']['format'] == scd.TIME_FORMAT_CODE
  assert data['subscription']['time_end']['format'] == scd.TIME_FORMAT_CODE
  assert (('notify_for_constraints' not in data['subscription']) or
          (data['subscription']['notify_for_constraints'] == False))
  assert (('implicit_subscription' not in data['subscription']) or
            (data['subscription']['implicit_subscription'] == False))
  if scd_api == scd.API_0_3_5:
    assert data['subscription']['notify_for_operations'] == True
    assert (('dependent_operations' not in data['subscription'])
            or len(data['subscription']['dependent_operations']) == 0)
  elif scd_api == scd.API_0_3_17:
    assert data['subscription']['notify_for_operational_intents'] == True
    assert (('dependent_operational_intents' not in data['subscription'])
            or len(data['subscription']['dependent_operational_intents']) == 0)


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_sub_does_not_exist_get(ids, scd_api, scd_session):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)))
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_sub_does_not_exist_query(ids, scd_api, scd_session):
  if scd_session is None:
    return
  resp = scd_session.post('/subscriptions/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(12, -34, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(SUB_TYPE) not in [sub['id'] for sub in resp.json().get('subscriptions', [])]


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_create_sub(ids, scd_api, scd_session):
  if scd_session is None:
    return
  req = _make_sub1_req(scd_api)
  resp = scd_session.put('/subscriptions/{}'.format(ids(SUB_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert_datetimes_are_equal(data['subscription']['time_start']['value'], req['extents']['time_start']['value'])
  assert_datetimes_are_equal(data['subscription']['time_end']['value'], req['extents']['time_end']['value'])
  _check_sub1(data, ids(SUB_TYPE), scd_api)


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_sub_by_id(ids, scd_api, scd_session):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  _check_sub1(data, ids(SUB_TYPE), scd_api)


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_sub_by_search(ids, scd_api, scd_session):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()
  resp = scd_session.post(
      '/subscriptions/query',
      json={
        "area_of_interest": scd.make_vol4(time_now, time_now, 0, 120,
                                          scd.make_circle(12.00001, -34.00001, 50))
      })
  if resp.status_code != 200:
    print(resp.content)
  assert resp.status_code == 200, resp.content
  assert ids(SUB_TYPE) in [x['id'] for x in resp.json()['subscriptions']]


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_mutate_sub(ids, scd_api, scd_session):
  if scd_session is None:
    return

  # GET current sub1 before mutation
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  req = _make_sub1_req(scd_api)
  req['notify_for_constraints'] = True

  if scd_api == scd.API_0_3_5:
    req['old_version'] = existing_sub['version']
    resp = scd_session.put('/subscriptions/{}'.format(ids(SUB_TYPE)), json=req)
  elif scd_api == scd.API_0_3_17:
    resp = scd_session.put('/subscriptions/{}/{}'.format(ids(SUB_TYPE), existing_sub['version']), json=req)
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert_datetimes_are_equal(data['subscription']['time_start']['value'], req['extents']['time_start']['value'])
  assert_datetimes_are_equal(data['subscription']['time_end']['value'], req['extents']['time_end']['value'])


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_delete_sub(ids, scd_api, scd_session):
  if scd_session is None:
    return
  if scd_api == scd.API_0_3_5:
    resp = scd_session.delete('/subscriptions/{}'.format(ids(SUB_TYPE)))
  elif scd_api == scd.API_0_3_17:
    resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)))
    assert resp.status_code == 200, resp.content
    resp = scd_session.delete('/subscriptions/{}/{}'.format(ids(SUB_TYPE), resp.json()['subscription']['version']))
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_sub_by_id(ids, scd_api, scd_session):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB_TYPE)))
  assert resp.status_code == 404, resp.content


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_get_deleted_sub_by_search(ids, scd_api, scd_session):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()
  resp = scd_session.post(
    '/subscriptions/query',
    json={
      "area_of_interest": scd.make_vol4(time_now, time_now, 0, 120,
                                        scd.make_circle(12.00001, -34.00001, 50))
    })
  assert resp.status_code == 200, resp.content
  assert ids(SUB_TYPE) not in [x['id'] for x in resp.json()['subscriptions']]
