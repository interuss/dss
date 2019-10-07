"""Basic ISA tests:

  - create the ISA with a 60 minute expiry
  - get by ID
  - search with earliest_time and latest_time
  - delete
"""

import datetime
import re

import common


def test_isa_does_not_exist(session, isa1_uuid):
  resp = session.get('/v1/dss/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 404
  assert resp.json()['message'] == 'resource not found: {}'.format(isa1_uuid)


def test_create_isa(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/v1/dss/identification_service_areas/{}'.format(isa1_uuid),
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

  data = resp.json()
  assert data['service_area']['id'] == isa1_uuid
  assert data['service_area']['flights_url'] == 'https://example.com/dss'
  assert data['service_area']['time_start'] == time_start.strftime(
      common.DATE_FORMAT)
  assert data['service_area']['time_end'] == time_end.strftime(
      common.DATE_FORMAT)
  assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
  assert 'subscribers' in data


def test_get_isa_by_id(session, isa1_uuid):
  resp = session.get('/v1/dss/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 200

  data = resp.json()
  assert data['service_area']['id'] == isa1_uuid
  assert data['service_area']['flights_url'] == 'https://example.com/dss'


def test_get_isa_by_search(session, isa1_uuid):
  resp = session.get('/v1/dss/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200
  assert isa1_uuid in [x['id'] for x in resp.json()['service_areas']]


def test_get_isa_by_search_earliest_time_included(session, isa1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = session.get('/v1/dss/identification_service_areas'
                     '?area={}&earliest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         earliest_time.strftime(common.DATE_FORMAT)))
  assert resp.status_code == 200
  assert isa1_uuid in [x['id'] for x in resp.json()['service_areas']]


def test_get_isa_by_search_earliest_time_excluded(session, isa1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = session.get('/v1/dss/identification_service_areas'
                     '?area={}&earliest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         earliest_time.strftime(common.DATE_FORMAT)))
  assert resp.status_code == 200
  assert isa1_uuid not in [x['id'] for x in resp.json()['service_areas']]


def test_get_isa_by_search_latest_time_included(session, isa1_uuid):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = session.get('/v1/dss/identification_service_areas'
                     '?area={}&latest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         latest_time.strftime(common.DATE_FORMAT)))
  assert resp.status_code == 200
  assert isa1_uuid in [x['id'] for x in resp.json()['service_areas']]


def test_get_isa_by_search_latest_time_excluded(session, isa1_uuid):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = session.get('/v1/dss/identification_service_areas'
                     '?area={}&latest_time={}'.format(
                         common.GEO_POLYGON_STRING,
                         latest_time.strftime(common.DATE_FORMAT)))
  assert resp.status_code == 200
  assert isa1_uuid not in [x['id'] for x in resp.json()['service_areas']]


def test_delete_isa(session, isa1_uuid):
  # GET the ISA first to find its version.
  resp = session.get('/v1/dss/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session.delete('/v1/dss/identification_service_areas/{}/{}'.format(
      isa1_uuid, version))
  assert resp.status_code == 200


def test_get_deleted_isa_by_id(session, isa1_uuid):
  resp = session.get('/v1/dss/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 404


def test_get_deleted_isa_by_search(session, isa1_uuid):
  resp = session.get('/v1/dss/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200
  assert isa1_uuid not in [x['id'] for x in resp.json()['service_areas']]
