"""Subscription update validation tests:

  - make sure Operation doesn't exist by ID
  - create Operation with implicit Subscription
  - make sure implicit Subscription is created
  - try mutate Subscription by shrinking 2d area which does not cover Operation
  - try mutate Subscription by shrinking altitude range which does not cover Operation
  - try mutate Subscription by shrinking time range  which does not cover Operation
  - mutate Subscription with the same 2d area
  - delete Operation
  - delete Subscription
"""

import datetime

from monitoring.monitorlib import scd
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.scd import SCOPE_SC

BASE_URL = 'https://example.com/uss'
OP_ID = '00000073-ff83-443b-aa56-36671e000000'

sub_id = ''


def _make_op_req():
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  return {
    'extents': [scd.make_vol4(time_start, time_end, 0, 1000, scd.make_circle(-56, 178, 500))],
    'old_version': 0,
    'state': 'Accepted',
    'uss_base_url': BASE_URL,
    'new_subscription': {
      'uss_base_url': BASE_URL,
      'notify_for_constraints': False
    }
  }


def _make_sub_req(time_start, time_end, alt_start, alt_end, radius):
  return {
    "extents": scd.make_vol4(time_start, time_end, alt_start, alt_end, scd.make_circle(-56, 178, radius)),
    "old_version": 0,
    "uss_base_url": BASE_URL,
    "notify_for_operations": True,
    "notify_for_constraints": False
  }


@default_scope(SCOPE_SC)
def test_ensure_clean_workspace(scd_session):
  resp = scd_session.get('/operation_references/{}'.format(OP_ID))
  if resp.status_code == 200:
    resp = scd_session.delete('/operation_references/{}'.format(OP_ID))
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


# Create operation normally (also creates implicit Subscription)
# Preconditions: None
# Mutations: Operation OP_ID and its implicit subscription created
@default_scope(SCOPE_SC)
def test_create_op(scd_session):
  req = _make_op_req()
  resp = scd_session.put('/operation_references/{}'.format(OP_ID), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == OP_ID
  assert op['uss_base_url'] == BASE_URL
  assert op['time_start']['value'] == req['extents'][0]['time_start']['value']
  assert op['time_end']['value'] == req['extents'][0]['time_end']['value']
  assert op['version'] == 1
  assert 'subscription_id' in op
  assert 'state' not in op

  # Make sure the implicit Subscription exists when queried separately
  global sub_id
  sub_id = op['subscription_id']

  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content


# Try to mutate subscription by shrinking its 2d area
# Preconditions: Operation OP_ID and subscription sub_id created
# Mutations: None
@default_scope(SCOPE_SC)
def test_mutate_sub_shrink_2d(scd_session):
  if scd_session is None:
    return

  # GET current sub before mutation
  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  req = _make_sub_req(time_start, time_end, 0, 1000, 50)
  req['old_version'] = existing_sub['version']
  req['notify_for_constraints'] = True

  resp = scd_session.put('/subscriptions/{}'.format(sub_id), json=req)
  assert resp.status_code == 400, resp.content


# Try to mutate subscription by shrinking its altitude range
# Preconditions: Operation OP_ID and subscription sub_id created
# Mutations: None
@default_scope(SCOPE_SC)
def test_mutate_sub_shrink_altitude(scd_session):
  if scd_session is None:
    return

  # GET current sub before mutation
  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  req = _make_sub_req(time_start, time_end, 200, 1000, 500)
  req['old_version'] = existing_sub['version']
  req['notify_for_constraints'] = True

  resp = scd_session.put('/subscriptions/{}'.format(sub_id), json=req)
  assert resp.status_code == 400, resp.content


# Try to mutate subscription by shrinking its time range
# Preconditions: Operation OP_ID and subscription sub_id created
# Mutations: None
@default_scope(SCOPE_SC)
def test_mutate_sub_shrink_time(scd_session):
  if scd_session is None:
    return

  # GET current sub before mutation
  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=40)
  req = _make_sub_req(time_start, time_end, 0, 1000, 500)
  req['old_version'] = existing_sub['version']
  req['notify_for_constraints'] = True

  resp = scd_session.put('/subscriptions/{}'.format(sub_id), json=req)
  assert resp.status_code == 400, resp.content


# Mutate sub, with the same 2d area
# Preconditions: Operation OP_ID and subscription sub_id created
# Mutations: Subscription mutated to new version
@default_scope(SCOPE_SC)
def test_mutate_sub_not_shrink(scd_session):
  if scd_session is None:
    return

  # GET current sub before mutation
  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content
  existing_sub = resp.json().get('subscription', None)
  assert existing_sub is not None

  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=20)
  time_end = time_start + datetime.timedelta(minutes=60)
  req = _make_sub_req(time_start, time_end, 0, 1000, 500)
  req['old_version'] = existing_sub['version']
  req['notify_for_constraints'] = True

  resp = scd_session.put('/subscriptions/{}'.format(sub_id), json=req)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert scd.iso8601_equal(data['subscription']['time_start']['value'], req['extents']['time_start']['value'])
  assert scd.iso8601_equal(data['subscription']['time_end']['value'], req['extents']['time_end']['value'])


# Preconditions: Operation OP_ID created
# Mutations: Operation OP_ID deleted
@default_scope(SCOPE_SC)
def test_delete_op(scd_session):
  resp = scd_session.delete('/operation_references/{}'.format(OP_ID))
  assert resp.status_code == 200, resp.content


# Preconditions: Operation OP_ID deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_op_by_id(scd_session):
  resp = scd_session.get('/operation_references/{}'.format(OP_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: Subscription sub_id created
# Mutations: Subscription sub_id deleted
@default_scope(SCOPE_SC)
def test_delete_sub(scd_session):
  if scd_session is None:
    return
  resp = scd_session.delete('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 200, resp.content


# Preconditions: Subscription sub_id deleted
# Mutations: None
@default_scope(SCOPE_SC)
def test_get_deleted_sub_by_id(scd_session):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(sub_id))
  assert resp.status_code == 404, resp.content
