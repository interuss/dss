"""Test subscriptions interact with ISAs:

  - Create an ISA.
  - Create a subscription, response should include the pre-existing ISA.
  - Modify the ISA, response should include the subscription.
  - Delete the ISA, response should include the subscription.
  - Delete the subscription.
"""

import datetime

from . import common


def test_create_isa(session, isa1_uuid):
  time_start = datetime.datetime.utcnow()
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
          'flights_url': 'https://example.com/dss',
      })
  assert resp.status_code == 200


def test_create_subscription(session, isa1_uuid, sub1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/subscriptions/{}'.format(sub1_uuid),
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
          'callbacks': {
              'identification_service_area_url': 'https://example.com/foo'
          },
      })
  assert resp.status_code == 200

  # The response should include our ISA.
  data = resp.json()
  assert data['subscription']['notification_index'] == 0
  assert isa1_uuid in [x['id'] for x in data['service_areas']]


def test_modify_isa(session, isa1_uuid, sub1_uuid):
  # GET the ISA first to find its version.
  resp = session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']

  # Then modify it.
  time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=60)
  resp = session.put(
      '/identification_service_areas/{}/{}'.format(isa1_uuid, version),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                      'vertices': common.VERTICES,
                  },
                  'altitude_lo': 12345,
                  'altitude_hi': 67890,
              },
  
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
      })
  assert resp.status_code == 200

  # The response should include our subscription.
  data = resp.json()
  assert {
      'url':
          'https://example.com/foo',
      'subscriptions': [{
          'notification_index': 1,
          'subscription_id': sub1_uuid,
      },],
  } in data['subscribers']


def test_delete_isa(session, isa1_uuid, sub1_uuid):
  # GET the ISA first to find its version.
  resp = session.get('/identification_service_areas/{}'.format(isa1_uuid))
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session.delete('/identification_service_areas/{}/{}'.format(
      isa1_uuid, version))
  assert resp.status_code == 200

  # The response should include our subscription.
  data = resp.json()
  assert {
      'url':
          'https://example.com/foo',
      'subscriptions': [{
          'notification_index': 2,
          'subscription_id': sub1_uuid,
      },],
  } in data['subscribers']


def test_delete_subscription(session, sub1_uuid):
  # GET the sub first to find its version.
  resp = session.get('/subscriptions/{}'.format(sub1_uuid))
  assert resp.status_code == 200

  data = resp.json()
  version = data['subscription']['version']
  assert data['subscription']['notification_index'] == 2

  # Then delete it.
  resp = session.delete('/subscriptions/{}/{}'.format(sub1_uuid, version))
  assert resp.status_code == 200
