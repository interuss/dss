"""Test Authentication validation
  - Try to read DSS without Token
  - Try to read DSS with Token that cannot be decoded
  - Try to read and write DSS with Token missing and wrong Scope

  ASTM Compliance Test: DSS0010_USS_AUTH
  This entire file is used to demonstrate that the DSS requires proper
  authentication tokens to perform actions on the DSS
"""

import datetime

from . import common
from .common import SCOPE_READ, SCOPE_WRITE


def test_put_isa_with_read_only_scope_token(session, isa2_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa2_uuid),
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


def test_create_isa(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
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


def test_get_isa_without_token(no_auth_session, isa1_uuid):
  resp = no_auth_session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'missing token'


def test_get_isa_with_fake_token(no_auth_session, isa1_uuid):
  no_auth_session.headers['Authorization'] = 'Bearer fake_token'
  resp = no_auth_session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'token contains an invalid number of segments'


def test_get_isa_without_scope(session, isa1_uuid):
  # TODO: A real OAuth server is unlikely to grant tokens without any scopes.
  # Adapt this test to work on a real OAuth server, or remove.
  resp = session.get('/identification_service_areas/{}'.format(isa1_uuid), scope='')
  assert resp.status_code == 403
