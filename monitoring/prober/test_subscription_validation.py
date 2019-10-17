
"""Subscription input validation tests:
  - check we can't create too many SUBS (common.MAX_SUBS_PER_AREA)
  - check we can't create the SUB with a huge area
  - check we can't create the SUB with missing fields
  - check we can't create the SUB with a time_start in the past
  - check we can't create the SUB with a time_start after time_end
"""
import datetime
import common
import uuid


def test_create_sub_empty_vertices(session, sub2_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(sub2_uuid),
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


def test_create_sub_missing_footprint(session, sub2_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(sub2_uuid),
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


def test_create_sub_with_huge_area(session, sub2_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=10)

  resp = session.put(
      '/subscriptions/{}'.format(sub2_uuid),
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


def test_create_too_many_subs(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=1)
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
    all_resp.append(
        resp.status_code == (429 if index == common.MAX_SUB_PER_AREA else 200)
    )

  assert all(all_resp)


def test_create_sub_with_too_long_end_time(session, sub2_uuid):
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(hours=(common.MAX_SUB_TIME_HRS + 1))

    resp = session.put(
        "/subscriptions/{}".format(sub2_uuid),
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