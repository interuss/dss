"""Test Authentication validation
  - Try to read DSS without Token
  - Try to read DSS with Token that cannot be decoded
  - Try to read and write DSS with Token missing and wrong Scope

  ASTM Compliance Test: DSS0010_USS_AUTH
  This entire file is used to demonstrate that the DSS requires proper
  authentication tokens to perform actions on the DSS
"""

import datetime

import pytest

from ..auth import DummyOAuth
from . import common
from .common import SCOPE_READ, SCOPE_WRITE

ISA_ID = '000000cc-e2e1-49e6-9102-b2a544000000'


def test_ensure_clean_workspace(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()["service_area"]['version']
    resp = session.delete('/identification_service_areas/{}/{}'.format(ISA_ID, version), scope=SCOPE_WRITE)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


def test_put_isa_with_read_only_scope_token(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(ISA_ID),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
      }, scope=SCOPE_READ)
  assert resp.status_code == 403


def test_create_isa(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(ISA_ID),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
      }, scope=SCOPE_WRITE)
  assert resp.status_code == 200


def test_get_isa_without_token(no_auth_session):
  resp = no_auth_session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'missing token'


def test_get_isa_with_fake_token(no_auth_session):
  no_auth_session.headers['Authorization'] = 'Bearer fake_token'
  resp = no_auth_session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'token contains an invalid number of segments'


def test_get_isa_without_scope(session):
  if not isinstance(session.auth_adapter, DummyOAuth):
    pytest.skip('General auth providers will not usually grant tokens without any scopes')
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope='')
  assert resp.status_code == 403
