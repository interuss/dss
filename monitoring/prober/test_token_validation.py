"""Test Authentication validation
  - Try to read DSS without Token
  - Try to read DSS with Token that cannot be decoded
  - Try to read and write DSS with Token missing and wrong Scope

  ASTM Compliance Test: DSS0010_USS_AUTH
  This entire file is used to demonstrate that the DSS requires proper
  authentication tokens to perform actions on the DSS
"""

import datetime
import re

import common

def test_put_isa_with_read_only_scope_token(rogue_session, session, isa2_uuid):
  read_only_token = session.issue_token(['dss.read.identification_service_areas'])
  rogue_session.headers['Authorization'] = f'Bearer {read_only_token}'

  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = rogue_session.put(
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
      })
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
      })
  assert resp.status_code == 200


def test_get_isa_without_token(rogue_session, isa1_uuid):
  resp = rogue_session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'missing token'


def test_get_isa_with_fake_token(rogue_session, isa1_uuid):
  rogue_session.headers['Authorization'] = 'Bearer fake_token'
  resp = rogue_session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 401
  assert resp.json()['message'] == 'token contains an invalid number of segments'


def test_get_isa_without_scope(rogue_session, session, isa1_uuid):
  no_scope_token = session.issue_token([])
  rogue_session.headers['Authorization'] = f'Bearer {no_scope_token}'
  resp = rogue_session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 403
