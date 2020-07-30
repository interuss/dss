"""Test subscriptions interact with ISAs:

  - Create an ISA.
  - Create a subscription, response should include the pre-existing ISA.
  - Modify the ISA, response should include the subscription.
  - Delete the ISA, response should include the subscription.
  - Delete the subscription.
"""

import datetime

from ..infrastructure import default_scope
from . import common
from .common import SCOPE_READ, SCOPE_WRITE

ISA_ID = '000000d5-aa3d-46b8-b2ec-dd22e7000000'
SUB_ID = '000000ee-85c7-4bc6-8995-aa5f81000000'


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

  resp = session.get('/subscriptions/{}'.format(SUB_ID), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session.delete('/subscriptions/{}/{}'.format(SUB_ID, version), scope=SCOPE_READ)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_WRITE)
def test_create_isa(session):
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
              'time_start': time_start.strftime(common.DATE_FORMAT),
              'time_end': time_end.strftime(common.DATE_FORMAT),
          },
          'flights_url': 'https://example.com/dss',
      })
  assert resp.status_code == 200


@default_scope(SCOPE_READ)
def test_create_subscription(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session.put(
      '/subscriptions/{}'.format(SUB_ID),
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
  assert ISA_ID in [x['id'] for x in data['service_areas']]


def test_modify_isa(session):
  # GET the ISA first to find its version.
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']

  # Then modify it.
  time_end = datetime.datetime.utcnow() + datetime.timedelta(minutes=60)
  resp = session.put(
      '/identification_service_areas/{}/{}'.format(ISA_ID, version),
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
      }, scope=SCOPE_WRITE)
  assert resp.status_code == 200

  # The response should include our subscription.
  data = resp.json()
  assert {
      'url':
          'https://example.com/foo',
      'subscriptions': [{
          'notification_index': 1,
          'subscription_id': SUB_ID,
      },],
  } in data['subscribers']


def test_delete_isa(session):
  # GET the ISA first to find its version.
  resp = session.get('/identification_service_areas/{}'.format(ISA_ID), scope=SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session.delete('/identification_service_areas/{}/{}'.format(
      ISA_ID, version), scope=SCOPE_WRITE)
  assert resp.status_code == 200

  # The response should include our subscription.
  data = resp.json()
  assert {
      'url':
          'https://example.com/foo',
      'subscriptions': [{
          'notification_index': 2,
          'subscription_id': SUB_ID,
      },],
  } in data['subscribers']


@default_scope(SCOPE_READ)
def test_delete_subscription(session):
  # GET the sub first to find its version.
  resp = session.get('/subscriptions/{}'.format(SUB_ID))
  assert resp.status_code == 200

  data = resp.json()
  version = data['subscription']['version']
  assert data['subscription']['notification_index'] == 2

  # Then delete it.
  resp = session.delete('/subscriptions/{}/{}'.format(SUB_ID, version))
  assert resp.status_code == 200
