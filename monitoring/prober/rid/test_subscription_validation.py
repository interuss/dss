
"""Subscription input validation tests:
  - check we can't create too many SUBS (common.MAX_SUBS_PER_AREA)
  - check we can't create the SUB with a huge area
  - check we can't create the SUB with missing fields
  - check we can't create the SUB with a time_start in the past
  - check we can't create the SUB with a time_start after time_end
"""
import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.rid import SCOPE_READ
from monitoring.prober.infrastructure import register_resource_type
from . import common


SUB_TYPE = register_resource_type(328, 'Subscription')
MULTI_SUB_TYPES = [register_resource_type(329 + i, 'Subscription limit Subscription {}'.format(i)) for i in range(11)]


def test_ensure_clean_workspace(ids, session):
  resp = session.get('/subscriptions/{}'.format(ids(SUB_TYPE)), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session.delete('/subscriptions/{}/{}'.format(ids(SUB_TYPE), version), scope=SCOPE_READ)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_READ)
def test_create_sub_empty_vertices(ids, session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(ids(SUB_TYPE)),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': [],
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
      })
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_create_sub_missing_footprint(ids, session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(ids(SUB_TYPE)),
      json={
          'extents': {
              'spatial_volume': {
                  'altitude_lo': 20,
                  'altitude_hi': 400,
              },
              'time_start': time_start.strftime(rid.DATE_FORMAT),
              'time_end': time_end.strftime(rid.DATE_FORMAT),
          },
          'callbacks': {
              'identification_service_area_url': 'https://example.com/foo'
          },
      })
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_create_sub_with_huge_area(ids, session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(ids(SUB_TYPE)),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.HUGE_VERTICES,
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
      })
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_create_too_many_subs(ids, session):
  """ASTM Compliance Test: DSS0050_MAX_SUBS_PER_AREA."""
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=30)

  # create 1 more than the max allowed Subscriptions per area
  versions = []
  for index in range(rid.MAX_SUB_PER_AREA + 1):
    resp = session.put(
        '/subscriptions/{}'.format(ids(MULTI_SUB_TYPES[index])),
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
                'time_start': time_start.strftime(rid.DATE_FORMAT),
                'time_end': time_end.strftime(rid.DATE_FORMAT),
            },
            'callbacks': {
                'identification_service_area_url': 'https://example.com/foo'
            },
        })
    if index < rid.MAX_SUB_PER_AREA:
      assert resp.status_code == 200, resp.content
      resp_json = resp.json()
      assert 'subscription' in resp_json
      assert 'version' in resp_json['subscription']
      versions.append(resp_json['subscription']['version'])
    else:
      assert resp.status_code == 429, resp.content

  # clean up Subscription-limit Subscriptions
  for index in range(rid.MAX_SUB_PER_AREA):
    resp = session.delete('/subscriptions/{}/{}'.format(ids(MULTI_SUB_TYPES[index]), versions[index]))
    assert resp.status_code == 200


@default_scope(SCOPE_READ)
def test_create_sub_with_too_long_end_time(ids, session):
    """ASTM Compliance Test: DSS0060_MAX_SUBS_DURATION."""
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(hours=(rid.MAX_SUB_TIME_HRS + 1))

    resp = session.put(
        "/subscriptions/{}".format(ids(SUB_TYPE)),
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(rid.DATE_FORMAT),
                "time_end": time_end.strftime(rid.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_update_sub_with_too_long_end_time(ids, session):
    """ASTM Compliance Test: DSS0060_MAX_SUBS_DURATION."""
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(seconds=10)

    resp = session.put(
        '/subscriptions/{}'.format(ids(SUB_TYPE)),
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(rid.DATE_FORMAT),
                "time_end": time_end.strftime(rid.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 200, resp.content

    time_end = time_start + datetime.timedelta(hours=(rid.MAX_SUB_TIME_HRS + 1))
    resp = session.put(
        '/subscriptions/{}'.format(ids(SUB_TYPE)) + '/' + resp.json()["subscription"]["version"],
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(rid.DATE_FORMAT),
                "time_end": time_end.strftime(rid.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 400, resp.content


@default_scope(SCOPE_READ)
def test_delete(ids, session):
  resp = session.get('/subscriptions/{}'.format(ids(SUB_TYPE)), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session.delete('/subscriptions/{}/{}'.format(ids(SUB_TYPE), version), scope=SCOPE_READ)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
