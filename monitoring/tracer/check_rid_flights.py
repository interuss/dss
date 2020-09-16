#!env/bin/python3

import argparse
import json
import logging

import yaml

from monitoring.monitorlib import fetch
import monitoring.monitorlib.fetch.rid
from monitoring.tracer.resources import ResourceSet


logging.basicConfig()
_logger = logging.getLogger('check_rid_flights')
_logger.setLevel(logging.DEBUG)


def main():
  parser = argparse.ArgumentParser()
  ResourceSet.add_arguments(parser)
  parser.add_argument('--include-recent-positions', action='store_true', default=False, help='If set, request recent positions when polling for flight data')
  args = parser.parse_args()
  resources = ResourceSet.from_arguments(args)

  result = fetch.rid.all_flights(resources.dss_client, resources.area, args.include_recent_positions, True)

  print(yaml.dump(result))


if __name__ == "__main__":
  main()
