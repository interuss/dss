"""ISA input validation tests:

  - check we can't create the ISA with a huge area
  - check we can't create the ISA with missing fields
  - check we can't create the ISA with a time_start in the past
  - check we can't create the ISA with a time_start after time_end
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid_v2
from monitoring.monitorlib.rid_v2 import SCOPE_DP, SCOPE_SP, ISA_PATH
from monitoring.prober.infrastructure import register_resource_type
from . import common


ISA_TYPE = register_resource_type(366, 'ISA')
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


@default_scope(SCOPE_SP)
def test_isa_huge_area(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': common.HUGE_VERTICES,
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
  assert 'too large' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_empty_vertices(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': [],
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
  assert 'Not enough points in polygon' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_missing_outline(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
  assert 'Error parsing Volume4D' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_missing_volume(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
  assert 'Error parsing Volume4D' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_missing_extents(ids, session_ridv2):
  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
  assert 'Error parsing Volume4D: Neither outline_polygon nor outline_circle were specified in volume' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_start_time_in_past(ids, session_ridv2):
  time_start = datetime.datetime.utcnow() - datetime.timedelta(minutes=10)
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
      })
  assert resp.status_code == 400, resp.content
  assert 'IdentificationServiceArea time_start must not be in the past' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_start_time_after_time_end(ids, session_ridv2):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
  time_end = time_start - datetime.timedelta(minutes=5)

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
      })
  assert resp.status_code == 400, resp.content
  assert 'IdentificationServiceArea time_end must be after time_start' in resp.json()['message']


@default_scope(SCOPE_SP)
def test_isa_not_on_earth(ids, session_ridv2):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': [
                        {'lat': 130.6205, 'lng': -23.6558},
                        {'lat': 130.6301, 'lng': -23.6898},
                        {'lat': 130.6700, 'lng': -23.6709},
                        {'lat': 130.6466, 'lng': -23.6407},
                      ],
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
              'time_start': rid_v2.Time.make(time_start),
              'time_end': rid_v2.Time.make(time_end),
          },
          'uss_base_url': BASE_URL,
      })
  assert resp.status_code == 400, resp.content
