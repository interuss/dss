"""Basic ISA tests with hundreds of ISAs in the SAME area created CONCURRENTLY:

  - create ISAs with a 60 minute expiry concurrently
  - get by IDs concurrently
  - search with area
  - delete concurrently
"""

import asyncio
import datetime
import re

from monitoring.monitorlib import rid
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.rid import SCOPE_READ, SCOPE_WRITE, ISA_PATH
from monitoring.monitorlib.testing import assert_datetimes_are_equal
from monitoring.prober.infrastructure import register_resource_type
from . import common


THREAD_COUNT = 10
FLIGHTS_URL = 'https://example.com/dss'
ISA_TYPES = [register_resource_type(224 + i, 'Operational intent {}'.format(i)) for i in range(100)]
# Semaphore is added to limit the number of simultaneous requests,
# default is 100.
SEMAPHORE = asyncio.Semaphore(20)


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


async def _put_isa(isa_id, req, session_ridv1):
  async with SEMAPHORE:
    return isa_id, await session_ridv1.put('{}/{}'.format(ISA_PATH, isa_id), json=req, scope=SCOPE_WRITE)

async def _get_isa(isa_id, session_ridv1):
  async with SEMAPHORE:
    return isa_id, await session_ridv1.get('{}/{}'.format(ISA_PATH, isa_id), scope=SCOPE_READ)


async def _delete_isa(isa_id, version, session_ridv1):
  async with SEMAPHORE:
    return isa_id, await session_ridv1.delete('{}/{}/{}'.format(ISA_PATH, isa_id, version), scope=SCOPE_WRITE)



def test_ensure_clean_workspace(ids, session_ridv1):
  for isa_id in map(ids, ISA_TYPES):
    resp = session_ridv1.get('{}/{}'.format(ISA_PATH, isa_id), scope=SCOPE_READ)
    if resp.status_code == 200:
      version = resp.json()['service_area']['version']
      resp = session_ridv1.delete('{}/{}/{}'.format(ISA_PATH, isa_id, version), scope=SCOPE_WRITE)
      assert resp.status_code == 200, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


@default_scope(SCOPE_WRITE)
def test_create_isa_concurrent(ids, session_ridv1_async):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)
  req = _make_isa_request(time_start, time_end)
  resp_map = {}

  # Create ISAs concurrently
  loop = asyncio.get_event_loop()
  results = loop.run_until_complete(
    asyncio.gather(*[_put_isa(isa_id, req, session_ridv1_async) for isa_id in map(ids, ISA_TYPES)]))
  for isa_id, resp in results:
    assert resp[0] == 200, resp[1]
    data = resp[1]
    assert data['service_area']['id'] == isa_id
    assert data['service_area']['flights_url'] == 'https://example.com/dss'
    assert_datetimes_are_equal(data['service_area']['time_start'], req['extents']['time_start'])
    assert_datetimes_are_equal(data['service_area']['time_end'], req['extents']['time_end'])
    assert re.match(r'[a-z0-9]{10,}$', data['service_area']['version'])
    assert 'subscribers' in data


@default_scope(SCOPE_READ)
def test_get_isa_by_ids_concurrent(ids, session_ridv1_async):
  resp_map = {}

  # Get ISAs concurrently
  loop = asyncio.get_event_loop()
  results = loop.run_until_complete(
    asyncio.gather(*[_get_isa(isa_id, session_ridv1_async) for isa_id in map(ids, ISA_TYPES)]))
  for isa_id, resp in results:
    assert resp[0] == 200, resp[1]

    data = resp[1]
    assert data['service_area']['id'] == isa_id
    assert data['service_area']['flights_url'] == FLIGHTS_URL


@default_scope(SCOPE_READ)
def test_get_isa_by_search(ids, session_ridv1):
  resp = session_ridv1.get('{}?area={}'.format(ISA_PATH, common.GEO_POLYGON_STRING))

  assert resp.status_code == 200, resp.content
  found_isa_ids = [x['id'] for x in resp.json()['service_areas']]
  assert len(_intersection(map(ids, ISA_TYPES), found_isa_ids)) == len(ISA_TYPES)


def test_delete_isa_concurrent(ids, session_ridv1_async):
  resp_map = {}
  version_map = {}

  # GET ISAs concurrently to find their versions
  loop = asyncio.get_event_loop()
  results = loop.run_until_complete(
    asyncio.gather(*[_get_isa(isa_id, session_ridv1_async) for isa_id in map(ids, ISA_TYPES)]))

  for isa_id, resp in results:
    assert resp[0] == 200, resp[1]
    version = resp[1]['service_area']['version']
    version_map[isa_id] = version


  # Delete ISAs concurrently
  loop = asyncio.get_event_loop()
  results = loop.run_until_complete(
    asyncio.gather(*[_delete_isa(isa_id, version_map[isa_id], session_ridv1_async) for isa_id in map(ids, ISA_TYPES)]))

  for isa_id, resp in results:
    assert resp[0], resp[1]
