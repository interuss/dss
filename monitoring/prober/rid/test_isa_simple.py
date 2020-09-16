"""Basic ISA tests:

  - create the ISA with a 60 minute expiry
  - get by ID
  - search with earliest_time and latest_time
  - delete
"""

import datetime
import re

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE
from . import common


ISA_ID = '00000007-cd0d-420b-b259-293b3c000000'


def test_ensure_clean_workspace(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['service_area']['version']
    resp = session.delete('/identification_service_areas/{}/{}'.format(ISA_ID, version), scope=SCOPE_WRITE)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_WRITE)
def test_create_isa(session):
  """ASTM Compliance Test: DSS0030_A_PUT_ISA."""
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
              'time_start': time_start.strftime(rid.DATE_FORMAT),
              'time_end': time_end.strftime(rid.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
      })
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['service_area']['id'] == ISA_ID
  assert data['service_area']['flights_url'] == 'https://example.com/dss'
  assert data['service_area']['time_start'] == time_start.strftime(
      rid.DATE_FORMAT)
  assert data['service_area']['time_end'] == time_end.strftime(
      rid.DATE_FORMAT)
  assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
  assert 'subscribers' in data


@default_scope(SCOPE_READ)
def test_get_isa_by_id(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  assert data['service_area']['id'] == ISA_ID
  assert data['service_area']['flights_url'] == 'https://example.com/dss'


@default_scope(SCOPE_WRITE)
def test_update_isa(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  version = resp.json()['service_area']['version']

  resp = session.put(
    '/identification_service_areas/{}/{}'.format(ISA_ID, version),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
          },
          'flights_url': 'https://example.com/dss/v2',
      })
  assert resp.status_code == 200

  data = resp.json()
  assert data['service_area']['id'] == ISA_ID
  assert data['service_area']['flights_url'] == 'https://example.com/dss/v2'
  assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
  assert 'subscribers' in data


@default_scope(SCOPE_READ)
def test_get_isa_by_search_missing_params(session):
  resp = session.get('/identification_service_areas')
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_get_isa_by_search(session):
  resp = session.get('/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ISA_ID in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_earliest_time_included(session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = session.get('/identification_service_areas'
                     '?area={}&earliest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         earliest_time.strftime(rid.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ISA_ID in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_earliest_time_excluded(session):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = session.get('/identification_service_areas'
                     '?area={}&earliest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         earliest_time.strftime(rid.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ISA_ID not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_latest_time_included(session):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = session.get('/identification_service_areas'
                     '?area={}&latest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         latest_time.strftime(rid.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ISA_ID in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_latest_time_excluded(session):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = session.get('/identification_service_areas'
                     '?area={}&latest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         latest_time.strftime(rid.DATE_FORMAT)))
  assert resp.status_code == 200, resp.content
  assert ISA_ID not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_area_only(session):
  resp = session.get('/identification_service_areas'
                     '?area={}'.format(common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ISA_ID in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_get_isa_by_search_huge_area(session):
  resp = session.get('/identification_service_areas'
                     '?area={}'.format(common.HUGE_GEO_POLYGON_STRING))
  assert resp.status_code == 413, resp.content


@default_scope(SCOPE_WRITE)
def test_delete_isa_wrong_version(session):
  resp = session.delete('/identification_service_areas/{}/fake_version'.format(ISA_ID))
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_WRITE)
def test_delete_isa_empty_version(session):
  resp = session.delete('/identification_service_areas/{}/'.format(ISA_ID))
  assert resp.status_code == 400, resp.content


def test_delete_isa(session):
  """ASTM Compliance Test: DSS0030_B_DELETE_ISA."""
  # GET the ISA first to find its version.
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  assert resp.status_code == 200, resp.content
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session.delete('/identification_service_areas/{}/{}'.format(ISA_ID, version), scope=SCOPE_WRITE)
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_READ)
def test_get_deleted_isa_by_id(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 404, resp.content


@default_scope(SCOPE_READ)
def test_get_deleted_isa_by_search(session):
  resp = session.get('/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ISA_ID not in [x['id'] for x in resp.json()['service_areas']]
