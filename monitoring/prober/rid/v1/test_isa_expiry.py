"""Test ISAs aren't returned after they expire."""

import datetime
import time

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE, ISA_PATH
from monitoring.prober.infrastructure import register_resource_type
from . import common


ISA_TYPE = register_resource_type(222, 'ISA')


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


@default_scope(SCOPE_WRITE)
def test_create(ids, session_ridv1):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(seconds=5)

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
def test_valid_immediately(ids, session_ridv1):
  # The ISA is still valid immediately after we create it.
  resp = session_ridv1.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 200, resp.content


def test_sleep_5_seconds():
  # But if we wait 5 seconds it will expire...
  time.sleep(5)


@default_scope(SCOPE_READ)
def test_returned_by_id(ids, session_ridv1):
  # We can get it explicitly by ID
  resp = session_ridv1.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)))
  assert resp.status_code == 200, resp.content


@default_scope(SCOPE_READ)
def test_not_returned_by_search(ids, session_ridv1):
  # ...but it's not included in a search.
  resp = session_ridv1.get('{}?area={}'.format(
      ISA_PATH, common.GEO_POLYGON_STRING))
  assert resp.status_code == 200, resp.content
  assert ids(ISA_TYPE) not in [x['id'] for x in resp.json()['service_areas']]


@default_scope(SCOPE_READ)
def test_delete(ids, session_ridv1):
  resp = session_ridv1.get('{}/{}'.format(ISA_PATH, ids(ISA_TYPE)), scope=SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['service_area']['version']
  resp = session_ridv1.delete('{}/{}/{}'.format(ISA_PATH, ids(ISA_TYPE), version), scope=SCOPE_WRITE)
  assert resp.status_code == 200, resp.content
