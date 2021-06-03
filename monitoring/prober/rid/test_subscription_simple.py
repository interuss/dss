"""Basic subscription tests:

  - create the subscription with a 60 minute expiry
  - get by ID
  - get by search
  - delete
"""

import datetime
import re

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.rid import SCOPE_READ
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from . import common

SUB_ID = '000000e0-cf69-456a-91fb-fc9532000000'


def test_ensure_clean_workspace(session):
  resp = session.get('/subscriptions/{}'.format(SUB_ID), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session.delete('/subscriptions/{}/{}'.format(SUB_ID, version), scope=SCOPE_READ)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_READ)
def test_sub_does_not_exist(session):
  resp = session.get('/subscriptions/{}'.format(SUB_ID))
  assert resp.status_code == 404, resp.content
  assert 'Subscription {} not found'.format(SUB_ID) in resp.json()['message']


@default_scope(SCOPE_READ)
def test_create_sub(session):
  """ASTM Compliance Test: DSS0030_C_PUT_SUB."""
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  req_body = {
    'extents': {
      'spatial_volume': {
        'footprint': {
          'vertices': common.VERTICES,
        },
        'altitude_lo': 20,
        'altitude_hi': 400,
      },
      'time_start': time_start.strftime(rid.DATE_FORMAT),
      'time_end': time_end.strftime(rid.DATE_FORMAT),
    },
    'callbacks': {
      'identification_service_area_url': 'https://example.com/foo'
    },
  }
  resp = session.put(
      '/subscriptions/{}'.format(SUB_ID),
      json=req_body)
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['subscription']['id'] == SUB_ID
  assert data['subscription']['notification_index'] == 0
  assert data['subscription']['callbacks'] == {
      'identification_service_area_url': 'https://example.com/foo'
  }
  assert_datetimes_are_equal(data['subscription']['time_start'], req_body['extents']['time_start'])
  assert_datetimes_are_equal(data['subscription']['time_end'], req_body['extents']['time_end'])
  assert re.match(r'[a-z0-9]{10,}$', data['subscription']['version'])
  assert 'service_areas' in data


@default_scope(SCOPE_READ)
def test_get_sub_by_id(session):
  """ASTM Compliance Test: DSS0030_E_GET_SUB_BY_ID."""
  resp = session.get('/subscriptions/{}'.format(SUB_ID))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['subscription']['id'] == SUB_ID
  assert data['subscription']['notification_index'] == 0
  assert data['subscription']['callbacks'] == {
      'identification_service_area_url': 'https://example.com/foo'
  }


@default_scope(SCOPE_READ)
def test_get_sub_by_search(session):
  """ASTM Compliance Test: DSS0030_F_GET_SUBS_BY_AREA."""
  resp = session.get('/subscriptions?area={}'.format(common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert SUB_ID in [x['id'] for x in resp.json()['subscriptions']]


@default_scope(SCOPE_READ)
def test_get_sub_by_searching_huge_area(session):
  resp = session.get('/subscriptions?area={}'.format(common.HUGE_GEO_POLYGON_STRING))
  assert resp.status_code == 413, resp.content


@default_scope(SCOPE_READ)
def test_delete_sub_empty_version(session):
  resp = session.delete('/subscriptions/{}/'.format(SUB_ID))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_delete_sub_wrong_version(session):
  resp = session.delete('/subscriptions/{}/fake_version'.format(SUB_ID))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_delete_sub(session):
  """ASTM Compliance Test: DSS0030_D_DELETE_SUB."""
  # GET the sub first to find its version.
  resp = session.get('/subscriptions/{}'.format(SUB_ID))
  assert resp.status_code == 200, resp.content
  version = resp.json()['subscription']['version']

  # Then delete it.
  resp = session.delete('/subscriptions/{}/{}'.format(SUB_ID, version))
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_READ)
def test_get_deleted_sub_by_id(session):
  resp = session.get('/subscriptions/{}'.format(SUB_ID))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_READ)
def test_get_deleted_sub_by_search(session):
  resp = session.get('/subscriptions?area={}'.format(common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert SUB_ID not in [x['id'] for x in resp.json()['subscriptions']]
