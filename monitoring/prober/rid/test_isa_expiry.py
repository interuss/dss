"""Test ISAs aren't returned after they expire."""

import datetime
import time

from . import common


def test_create(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=5)

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


def test_valid_immediately(session, isa1_uuid):
  # The ISA is still valid immediately after we create it.
  resp = session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 200


def test_sleep_5_seconds():
  # But if we wait 5 seconds it will expire...
  time.sleep(5)


def test_not_returned_by_id(session, isa1_uuid):
  # And we can't get it by ID...
  resp = session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 404
  assert resp.json()['message'] == 'resource not found: {}'.format(isa1_uuid)


def test_not_returned_by_search(session, isa1_uuid):
  # Or by search.
  resp = session.get('/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200
  assert isa1_uuid not in [x['id'] for x in resp.json()['service_areas']]
