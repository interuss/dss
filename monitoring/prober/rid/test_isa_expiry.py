"""Test ISAs aren't returned after they expire."""

import datetime
import time

from ..infrastructure import default_scope
from . import common
from .common import SCOPE_READ, SCOPE_WRITE

ISA_ID = '00000098-ba6d-4c20-a575-6e412e000000'


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
def test_create(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=5)

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
      })
  assert resp.status_code == 200


@default_scope(SCOPE_READ)
def test_valid_immediately(session):
  # The ISA is still valid immediately after we create it.
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 200


def test_sleep_5_seconds():
  # But if we wait 5 seconds it will expire...
  time.sleep(5)


@default_scope(SCOPE_READ)
def test_returned_by_id(session):
  # We can get it explicitly by ID
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID))
  assert resp.status_code == 200


@default_scope(SCOPE_READ)
def test_not_returned_by_search(session):
  # ...but it's not included in a search.
  resp = session.get('/identification_service_areas?area={}'.format(
      common.GEO_POLYGON_STRING))
  assert resp.status_code == 200
  assert ISA_ID not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_delete(session):
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']
  resp = session.delete('/identification_service_areas/{}/{}'.format(ISA_ID, version), scope=SCOPE_WRITE)
  assert resp.status_code == 200, resp.content
