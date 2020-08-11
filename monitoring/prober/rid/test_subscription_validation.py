
"""Subscription input validation tests:
  - check we can't create too many SUBS (common.MAX_SUBS_PER_AREA)
  - check we can't create the SUB with a huge area
  - check we can't create the SUB with missing fields
  - check we can't create the SUB with a time_start in the past
  - check we can't create the SUB with a time_start after time_end
"""
import datetime
import uuid

from ..infrastructure import default_scope
from . import common
from .common import SCOPE_READ

SUB_ID = '000000e8-46a4-4df1-b924-f455ad000000'


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
def test_create_sub_empty_vertices(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(SUB_ID),
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
          'callbacks': {
              'identification_service_area_url': 'https://example.com/foo'
          },
      })
  assert resp.status_code == 400


@default_scope(SCOPE_READ)
def test_create_sub_missing_footprint(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(SUB_ID),
      json={
          'extents': {
              'spatial_volume': {
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'callbacks': {
              'identification_service_area_url': 'https://example.com/foo'
          },
      })
  assert resp.status_code == 400


@default_scope(SCOPE_READ)
def test_create_sub_with_huge_area(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(SUB_ID),
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
          'callbacks': {
              'identification_service_area_url': 'https://example.com/foo'
          },
      })
  assert resp.status_code == 400


@default_scope(SCOPE_READ)
def test_create_too_many_subs(session):
  """ASTM Compliance Test: DSS0050_MAX_SUBS_PER_AREA."""
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=30)
  all_resp = []

  # create 1 more than the max allowed Subscriptions per area
  for index in range(common.MAX_SUB_PER_AREA + 1):
    resp = session.put(
        '/subscriptions/{}'.format(str(uuid.uuid4())),
        json={
            'extents': {
                'spatial_volume': {
                    'footprint': {
                        'vertices': [
                            {
                                "lat": 37.440,
                                "lng": -131.745,
                            },
                            {
                                "lat": 37.459,
                                "lng": -131.745,
                            },
                            {
                                "lat": 37.459,
                                "lng": -131.706,
                            },
                            {
                                "lat": 37.440,
                                "lng": -131.706,
                            },
                        ],
                    },
                    'altitude_lo': 20,
                    'altitude_hi': 400,
                },
                'time_start': time_start.strftime(common.DATE_FORMAT),
                'time_end': time_end.strftime(common.DATE_FORMAT),
            },
            'callbacks': {
                'identification_service_area_url': 'https://example.com/foo'
            },
        })
    if index < common.MAX_SUB_PER_AREA:
      assert resp.status_code == 200, resp.content
    else:
      assert resp.status_code == 429, resp.content


@default_scope(SCOPE_READ)
def test_create_sub_with_too_long_end_time(session):
    """ASTM Compliance Test: DSS0060_MAX_SUBS_DURATION."""
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(hours=(common.MAX_SUB_TIME_HRS + 1))

    resp = session.put(
        "/subscriptions/{}".format(SUB_ID),
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(common.DATE_FORMAT),
                "time_end": time_end.strftime(common.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 400


@default_scope(SCOPE_READ)
def test_update_sub_with_too_long_end_time(session):
    """ASTM Compliance Test: DSS0060_MAX_SUBS_DURATION."""
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(seconds=10)

    resp = session.put(
        '/subscriptions/{}'.format(SUB_ID),
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(common.DATE_FORMAT),
                "time_end": time_end.strftime(common.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 200

    time_end = time_start + datetime.timedelta(hours=(common.MAX_SUB_TIME_HRS + 1))
    resp = session.put(
        '/subscriptions/{}'.format(SUB_ID) + '/' + resp.json()["subscription"]["version"],
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(common.DATE_FORMAT),
                "time_end": time_end.strftime(common.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 400
