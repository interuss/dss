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

from ..infrastructure import default_scope
from . import common
from .common import SCOPE_CI, SCOPE_CM, SCOPE_SC


CONSTRAINT_BASE_URL_1 = 'https://example.com/con1/uss'
CONSTRAINT_BASE_URL_2 = 'https://example.com/con2/uss'
CONSTRAINT_BASE_URL_3 = 'https://example.com/con3/uss'
SUB_BASE_URL_A = 'https://example.com/sub1/uss'
SUB_BASE_URL_B = 'https://example.com/sub2/uss'
CONSTRAINT_ID = '000000a2-2629-49c9-a688-23afb3000000'
SUB1_ID = '00000007-e548-48bb-b9f2-68e0e0000000'
SUB2_ID = '00000068-6289-46cc-a402-fbc0f7000000'
SUB3_ID = '00000089-b954-4d3f-8afa-2c4e3b000000'


def _make_c1_request():
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [common.make_vol4(time_start, time_end, 0, 120, common.make_circle(-12.00001, 33.99999, 50))],
    'old_version': 0,
    'uss_base_url': CONSTRAINT_BASE_URL_1,
  }


def _make_sub_req(base_url: str, notify_ops: bool, notify_constraints: bool) -> Dict:
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    "extents": common.make_vol4(time_start, time_end, 0, 1000, common.make_circle(-12, 34, 300)),
    "old_version": 0,
    "uss_base_url": base_url,
    "notify_for_operations": notify_ops,
    "notify_for_constraints": notify_constraints
  }


def test_ensure_clean_workspace(scd_session, scd_session2):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
  if resp.status_code == 200:
    resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID), scope=SCOPE_CM)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content

  for sub_id in (SUB1_ID, SUB2_ID, SUB3_ID):
    resp = scd_session2.get('/subscriptions/{}'.format(sub_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session2.delete('/subscriptions/{}'.format(sub_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_CI)
def test_subs_do_not_exist(scd_session, scd_session2):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(SUB1_ID))
  assert resp.status_code == 404, resp.content
  resp = scd_session.get('/subscriptions/{}'.format(SUB2_ID))
  assert resp.status_code == 404, resp.content
  resp = scd_session.get('/subscriptions/{}'.format(SUB3_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: None
# Mutations: {Sub1, Sub2, Sub3} created by scd_session2 user
@default_scope(SCOPE_CI)
def test_create_subs(scd_session, scd_session2):
  if scd_session2 is None:
    return
  req = _make_sub_req(SUB_BASE_URL_A, notify_ops=True, notify_constraints=False)
  resp = scd_session2.put('/subscriptions/{}'.format(SUB1_ID), json=req)
  assert resp.status_code == 200, resp.content

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=False, notify_constraints=True)
  resp = scd_session2.put('/subscriptions/{}'.format(SUB2_ID), json=req)
  assert resp.status_code == 200, resp.content

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=True, notify_constraints=True)
  resp = scd_session2.put('/subscriptions/{}'.format(SUB3_ID), json=req)
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_CM)
def test_constraint_does_not_exist(scd_session, scd_session2):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: Constraint CONSTRAINT_ID created by scd_session user
@default_scope(SCOPE_CM)
def test_create_constraint(scd_session, scd_session2):
  req = _make_c1_request()
  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']
  assert SUB_BASE_URL_A not in [subscriber['uss_base_url'] for subscriber in subscribers], subscribers
  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  assert SUB2_ID in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert SUB3_ID in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == SUB2_ID][0]
  assert sub2_index == 1, subscriberb
  sub3_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == SUB3_ID][0]
  assert sub3_index == 1, subscriberb


