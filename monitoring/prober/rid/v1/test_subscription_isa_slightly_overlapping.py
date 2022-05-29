"""Test subscriptions interact with ISAs:

  - Create an ISA.
  - Create a subscription slightly overlapping with ISA, response should include the pre-existing ISA.
  - Delete the ISA, response should include the subscription.
  - Delete the subscription.
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE, ISA_PATH, SUBSCRIPTION_PATH
from monitoring.prober.infrastructure import register_resource_type
from . import common


ISA_TYPE = register_resource_type(427, 'ISA')
SUB_TYPE = register_resource_type(428, 'Subscription')


def test_ensure_clean_workspace(ids, session_ridv1):
  resp = session_ridv1.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['service_area']['version']
    resp = session_ridv1.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_WRITE)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected.
    pass
  else:
    assert False, resp.content

  resp = session_ridv1.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)), scope=SCOPE_READ)
  if resp.status_code == 200:
    version = resp.json()['subscription']['version']
    resp = session_ridv1.delete('{}/{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE), version), scope=SCOPE_READ)
    assert resp.status_code == 200, resp.content
  elif resp.status_code == 404:
    # As expected
    pass
  else:
    assert False, resp.content


@default_scope(SCOPE_WRITE)
def test_create_isa(ids, session_ridv1):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv1.put(
      '{}/{}'.format(ISA_PATH, ids(ISA_TYPE)),
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


@default_scope(SCOPE_READ)
def test_create_subscription(ids, session_ridv1):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = session_ridv1.put(
      '{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)),
      json={
          'extents': {
              'spatial_volume': {
                  'footprint': {
                    #   create subscription closer to last created ISA.
                      'vertices': common.CLOSER_POLYGON_VERTICES,
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
  assert resp.status_code == 200, resp.content

  # The response should include our ISA.
  data = resp.json()
  assert data['subscription']['notification_index'] == 0
  assert ids(ISA_TYPE) in [x['id'] for x in data['service_areas']]


def test_delete_isa(ids, session_ridv1):
  # GET the ISA first to find its version.
  resp = session_ridv1.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_READ)
  assert resp.status_code == 200, resp.content
  version = resp.json()['service_area']['version']

  # Then delete it.
  resp = session_ridv1.delete('{}/{}/{}'.format(
      ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_WRITE)
  assert resp.status_code == 200, resp.content

  # The response should include our subscription.
  data = resp.json()
  assert {
      'url':
          'https://example.com/foo',
      'subscriptions': [{
          'notification_index': 1,
          'subscription_id': ids(SUB_TYPE),
      },],
  } in data['subscribers']


@default_scope(SCOPE_READ)
def test_delete_subscription(ids, session_ridv1):
  # GET the sub first to find its version.
  resp = session_ridv1.get('{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE)))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  version = data['subscription']['version']
  assert data['subscription']['notification_index'] == 1

  # Then delete it.
  resp = session_ridv1.delete('{}/{}/{}'.format(SUBSCRIPTION_PATH, ids(SUB_TYPE), version))
  assert resp.status_code == 200, resp.content
