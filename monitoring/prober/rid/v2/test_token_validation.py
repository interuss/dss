"""Test Authentication validation
  - Try to read DSS without Token
  - Try to read DSS with Token that cannot be decoded
  - Try to read and write DSS with Token missing and wrong Scope

  ASTM Compliance Test: DSS0010_USS_AUTH
  This entire file is used to demonstrate that the DSS requires proper
  authentication tokens to perform actions on the DSS
"""

import datetime

from monitoring.monitorlib import rid_v2
from monitoring.monitorlib.rid_v2 import SCOPE_DP, SCOPE_SP, ISA_PATH
from monitoring.prober.infrastructure import register_resource_type
from . import common


ISA_TYPE = register_resource_type(363, 'ISA')
BASE_URL = 'https://example.com/rid/v2'


def test_ensure_clean_workspace(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_DP)
  if resp.status_code == 200:
    version = resp.json()["service_area"]['version']
    resp = session_ridv2.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_SP)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


def test_put_isa_with_read_only_scope_token(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      }, scope=SCOPE_DP)
  assert resp.status_code == 403, resp.content


def test_create_isa(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      }, scope=SCOPE_SP)
  assert resp.status_code == 200, resp.content


def test_get_isa_without_token(ids, no_auth_session_ridv2):
  resp = no_auth_session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 401, resp.content
  assert 'Missing access token' in resp.json()['message']


def test_get_isa_with_fake_token(ids, no_auth_session_ridv2):
  no_auth_session_ridv2.headers['Authorization'] = 'Bearer fake_token'
  resp = no_auth_session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 401, resp.content
  assert 'token contains an invalid number of segments' in resp.json()['message']


def test_delete(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_DP)
  if resp.status_code == 200:
    version = resp.json()["service_area"]['version']
    resp = session_ridv2.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_SP)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
