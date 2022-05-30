"""Basic subscription tests:

  - create the subscription with a 60 minute expiry
  - get by ID
  - get by search
  - delete
"""

import datetime
import re

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid_v2
from monitoring.monitorlib.rid_v2 import SCOPE_DP, SUBSCRIPTION_PATH
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import register_resource_type
from . import common


SUB_TYPE = register_resource_type(349, 'Subscription')
BASE_URL = 'https://example.com/rid/v2'


def test_ensure_clean_workspace(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)), scope=SCOPE_DP)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session_ridv2.delete('{}/{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE), version), scope=SCOPE_DP)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_DP)
def test_sub_does_not_exist(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 404, resp.content
  assert 'Subscription {} not found'.format(ids(SUB_TYPE)) in resp.json()['message']


@default_scope(SCOPE_DP)
def test_create_sub(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_C_PUT_SUB."""
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
    'uss_base_url': BASE_URL
  }
  resp = session_ridv2.put(
      '{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)),
      json=req_body)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['subscription']['id'] == ids(SUB_TYPE)
  assert data['subscription']['notification_index'] == 0
  assert data['subscription']['uss_base_url'] == BASE_URL
  assert_datetimes_are_equal(data['subscription']['time_start']['value'], req_body['extents']['time_start']['value'])
  assert_datetimes_are_equal(data['subscription']['time_end']['value'], req_body['extents']['time_end']['value'])
  assert re.match(r'[a-z0-9]{10,}$', data['subscription']['version'])
  assert 'service_areas' in data


@default_scope(SCOPE_DP)
def test_get_sub_by_id(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_E_GET_SUB_BY_ID."""
  resp = session_ridv2.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['subscription']['id'] == ids(SUB_TYPE)
  assert data['subscription']['notification_index'] == 0
  assert data['subscription']['uss_base_url'] == BASE_URL


@default_scope(SCOPE_DP)
def test_get_sub_by_search(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_F_GET_SUBS_BY_AREA."""
  resp = session_ridv2.get('{}?area={}'.format(SUBSCRIPTION_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(SUB_TYPE) in [x['id'] for x in resp.json()['subscriptions']]


@default_scope(SCOPE_DP)
def test_get_sub_by_searching_huge_area(session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(SUBSCRIPTION_PATH, common.HUGE_GEO_POLYGON_STRING))
  assert resp.status_code == 413, resp.content


@default_scope(SCOPE_DP)
def test_delete_sub_empty_version(ids, session_ridv2):
  resp = session_ridv2.delete('{}/{}/'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_DP)
def test_delete_sub_wrong_version(ids, session_ridv2):
  resp = session_ridv2.delete('{}/{}/fake_version'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_DP)
def test_delete_sub(ids, session_ridv2):
  """ASTM Compliance Test: DSS0030_D_DELETE_SUB."""
  # GET the sub first to find its version.
  resp = session_ridv2.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 200, resp.content
  version = resp.json()['subscription']['version']

  # Then delete it.
  resp = session_ridv2.delete('{}/{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE), version))
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_DP)
def test_get_deleted_sub_by_id(ids, session_ridv2):
  resp = session_ridv2.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_DP)
def test_get_deleted_sub_by_search(ids, session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(SUBSCRIPTION_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(SUB_TYPE) not in [x['id'] for x in resp.json()['subscriptions']]


@default_scope(SCOPE_DP)
def test_get_sub_with_loop_area(session_ridv2):
  resp = session_ridv2.get('{}?area={}'.format(SUBSCRIPTION_PATH, common.LOOP_GEO_POLYGON_STRING))
  assert resp.status_code == 400, resp.content
