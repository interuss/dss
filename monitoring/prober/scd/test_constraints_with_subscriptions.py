"""Basic Constraint-Subscription interaction tests:

  - make sure the Constraint doesn't exist with get or query
  - create the Constraint with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime
from typing import Dict

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_CI, SCOPE_CM, SCOPE_SC, SCOPE_CP
from monitoring.prober.infrastructure import for_api_versions, register_resource_type
from monitoring.prober.scd import actions


CONSTRAINT_BASE_URL_1 = 'https://example.com/con1/uss'
CONSTRAINT_BASE_URL_2 = 'https://example.com/con2/uss'
CONSTRAINT_BASE_URL_3 = 'https://example.com/con3/uss'
SUB_BASE_URL_A = 'https://example.com/sub1/uss'
SUB_BASE_URL_B = 'https://example.com/sub2/uss'

CONSTRAINT_TYPE = register_resource_type(2, 'Single constraint')
SUB1_TYPE = register_resource_type(3, 'Constraint subscription 1')
SUB2_TYPE = register_resource_type(4, 'Constraint subscription 2')
SUB3_TYPE = register_resource_type(5, 'Constraint subscription 3')


def _make_c1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 120, scd.make_circle(-12.00001, 33.99999, 50))],
    'old_version': 0,
    'uss_base_url': CONSTRAINT_BASE_URL_1,
  }


def _make_sub_req(base_url: str, notify_ops: bool, notify_constraints: bool) -> Dict:
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    "extents": scd.make_vol4(time_start, time_end, 0, 1000, scd.make_circle(-12, 34, 300)),
    "old_version": 0,
    "uss_base_url": base_url,
    "notify_for_operations": notify_ops,
    "notify_for_operational_intents": notify_ops,
    "notify_for_constraints": notify_constraints
  }


def _read_both_scope(scd_api: str) -> str:
  if scd_api == scd.API_0_3_5:
    return '{} {}'.format(SCOPE_SC, SCOPE_CI)
  elif scd_api == scd.API_0_3_17:
    return '{} {}'.format(SCOPE_SC, SCOPE_CP)
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))


def _read_ops_scope(scd_api: str) -> str:
  return SCOPE_SC


_read_subs_scope = _read_ops_scope


def _read_constraints_scope(scd_api: str) -> str:
  if scd_api == scd.API_0_3_5:
    return SCOPE_CI
  elif scd_api == scd.API_0_3_17:
    return SCOPE_CP
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session, scd_session2):
  actions.delete_constraint_reference_if_exists(ids(CONSTRAINT_TYPE), scd_session, scd_api)

  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    actions.delete_subscription_if_exists(sub_id, scd_session2, scd_api)


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_subs_do_not_exist(ids, scd_api, scd_session, scd_session2):
  if scd_session is None:
    return

  sub_scope = _read_subs_scope(scd_api)

  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB1_TYPE)), scope=sub_scope)
  assert resp.status_code == 404, resp.content
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB2_TYPE)), scope=sub_scope)
  assert resp.status_code == 404, resp.content
  resp = scd_session.get('/subscriptions/{}'.format(ids(SUB3_TYPE)), scope=sub_scope)
  assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: {Sub1, Sub2, Sub3} created by scd_session2 user
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_create_subs(ids, scd_api, scd_session, scd_session2):
  if scd_session2 is None:
    return

  req = _make_sub_req(SUB_BASE_URL_A, notify_ops=True, notify_constraints=False)
  resp = scd_session2.put('/subscriptions/{}'.format(ids(SUB1_TYPE)), json=req, scope=_read_ops_scope(scd_api))
  assert resp.status_code == 200, resp.content

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=False, notify_constraints=True)
  resp = scd_session2.put('/subscriptions/{}'.format(ids(SUB2_TYPE)), json=req, scope=_read_constraints_scope(scd_api))
  assert resp.status_code == 200, resp.content

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=True, notify_constraints=True)
  resp = scd_session2.put('/subscriptions/{}'.format(ids(SUB3_TYPE)), json=req, scope=_read_both_scope(scd_api))
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_constraint_does_not_exist(ids, scd_api, scd_session, scd_session2):
  resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)))
  assert resp.status_code == 404, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: Constraint ids(CONSTRAINT_ID) created by scd_session user
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_create_constraint(ids, scd_api, scd_session, scd_session2):
  req = _make_c1_request()
  resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']
  assert SUB_BASE_URL_A not in [subscriber['uss_base_url'] for subscriber in subscribers], subscribers
  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  assert ids(SUB2_TYPE) in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert ids(SUB3_TYPE) in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == ids(SUB2_TYPE)][0]
  assert sub2_index == 1, subscriberb
  sub3_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == ids(SUB3_TYPE)][0]
  assert sub3_index == 1, subscriberb


# Preconditions:
#   * Sub1 created by scd_session2 user
#   * {Sub2, Sub3} received one notification
#   * Constraint ids(CONSTRAINT_ID) created by scd_session user
# Mutations: Constraint ids(CONSTRAINT_ID) mutated to second version
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_mutate_constraint(ids, scd_api, scd_session, scd_session2):
  # GET current constraint
  resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), scope=_read_constraints_scope(scd_api))
  assert resp.status_code == 200, resp.content
  existing_constraint = resp.json().get('constraint_reference', None)
  assert existing_constraint is not None

  req = _make_c1_request()
  req = {
    'key': [existing_constraint['ovn']],
    'extents': req['extents'],
    'old_version': existing_constraint['version'],
    'uss_base_url': CONSTRAINT_BASE_URL_2
  }

  if scd_api == scd.API_0_3_5:
    resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req, scope=SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    resp = scd_session.put('/constraint_references/{}/{}'.format(ids(CONSTRAINT_TYPE), existing_constraint['ovn']), json=req, scope=SCOPE_CM)
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']
  assert SUB_BASE_URL_A not in [subscriber['uss_base_url'] for subscriber in subscribers], subscribers
  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  assert ids(SUB2_TYPE) in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert ids(SUB3_TYPE) in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == ids(SUB2_TYPE)][0]
  assert sub2_index == 2, subscriberb
  sub3_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == ids(SUB3_TYPE)][0]
  assert sub3_index == 2, subscriberb


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: Sub1 listens for Constraints, Sub3 doesn't listen for Constraints
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_mutate_subs(ids, scd_api, scd_session2, scd_session):
  # GET current sub1 before mutation
  resp = scd_session2.get('/subscriptions/{}'.format(ids(SUB1_TYPE)), scope=_read_subs_scope(scd_api))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  req = _make_sub_req(SUB_BASE_URL_A, notify_ops=True, notify_constraints=True)
  req['old_version'] = existing_sub['version']
  if scd_api == scd.API_0_3_5:
    resp = scd_session2.put('/subscriptions/{}'.format(ids(SUB1_TYPE)), json=req, scope=_read_both_scope(scd_api))
    key = 'constraints'
  elif scd_api == scd.API_0_3_17:
    resp = scd_session2.put('/subscriptions/{}/{}'.format(ids(SUB1_TYPE), existing_sub['version']), json=req, scope=_read_both_scope(scd_api))
    key = 'constraint_references'
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert ids(CONSTRAINT_TYPE) in [constraint['id'] for constraint in data.get(key, [])], data

  # GET current sub3 before mutation
  resp = scd_session2.get('/subscriptions/{}'.format(ids(SUB3_TYPE)), scope=_read_subs_scope(scd_api))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=True, notify_constraints=False)
  req['old_version'] = existing_sub['version']

  if scd_api == scd.API_0_3_5:
    resp = scd_session2.put('/subscriptions/{}'.format(ids(SUB3_TYPE)), json=req, scope=_read_both_scope(scd_api))
  elif scd_api == scd.API_0_3_17:
    resp = scd_session2.put('/subscriptions/{}/{}'.format(ids(SUB3_TYPE), existing_sub['version']), json=req, scope=_read_both_scope(scd_api))
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  if scd_api == scd.API_0_3_5:
    assert not data.get('constraints', []), data
  elif scd_api == scd.API_0_3_17:
    assert not data.get('constraint_references', []), data
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))


# Preconditions:
#   * Sub1 mutated by scd_session2 user to receive Constraints
#   * Sub2 received one notification
#   * Sub3 received one notification and mutated by scd_session2 user to not receive Constraints
#   * Constraint ids(CONSTRAINT_ID) mutated by scd_session user to second version
# Mutations: Constraint ids(CONSTRAINT_ID) mutated to third version
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_mutate_constraint2(ids, scd_api, scd_session, scd_session2):
  # GET current constraint
  resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)))
  assert resp.status_code == 200, resp.content
  existing_constraint = resp.json().get('constraint_reference', None)
  assert existing_constraint is not None

  req = _make_c1_request()
  req = {
    'key': [existing_constraint['ovn']],
    'extents': req['extents'],
    'old_version': existing_constraint['version'],
    'uss_base_url': CONSTRAINT_BASE_URL_3
  }

  if scd_api == scd.API_0_3_5:
    resp = scd_session.put('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)), json=req, scope=SCOPE_CM)
  elif scd_api == scd.API_0_3_17:
    resp = scd_session.put('/constraint_references/{}/{}'.format(ids(CONSTRAINT_TYPE), existing_constraint['ovn']), json=req, scope=SCOPE_CM)
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']

  subscribera = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_A]
  assert len(subscribera) == 1, subscribers
  subscribera = subscribera[0]
  subscribera_subscriptions = [subscription['subscription_id'] for subscription in subscribera['subscriptions']]
  assert ids(SUB1_TYPE) in subscribera_subscriptions
  assert ids(SUB2_TYPE) not in subscribera_subscriptions
  assert ids(SUB3_TYPE) not in subscribera_subscriptions
  sub1_index = [subscription['notification_index'] for subscription in subscribera['subscriptions']
                if subscription['subscription_id'] == ids(SUB1_TYPE)][0]
  assert sub1_index == 1, subscribera

  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  subscriberb_subscriptions = [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert ids(SUB1_TYPE) not in subscriberb_subscriptions
  assert ids(SUB2_TYPE) in subscriberb_subscriptions
  assert ids(SUB3_TYPE) not in subscriberb_subscriptions
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == ids(SUB2_TYPE)][0]
  assert sub2_index == 3, subscriberb


# Preconditions: Constraint ids(CONSTRAINT_ID) mutated to second version
# Mutations: Constraint ids(CONSTRAINT_ID) deleted
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
@default_scope(SCOPE_CM)
def test_delete_constraint(ids, scd_api, scd_session, scd_session2):
  if scd_api == scd.API_0_3_5:
    resp = scd_session.delete('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)))
  elif scd_api == scd.API_0_3_17:
    resp = scd_session.get('/constraint_references/{}'.format(ids(CONSTRAINT_TYPE)))
    assert resp.status_code == 200, resp.content
    existing_constraint = resp.json().get('constraint_reference', None)
    resp = scd_session.delete('/constraint_references/{}/{}'.format(ids(CONSTRAINT_TYPE), existing_constraint['ovn']))
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: {Sub1, Sub2, Sub3} deleted
@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_delete_subs(ids, scd_api, scd_session2, scd_session):
  if scd_session2 is None:
    return
  for sub_id in (ids(SUB1_TYPE), ids(SUB2_TYPE), ids(SUB3_TYPE)):
    if scd_api == scd.API_0_3_5:
      resp = scd_session2.delete('/subscriptions/{}'.format(sub_id), scope=SCOPE_CI)
    elif scd_api == scd.API_0_3_17:
      resp = scd_session2.get('/subscriptions/{}'.format(sub_id), scope=_read_both_scope(scd_api))
      assert resp.status_code == 200, resp.content
      sub = resp.json().get('subscription', None)
      resp = scd_session2.delete('/subscriptions/{}/{}'.format(sub_id, sub['version']), scope=_read_both_scope(scd_api))
    else:
      raise NotImplementedError('Unsupported API version {}'.format(scd_api))
    assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_5, scd.API_0_3_17)
def test_final_cleanup(ids, scd_api, scd_session, scd_session2):
    test_ensure_clean_workspace(ids, scd_api, scd_session, scd_session2)
