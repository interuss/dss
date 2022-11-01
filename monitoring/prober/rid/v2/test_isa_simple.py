"""Basic ISA tests:

  - create the ISA with a 60 minute expiry
  - get by ID
  - search with earliest_time and latest_time
  - delete
"""

import datetime
import re

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid_v2
from monitoring.monitorlib.rid_v2 import SCOPE_DP, SCOPE_SP, ISA_PATH
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import register_resource_type
from . import common


ISA_TYPE = register_resource_type(348, 'ISA')
BASE_URL_V1 = 'https://example.com/rid/v2'
BASE_URL_V2 = 'https://s2.example.com/rid/v2'

def test_ensure_clean_workspace(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_DP)
  if resp.status_code == 200:
    version = resp.json()['service_area']['version']
    resp = session_ridv2.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_SP)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_SP)
def test_create_isa(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_A_PUT_ISA."""
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  req_body = {
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
    'uss_base_url': BASE_URL_V1,
  }
  resp = session_ridv2.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
      json=req_body)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['service_area']['id'] == ids(ISA_TYPE)
  assert data['service_area']['uss_base_url'] == BASE_URL_V1
  assert_datetimes_are_equal(data['service_area']['time_start']['value'], req_body['extents']['time_start']['value'])
  assert_datetimes_are_equal(data['service_area']['time_end']['value'], req_body['extents']['time_end']['value'])
  assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
  assert 'subscribers' in data


@default_scope(SCOPE_DP)
def test_get_isa_by_id(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['service_area']['id'] == ids(ISA_TYPE)
  assert data['service_area']['uss_base_url'] == BASE_URL_V1


@default_scope(SCOPE_SP)
def test_update_isa(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_DP)
  assert resp.status_code == 200, resp.content
  version = resp.json()['service_area']['version']

  resp = session_ridv2.put(
    '{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version),
      json={
          'extents': {
              'volume': {
                  'outline_polygon': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lower': rid_v2.Altitude.make(20),
                  'altitude_upper': rid_v2.Altitude.make(400),
              },
          },
          'uss_base_url': BASE_URL_V2,
      })
  assert resp.status_code == 200

  data = resp.json()
  assert data['service_area']['id'] == ids(ISA_TYPE)
  assert data['service_area']['uss_base_url'] == BASE_URL_V2
  assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
  assert 'subscribers' in data


@default_scope(SCOPE_DP)
def test_get_isa_by_id_after_update(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['service_area']['id'] == ids(ISA_TYPE)
  assert data['service_area']['uss_base_url'] == BASE_URL_V2


@default_scope(SCOPE_DP)
def test_get_isa_by_search_missing_params(session_ridv2):
  resp = session_ridv2.get(ISA_PATH)
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_DP)
def test_get_isa_by_search(ids, session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_earliest_time_included(ids, session_ridv2):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = session_ridv2.get('{}?area={}&earliest_time={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING,
      earliest_time.strftime(rid_v2.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_earliest_time_excluded(ids, session_ridv2):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = session_ridv2.get('{}?area={}&earliest_time={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING,
      earliest_time.strftime(rid_v2.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_latest_time_included(ids, session_ridv2):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = session_ridv2.get('{}?area={}&latest_time={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING,
      latest_time.strftime(rid_v2.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_latest_time_excluded(ids, session_ridv2):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = session_ridv2.get('{}?area={}&latest_time={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING,
      latest_time.strftime(rid_v2.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_area_only(ids, session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(ISA_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_by_search_huge_area(session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(ISA_PATH, common.HUGE_GEO_POLYGON_STRING))
  assert resp.status_code == 413, resp.content


@default_scope(SCOPE_SP)
def test_delete_isa_wrong_version(ids, session_ridv2):
  resp = session_ridv2.delete('{}/{}/fake_version'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_SP)
def test_delete_isa_empty_version(ids, session_ridv2):
  resp = session_ridv2.delete('{}/{}/'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 400, resp.content


def test_delete_isa(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_B_DELETE_ISA."""
  # GET the ISA first to find its version.
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_DP)
  assert resp.status_code == 200, resp.content
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session_ridv2.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_SP)
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_DP)
def test_get_deleted_isa_by_id(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_DP)
def test_get_deleted_isa_by_search(ids, session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(ISA_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_DP)
def test_get_isa_search_area_with_loop(session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(ISA_PATH, common.LOOP_GEO_POLYGON_STRING))
  assert resp.status_code == 400, resp.content
