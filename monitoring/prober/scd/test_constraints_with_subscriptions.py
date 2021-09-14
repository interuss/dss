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
from monitoring.monitorlib.scd import SCOPE_CI, SCOPE_CM, SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions


CONSTRAINT_BASE_URL_1 = 'https://example.com/con1/uss'
CONSTRAINT_BASE_URL_2 = 'https://example.com/con2/uss'
CONSTRAINT_BASE_URL_3 = 'https://example.com/con3/uss'
SUB_BASE_URL_A = 'https://example.com/sub1/uss'
SUB_BASE_URL_B = 'https://example.com/sub2/uss'

CONSTRAINT_ID = ''
SUB1_ID = ''
SUB2_ID = ''
SUB3_ID = ''


def _bin_to_hex(bin_string):
  return format(int(bin_string,2), "02x")


def _bin_to_dec(bin_string):
  return int(bin_string, 2)


def _hex_to_bin(hex_string):
  return format(int(hex_string,16), "08b")


def _dec_to_bin(num):
  return format(num, "06b")


def _split_by(string_val, num=8):
  """Splits a string into substrings with a length of given number."""
  return [string_val[i:i+num] for i in range(0, len(string_val), num)]


def _get_ord_val(letter):
  """Encodes new ord value for ascii letters."""
  ord_val = ord(letter)
  if ord_val >= 48 and ord_val <= 57: # decimal numbers
    return ord_val - 48
  if ord_val >= 65 and ord_val <= 90: # capitals
    return ord_val - 65 + 10
  if ord_val >= 97 and ord_val <= 122: # small letters.
    return ord_val - 97 + 10 + 26
  if ord_val == 95:
    return 63


def _get_ascii_val_from_bit_value(num):
  """Decodes new ord value to ascii letters."""
  if num >=0 and num <= 9:
    return chr(num + 48)
  if num >= 10 and num <= 35:
    return chr(num + 65 - 10)
  if num >= 36 and num <= 61:
    return chr(num + 97 - 10 - 26)
  if num == 63:
    return '_'


def _encode_owner(string_val, fixed_id):
  bits = ''
  for letter in string_val:
    ord_val = _get_ord_val(letter)
    bits += _dec_to_bin(ord_val)
  hex_codes = ''.join((_bin_to_hex(s) for s in _split_by(bits)[:6]))
  fixed_code = list(fixed_id)
  curr_pos = 16
  hex_ptr = 0
  while curr_pos <= 30 and hex_ptr < len(hex_codes):
    if fixed_code[curr_pos] == '-':
      curr_pos += 1
    fixed_code[curr_pos] = hex_codes[hex_ptr]
    curr_pos += 1
    hex_ptr += 1
  return ''.join(fixed_code)


def _decode_owner(owner_id):
  if len(owner_id) < 30:
    raise ValueError('Invalid owner id.')
  owner_hex_code = (owner_id[16:30]).replace('-', '')
  hex_splits = _split_by(owner_hex_code, num=2)
  bits = ''
  for h in hex_splits:
    print(f'h: {h}, bits: {_hex_to_bin(h)}')
    bits += _hex_to_bin(h)
  test_owner = ''
  for seq in _split_by(bits, 6):
    num = _bin_to_dec(seq)
    test_owner += _get_ascii_val_from_bit_value(num)
  return test_owner


def test_set_test_owner_ids(test_owner):
  print(f'test_owner: {test_owner}')
  global CONSTRAINT_ID
  global SUB1_ID
  global SUB2_ID
  global SUB3_ID
  CONSTRAINT_ID = _encode_owner(test_owner, '000000a2-2629-49c9-a688-23afb3000000')
  SUB1_ID = _encode_owner(test_owner, '00000007-e548-48bb-b9f2-68e0e0000000')
  SUB2_ID = _encode_owner(test_owner, '00000068-6289-46cc-a402-fbc0f7000000')
  SUB3_ID = _encode_owner(test_owner, '00000089-b954-4d3f-8afa-2c4e3b000000')


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
    "notify_for_constraints": notify_constraints
  }


@for_api_versions(scd.API_0_3_5)
def test_ensure_clean_workspace(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CI)
def test_subs_do_not_exist(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CI)
def test_create_subs(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CM)
def test_constraint_does_not_exist(scd_api, scd_session, scd_session2):
  resp = scd_session.get('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 404, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: Constraint CONSTRAINT_ID created by scd_session user
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CM)
def test_create_constraint(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CM)
def test_mutate_constraint(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CI)
def test_mutate_subs(scd_api, scd_session2, scd_session):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CM)
def test_mutate_constraint2(scd_api, scd_session, scd_session2):
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
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CM)
def test_delete_constraint(scd_api, scd_session, scd_session2):
  resp = scd_session.delete('/constraint_references/{}'.format(CONSTRAINT_ID))
  assert resp.status_code == 200, resp.content


# Preconditions: {Sub1, Sub2, Sub3} created by scd_session2 user
# Mutations: {Sub1, Sub2, Sub3} deleted
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_CI)
def test_delete_subs(scd_api, scd_session2, scd_session):
  if scd_session2 is None:
    return
  for sub_id in (SUB1_ID, SUB2_ID, SUB3_ID):
    resp = scd_session2.delete('/subscriptions/{}'.format(sub_id))
    assert resp.status_code == 200, resp.content
