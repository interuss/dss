"""Basic ISA tests with hundreds of ISAs in the SAME area created CONCURRENTLY:

  - create ISAs with a 60 minute expiry concurrently
  - get by IDs concurrently
  - search with area
  - delete concurrently
"""

import datetime
import functools
import json
import re
from concurrent.futures.thread import ThreadPoolExecutor

from monitoring.monitorlib import rid
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from . import common


def _load_isa_ids():
  with open('./rid/resources/isa_ids_heavy_traffic_concurrent.json', 'r') as f:
    return json.load(f)


THREAD_COUNT = 10
FLIGHTS_URL = 'https://example.com/dss'
ISA_IDS = _load_isa_ids()


def _intersection(list1, list2):
  return list(set(list1) & set(list2))


def _make_isa_request(time_start, time_end):
  return {
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
    'flights_url': FLIGHTS_URL,
  }


def _put_isa(isa_id, req, session):
  return session.put('/identification_service_areas/{}'.format(isa_id), json=req, scope=SCOPE_WRITE)


def _get_isa(isa_id, session):
  return session.get('/identification_service_areas/{}'.format(isa_id), scope=SCOPE_READ)


def _delete_isa(isa_id, version, session):
  return session.delete('/identification_service_areas/{}/{}'.format(isa_id, version), scope=SCOPE_WRITE)


def _collect_resp_callback(key, resp_map, future):
  resp_map[key] = future.result()


def test_ensure_clean_workspace(session):
  for isa_id in ISA_IDS:
    resp = session.get('/identification_service_areas/{}'.format(isa_id), scope=SCOPE_READ)
    if resp.status_code == 200:
      version = resp.json()['service_area']['version']
      resp = session.delete('/identification_service_areas/{}/{}'.format(isa_id, version), scope=SCOPE_WRITE)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


@default_scope(SCOPE_WRITE)
def test_create_isa_concurrent(session):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  req = _make_isa_request(time_start, time_end)
  resp_map = {}

  # Create ISAs concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    for isa_id in ISA_IDS:
      future = executor.submit(_put_isa, isa_id, req, session)
      future.add_done_callback(functools.partial(_collect_resp_callback, isa_id, resp_map))

  for isa_id, resp in resp_map.items():
    assert resp.status_code == 200, resp.content
    data = resp.json()
    assert data['service_area']['id'] == isa_id
    assert data['service_area']['flights_url'] == 'https://example.com/dss'
    assert_datetimes_are_equal(data['service_area']['time_start'], req['extents']['time_start'])
    assert_datetimes_are_equal(data['service_area']['time_end'], req['extents']['time_end'])
    assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
    assert 'subscribers' in data


@default_scope(SCOPE_READ)
def test_get_isa_by_ids_concurrent(session):
  resp_map = {}

  # Get ISAs concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    for isa_id in ISA_IDS:
      future = executor.submit(_get_isa, isa_id, session)
      future.add_done_callback(functools.partial(_collect_resp_callback, isa_id, resp_map))

  for isa_id, resp in resp_map.items():
    assert resp.status_code == 200, resp.content

    data = resp.json()
    assert data['service_area']['id'] == isa_id
    assert data['service_area']['flights_url'] == FLIGHTS_URL


@default_scope(SCOPE_READ)
def test_get_isa_by_search(session):
  resp = session.get('/identification_service_areas?area={}'.format(common.GEO_POLYGON_STRING))

  assert resp.status_code == 200, resp.content
  found_isa_ids = [x['id'] for x in resp.json()['service_areas']]
  assert len(_intersection(ISA_IDS, found_isa_ids)) == len(ISA_IDS)


def test_delete_isa_concurrent(session):
  resp_map = {}
  version_map = {}

  # GET ISAs concurrently to find their versions
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    for isa_id in ISA_IDS:
      future = executor.submit(_get_isa, isa_id, session)
      future.add_done_callback(functools.partial(_collect_resp_callback, isa_id, resp_map))

  for isa_id, resp in resp_map.items():
    assert resp.status_code == 200, resp.content
    version = resp.json()['service_area']['version']
    version_map[isa_id] = version

  resp_map = {}

  # Delete ISAs concurrently
  with ThreadPoolExecutor(max_workers=THREAD_COUNT) as executor:
    for isa_id in ISA_IDS:
      future = executor.submit(_delete_isa, isa_id, version_map[isa_id], session)
      future.add_done_callback(functools.partial(_collect_resp_callback, isa_id, resp_map))

  for isa_id, resp in resp_map.items():
    assert resp.status_code == 200, resp.content
