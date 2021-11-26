
"""Basic Operation tests:

  - make sure the Operation doesn't exist with get or query
  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC, SCOPE_CI, SCOPE_CM, SCOPE_CP
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import depends_on, for_api_versions, register_resource_type


BASE_URL = 'https://example.com/uss'
OP_TYPE = register_resource_type(341, 'Operational intent')


@for_api_versions(scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  if resp.status_code == 200:
    resp = scd_session.delete('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


def _make_op1_request():
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-56, 178, 50))],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': BASE_URL,
    'new_subscription': {
      'uss_base_url': BASE_URL,
      'notify_for_constraints': False
    }
  }


@default_scope(SCOPE_SC)
@depends_on(test_ensure_clean_workspace)
def test_op_does_not_exist_get(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_ensure_clean_workspace)
def test_op_does_not_exist_query(ids, scd_api, scd_session, scd_session_cp, scd_session_cm):
  time_now = datetime.datetime.utcnow()
  end_time = time_now + datetime.timedelta(hours=1)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
  }, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [op['id'] for op in resp.json().get('operational_intent_reference', [])]

  if scd_session_cp:
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=SCOPE_CP)
    assert resp.status_code == 403, resp.content

  if scd_session_cm:
    resp = scd_session.post('/operational_intent_references/query', json={
      'area_of_interest': scd.make_vol4(time_now, end_time, 0, 5000, scd.make_circle(-56, 178, 300))
    }, scope=SCOPE_CM)
    assert resp.status_code == 403, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_ensure_clean_workspace)
def test_create_op_single_extent(ids, scd_api, scd_session):
  req = _make_op1_request()
  req['extents'] = req['extents'][0]
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_ensure_clean_workspace)
def test_create_op_missing_time_start(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_start']
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_ensure_clean_workspace)
def test_create_op_missing_time_end(ids, scd_api, scd_session):
  req = _make_op1_request()
  del req['extents'][0]['time_end']
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


@depends_on(test_ensure_clean_workspace)
def test_create_op(ids, scd_api, scd_session, scd_session_cp, scd_session_cm):
  req = _make_op1_request()

  if scd_session_cp:
    resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CP)
    assert resp.status_code == 403, resp.content

  if scd_session_cm:
    resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
    assert resp.status_code == 403, resp.content

  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == "Unknown"
  assert_datetimes_are_equal(op['time_start']['value'], req['extents'][0]['time_start']['value'])
  assert_datetimes_are_equal(op['time_end']['value'], req['extents'][0]['time_end']['value'])
  assert op['version'] == 1
  assert 'subscription_id' in op
  assert op['state'] == 'Accepted'


@depends_on(test_create_op)
def test_get_op_by_id(ids, scd_api, scd_session, scd_session_cp, scd_session_cm):
  if scd_session_cp:
    resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CP)
    assert resp.status_code == 403, resp.content

  if scd_session_cm:
    resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_CM)
    assert resp.status_code == 403, resp.content

  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == "Unknown"
  assert op['version'] == 1
  assert 'state' in op
  assert op['state'] == 'Accepted', \
    "The response has a state = '{}'" \
      .format(data['operational_intent_reference']['state'])


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search_missing_params(scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query')
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search(ids, scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json().get('operational_intent_references', [])], resp.json()


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search_earliest_time_included(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operational_intent_references']]


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search_earliest_time_excluded(ids, scd_api, scd_session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=81)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(earliest_time, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search_latest_time_included(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) in [x['id'] for x in resp.json()['operational_intent_references']]


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_id_other_uss(ids, scd_session2):
  resp = scd_session2.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == 'Unknown'
  assert op['version'] == 1
  assert 'state' in op
  assert op['state'] == 'Accepted', \
    "The response has a state = '{}'".format(op['state'])
  assert op.get('ovn', '') in scd.NO_OVN_PHRASES, op


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_query_other_uss(ids, scd_session2):
  resp = scd_session2.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  matching_ops = [x for x in resp.json().get('operational_intent_references', []) if x['id'] == ids(OP_TYPE)]
  assert len(matching_ops) == 1, resp.json()
  op = matching_ops[0]

  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == BASE_URL
  assert op['uss_availability'] == 'Unknown'
  assert op['version'] == 1
  assert 'state' in op
  assert op['state'] == 'Accepted', \
    "The response has a state = '{}'".format(op['state'])
  assert op.get('ovn', '') in scd.NO_OVN_PHRASES, op


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_get_op_by_search_latest_time_excluded(ids, scd_api, scd_session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, latest_time, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]


@default_scope(SCOPE_SC)
@depends_on(test_create_op)
def test_mutate_op(ids, scd_api, scd_session, scd_session_cp, scd_session_cm):
  # GET current op
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_op = resp.json().get('operational_intent_reference', None)
  assert existing_op is not None, resp.json()

  req = _make_op1_request()
  req = {
    'key': [existing_op["ovn"]],
    'extents': req['extents'],
    'old_version': existing_op['version'],
    'state': 'Activated',
    'uss_base_url': 'https://example.com/uss2',
    'subscription_id': existing_op['subscription_id']
  }

  if scd_session_cp:
    resp = scd_session.put(
      '/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CP)
    assert resp.status_code == 403, resp.content

  if scd_session_cm:
    resp = scd_session.put(
      '/operational_intent_references/{}'.format(ids(OP_TYPE)), json=req, scope=SCOPE_CM)
    assert resp.status_code == 403, resp.content

  resp = scd_session.put(
    '/operational_intent_references/{}/{}'.format(ids(OP_TYPE), existing_op["ovn"]), json=req, scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operational_intent_reference']
  assert op['id'] == ids(OP_TYPE)
  assert op['uss_base_url'] == 'https://example.com/uss2'
  assert op['version'] == 2
  assert op['subscription_id'] == existing_op['subscription_id']
  # assert 'state' not in op


@depends_on(test_mutate_op)
def test_delete_op(ids, scd_api, scd_session, scd_session_cp, scd_session_cm):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content
  ovn = resp.json()['operational_intent_reference']['ovn']

  if scd_session_cp:
    resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_CP)
    assert resp.status_code == 403, resp.content

  if scd_session_cm:
    resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_CM)
    assert resp.status_code == 403, resp.content

  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP_TYPE), ovn), scope=SCOPE_SC)
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_delete_op)
def test_get_deleted_op_by_id(ids, scd_api, scd_session):
  resp = scd_session.get('/operational_intent_references/{}'.format(ids(OP_TYPE)))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_SC)
@depends_on(test_delete_op)
def test_get_deleted_op_by_search(ids, scd_api, scd_session):
  resp = scd_session.post('/operational_intent_references/query', json={
    'area_of_interest': scd.make_vol4(None, None, 0, 5000, scd.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert ids(OP_TYPE) not in [x['id'] for x in resp.json()['operational_intent_references']]
