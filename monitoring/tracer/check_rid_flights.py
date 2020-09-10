#!env/bin/python3

import argparse
import logging
from typing import Dict

import requests
import yaml

from monitoring.monitorlib import rid
from monitoring.tracer import polling
from monitoring.tracer.resources import ResourceSet


logging.basicConfig()
_logger = logging.getLogger('check_rid_flights')
_logger.setLevel(logging.DEBUG)


def _json_or_error(resp: requests.Response) -> Dict:
  try:
    json = resp.json()
  except ValueError:
    json = None
  if resp == 200 and json:
    return json
  else:
    info = {
      'request': {
        'url': resp.request.url,
        'token': resp.request.headers.get('Authorization', '<None>'),
      },
      'response': {
        'code': resp.status_code,
        'elapsed': resp.elapsed.total_seconds()
      }
    }
    if json is None:
      info['response']['body'] = resp.content
    else:
      info['response']['json'] = json
    return info


def get_flights(resources: ResourceSet, flights_url: str, include_recent_positions: bool) -> Dict:
  resp = resources.dss_client.get(flights_url, params={
    'view': '{},{},{},{}'.format(
      resources.area.lat_lo().degrees,
      resources.area.lng_lo().degrees,
      resources.area.lat_hi().degrees,
      resources.area.lng_hi().degrees,
    ),
    'include_recent_positions': 'true' if include_recent_positions else 'false',
  }, scope=rid.SCOPE_READ)
  return _json_or_error(resp)


def get_flight_details(resources: ResourceSet, flights_url: str, id: str) -> Dict:
  resp = resources.dss_client.get(flights_url + '/{}/details'.format(id), scope=rid.SCOPE_READ)
  return _json_or_error(resp)


def main():
  parser = argparse.ArgumentParser()
  ResourceSet.add_arguments(parser)
  parser.add_argument('--include-recent-positions', action='store_true', default=False, help='If set, request recent positions when polling for flight data')
  args = parser.parse_args()
  resources = ResourceSet.from_arguments(args)

  isa_result = polling.poll_rid_isas(resources)
  if not isa_result.success:
    print(isa_result.to_json())
    print('Failed to obtain ISAs')

  if not isa_result.success.objects:
    print('No ISAs present in requested area')

  result = {}
  for isa_id, isa in isa_result.success.objects.items():
    flights_url = isa.get('flights_url', None)
    if flights_url is None:
      result[isa_id] = {'error': 'Missing flights_url'}
      continue
    isa_flights = get_flights(resources, flights_url, args.include_recent_positions)
    if 'flights' not in isa_flights['response'].get('json', {}):
      isa_flights['description'] = 'Missing flights field'
      result[isa_id] = {'error': isa_flights}
      continue
    for flight in isa_flights['response']['json']['flights']:
      flight_id = flight.get('id', None)
      if flight_id is None:
        flight['details'] = {'error': {'description': 'Missing id field'}}
        continue
      flight['details'] = get_flight_details(resources, flights_url, flight['id'])
    result[isa_id] = isa_flights

  print(yaml.dump(result))


if __name__ == "__main__":
  main()