# Preconditions:
#   * Sub1 created by scd_session2 user
#   * {Sub2, Sub3} received one notification
#   * Constraint CONSTRAINT_ID created by scd_session user
# Mutations: Constraint CONSTRAINT_ID mutated to second version
@default_scope(SCOPE_CM)
def test_mutate_constraint(scd_session, scd_session2):
  # GET current constraint
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 200, resp.content
  existing_constraint = resp.json().get('constraint_reference', None)
  assert existing_constraint is not None

  req = _make_c1_request()
  req = {
    'key': [existing_constraint["ovn"]],
    'extents': req['extents'],
    'old_version': existing_constraint['version'],
    'uss_base_url': CONSTRAINT_BASE_URL_2
  }

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']
  assert SUB_BASE_URL_A not in [subscriber['uss_base_url'] for subscriber in subscribers], subscribers
  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  assert SUB2_ID in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert SUB3_ID in [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == SUB2_ID][0]
  assert sub2_index == 2, subscriberb
  sub3_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == SUB3_ID][0]
  assert sub3_index == 2, subscriberb


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: Sub1 listens for Constraints, Sub3 doesn't listen for Constraints
@default_scope(SCOPE_CI)
def test_mutate_subs(scd_session2, scd_session):
  # GET current sub1 before mutation
  resp = scd_session2.get('/subscriptions/{}'.format(SUB1_ID))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  req = _make_sub_req(SUB_BASE_URL_A, notify_ops=True, notify_constraints=True)
  req['old_version'] = existing_sub['version']
  resp = scd_session2.put('/subscriptions/{}'.format(SUB1_ID), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert CONSTRAINT_ID in [constraint['id'] for constraint in data.get('constraints', [])], data

  # GET current sub3 before mutation
  resp = scd_session2.get('/subscriptions/{}'.format(SUB3_ID))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  req = _make_sub_req(SUB_BASE_URL_B, notify_ops=True, notify_constraints=False)
  req['old_version'] = existing_sub['version']
  resp = scd_session2.put('/subscriptions/{}'.format(SUB3_ID), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert not data.get('constraints', []), data


# Preconditions:
#   * Sub1 mutated by scd_session2 user to receive Constraints
#   * Sub2 received one notification
#   * Sub3 received one notification and mutated by scd_session2 user to not receive Constraints
#   * Constraint CONSTRAINT_ID mutated by scd_session user to second version
# Mutations: Constraint CONSTRAINT_ID mutated to third version
@default_scope(SCOPE_CM)
def test_mutate_constraint2(scd_session, scd_session2):
  # GET current constraint
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 200, resp.content
  existing_constraint = resp.json().get('constraint_reference', None)
  assert existing_constraint is not None

  req = _make_c1_request()
  req = {
    'key': [existing_constraint["ovn"]],
    'extents': req['extents'],
    'old_version': existing_constraint['version'],
    'uss_base_url': CONSTRAINT_BASE_URL_3
  }

  resp = scd_session.put('/constraint_references/{}'.format(CONSTRAINT_ID), json=req, scope=SCOPE_CM)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  subscribers = data['subscribers']

  subscribera = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_A]
  assert len(subscribera) == 1, subscribers
  subscribera = subscribera[0]
  subscribera_subscriptions = [subscription['subscription_id'] for subscription in subscribera['subscriptions']]
  assert SUB1_ID in subscribera_subscriptions
  assert SUB2_ID not in subscribera_subscriptions
  assert SUB3_ID not in subscribera_subscriptions
  sub1_index = [subscription['notification_index'] for subscription in subscribera['subscriptions']
                if subscription['subscription_id'] == SUB1_ID][0]
  assert sub1_index == 1, subscribera

  subscriberb = [subscriber for subscriber in subscribers if subscriber['uss_base_url'] == SUB_BASE_URL_B]
  assert len(subscriberb) == 1, subscribers
  subscriberb = subscriberb[0]
  subscriberb_subscriptions = [subscription['subscription_id'] for subscription in subscriberb['subscriptions']]
  assert SUB1_ID not in subscriberb_subscriptions
  assert SUB2_ID in subscriberb_subscriptions
  assert SUB3_ID not in subscriberb_subscriptions
  sub2_index = [subscription['notification_index'] for subscription in subscriberb['subscriptions']
                if subscription['subscription_id'] == SUB2_ID][0]
  assert sub2_index == 3, subscriberb


# Preconditions: Constraint CONSTRAINT_ID mutated to second version
# Mutations: Constraint CONSTRAINT_ID deleted
@default_scope(SCOPE_CM)
def test_delete_constraint(scd_session, scd_session2):
  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 200, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: {Sub1, Sub2, Sub3} deleted
@default_scope(SCOPE_CI)
def test_delete_subs(scd_session2, scd_session):
  if scd_session2 is None:
    return
  for sub_id in (SUB1_ID, SUB2_ID, SUB3_ID):
    resp = scd_session2.delete('/subscriptions/{}'.format(sub_id))
    assert resp.status_code == 200, resp.content
