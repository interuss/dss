"""ISA input validation tests:

  - check we can't create the ISA with a huge area
  - check we can't create the ISA with missing fields
  - check we can't create the ISA with a time_start in the past
  - check we can't create the ISA with a time_start after time_end
"""

import datetime

from ..infrastructure import default_scope
from . import common
from .common import SCOPE_WRITE


@default_scope(SCOPE_WRITE)
def test_isa_huge_area(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.HUGE_VERTICES,
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert 'area is too large' in resp.json()['message']


@default_scope(SCOPE_WRITE)
def test_isa_empty_vertices(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': [],
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json()['message'] == 'bad extents: not enough points in polygon'


@default_scope(SCOPE_WRITE)
def test_isa_missing_footprint(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'extents': {
              'spatial_volume': {
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json(
  )['message'] == 'bad extents: spatial_volume missing required footprint'


@default_scope(SCOPE_WRITE)
def test_isa_missing_spatial_volume(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'extents': {
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json(
  )['message'] == 'bad extents: missing required spatial_volume'


@default_scope(SCOPE_WRITE)
def test_isa_missing_extents(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json()['message'] == 'missing required extents'


@default_scope(SCOPE_WRITE)
def test_isa_start_time_in_past(session, isa1_uuid):
  time_start = datetime.datetime.utcnow() - datetime.timedelta(minutes=10)
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
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json(
  )['message'] == 'IdentificationServiceArea time_start must not be in the past'


@default_scope(SCOPE_WRITE)
def test_isa_start_time_after_time_end(session, isa1_uuid):
  time_start = datetime.datetime.utcnow() + datetime.timedelta(minutes=10)
  time_end = time_start - datetime.timedelta(minutes=5)

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
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
  assert resp.json(
  )['message'] == 'IdentificationServiceArea time_end must be after time_start'


@default_scope(SCOPE_WRITE)
def test_isa_not_on_earth(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/identification_service_areas/{}'.format(isa1_uuid),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': [
                        {'lat': 130.6205, 'lng': -23.6558},
                        {'lat': 130.6301, 'lng': -23.6898},
                        {'lat': 130.6700, 'lng': -23.6709},
                        {'lat': 130.6466, 'lng': -23.6407},
                      ],
                  },
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/uss/flights',
      })
  assert resp.status_code == 400
