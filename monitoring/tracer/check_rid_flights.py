#!env/bin/python3

import argparse
import datetime
import logging
from typing import Dict

import requests
import s2sphere
import yaml

from monitoring.monitorlib import rid
from monitoring.tracer import formatting, polling
from monitoring.tracer.resources import ResourceSet


logging.basicConfig()
_logger = logging.getLogger('check_rid_flights')
_logger.setLevel(logging.DEBUG)


def get_flights(resources: ResourceSet, flights_url: str, area: s2sphere.LatLngRect, include_recent_positions: bool) -> Dict:
  t0 = datetime.datetime.utcnow()
  resp = resources.dss_client.get(flights_url, params={
    'view': '{},{},{},{}'.format(
      area.lat_lo().degrees,
      area.lng_lo().degrees,
      area.lat_hi().degrees,
      area.lng_hi().degrees,
    ),
    'include_recent_positions': 'true' if include_recent_positions else 'false',
  }, scope=rid.SCOPE_READ)
  return {
    'request': formatting.describe_request(resp.request, t0),
    'response': formatting.describe_response(resp),
  }


def get_flight_details(resources: ResourceSet, flights_url: str, id: str) -> Dict:
  t0 = datetime.datetime.utcnow()
  resp = resources.dss_client.get(flights_url + '/{}/details'.format(id), scope=rid.SCOPE_READ)
  return {
    'request': formatting.describe_request(resp.request, t0),
    'response': formatting.describe_response(resp),
  }


def get_all_flights(resources: ResourceSet, area: s2sphere.LatLngRect, include_recent_positions: bool) -> Dict:
  isa_result = polling.poll_rid_isas(resources, area)
  if not isa_result.success:
    return {
      'error': {
        'description': 'Failed to obtain ISAs',
        'response': isa_result.to_json(),
      }
    }
  if not isa_result.success.objects:
    return {
      'error': {
        'description': 'No ISAs present in requested area',
      }
    }

  result = {}
  for isa_id, isa in isa_result.success.objects.items():
    flights_url = isa.get('flights_url', None)
    if flights_url is None:
      result[isa_id] = {'error': {'description': 'Missing flights_url'}}
      continue
    isa_flights = get_flights(resources, flights_url, area, include_recent_positions)
    if 'flights' not in isa_flights['response'].get('json', {}):
      if isa_flights['response']['code'] != 200:
        isa_flights['description'] = 'USS returned {}'.format(isa_flights['response']['code'])
      else:
        isa_flights['description'] = 'Missing flights field'
      result[isa_id] = {'error': isa_flights}
      continue
    for flight in isa_flights['response']['json']['flights']:
      flight_id = flight.get('id', None)
      if flight_id is None:
        flight['details (separate query)'] = {'error': {'description': 'Missing id field'}}
        continue
      flight['details (separate query)'] = get_flight_details(resources, flights_url, flight['id'])
    result[isa_id] = isa_flights

  return result


def main():
  parser = argparse.ArgumentParser()
  ResourceSet.add_arguments(parser)
  parser.add_argument('--include-recent-positions', action='store_true', default=False, help='If set, request recent positions when polling for flight data')
  args = parser.parse_args()
  resources = ResourceSet.from_arguments(args)

  result = get_all_flights(resources, resources.area, args.include_recent_positions)

  print(yaml.dump(result))


if __name__ == "__main__":
  main()
