"""A simulated USS exposing public portal endpoints and flying virtual circles.

Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

import collections
import copy
import datetime
import json
import jwt
import logging
import math
import sys

from flask import abort
from flask import Flask
from flask import jsonify
from flask import render_template
from flask import request
from flask import url_for
import iso8601
import jinja2
import pytz
import requests
from rest_framework import status

import config
import interuss_platform

EARTH_CIRCUMFERENCE = 40.075e6  # meters
MAX_QUERY_DIAGONAL = 3600  # meters
REQUIRED_GRID_FIELDS = {'x', 'y', 'zoom', 'version', 'uss', 'timestamp',
                        'minimum_operation_timestamp',
                        'maximum_operation_timestamp', 'flight_info_endpoint',
                        'public_portal_endpoint'}

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('FlightViewer')
webapp = Flask(__name__)  # Global object serving the API

LatLng = collections.namedtuple('LatLng', 'lat lng')

# Set up Jinja
jenv = jinja2.Environment(
  loader=jinja2.FileSystemLoader('./templates'),
  autoescape=jinja2.select_autoescape(['jinja']))
# Make zip function available to Jinja (https://stackoverflow.com/a/5223810)
jenv.globals['zip'] = zip


grid_client = None


@webapp.route('/', methods=['GET'])
@webapp.route('/status', methods=['GET'])
def Status():
  log.debug('Status requested')
  return jsonify({'status': 'success',
                  'message': 'Query /listoperators for more info'})


@webapp.route('/listoperators', methods=['GET'])
def ListOperators():
  log.debug('List operators requested')

  # Retrieve and validate request parameters
  try:
    if 'area' in request.args:
      values = [float(v) for v in request.args['area'].split(',')]
      if len(values) % 2 != 0:
        raise ValueError('CSV of coordinates must in lat,lon pairs')
      border = zip(values[0::2], values[1::2])
      if len(border) < 3:
        raise ValueError('Must specify at least 3 points as an area boundary')
    elif 'center' in request.args:
      values = [float(v) for v in request.args['center'].split(',')]
      if len(values) != 2:
        raise ValueError('Expected center=lat,lng, instead found %d values' %
                         len(values))
      center = LatLng(values[0], values[1])
      border = None
    else:
      raise ValueError('Argument required: either center=lat,lng or '
                       'area=lat,lng,lat,lng,...')
  except ValueError as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)

  if border:
    # Make sure the query area isn't too big
    ll = LatLng(min(p[0] for p in border), min(p[1] for p in border))
    ur = LatLng(max(p[0] for p in border), max(p[1] for p in border))
    dy = (ur.lat - ll.lat) * EARTH_CIRCUMFERENCE / 360
    dx = ((ur.lng - ll.lng) * EARTH_CIRCUMFERENCE *
          math.cos(math.radians(ll.lat)) / 360)
    if math.sqrt(dx*dx + dy*dy) > MAX_QUERY_DIAGONAL:
      abort(status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
            'Bounding area exceeds maximum for privacy')
  else:
    delta_meters = MAX_QUERY_DIAGONAL / (2 * math.sqrt(2))
    dlat = delta_meters * 360 / EARTH_CIRCUMFERENCE
    dlng = delta_meters * math.cos(math.radians(
        center.lat)) * 360 / EARTH_CIRCUMFERENCE
    border = [LatLng(center.lat - dlat, center.lng - dlng),
              LatLng(center.lat - dlat, center.lng + dlng),
              LatLng(center.lat + dlat, center.lng + dlng),
              LatLng(center.lat + dlat, center.lng - dlng)]

  try:
    operators = grid_client.get_operators(border)
  except Exception as e:
    abort(status.HTTP_417_EXPECTATION_FAILED, 'Error querying grid: ' + str(e))

  coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in border)

  operators_args = []
  for operator in operators:
    args = copy.deepcopy(operator)
    args['raw_grid_entry'] = json.dumps(operator, indent=2, sort_keys=True)
    args['entry_name'] = '%s_%d_%d_%d' % (args['uss'], args['x'],
                                          args['y'], args['zoom'])

    args['missing_fields'] = [f for f in REQUIRED_GRID_FIELDS
                              if f not in operator]

    no_id_index = 1
    if 'public_portal_endpoint' in args:
      try:
        args['public_portal_flights'] = grid_client.get_flights(
            args['public_portal_endpoint'], border)
        args['public_portal_flights_raw'] = json.dumps(
            args['public_portal_flights'], indent=2, sort_keys=True)

        for flight in args['public_portal_flights']:
          flight['raw_entry'] = json.dumps(flight, indent=2, sort_keys=True)
          if 'uuid_operation' in flight:
            flight['op_id'] = _shorten(flight['uuid_operation'])
          else:
            flight['op_id'] = '<NO ID %d>' % no_id_index
            no_id_index += 1
          if 'position_history' in flight:
            entry = _most_recent_position(flight['position_history'])
            if not entry:
              flight['position'] = '<UNKNOWN LOCATION>'
            else:
              flight['position'] = '%.5f, %.5f @%.1f' % (entry['latitude'],
                                                         entry['longitude'],
                                                         entry['height'])
          else:
            flight['position'] = '<UNKNOWN LOCATION>'
          flight['short_reg'] = _shorten(flight.get('uuid_reg', ''))
          flight['short_serial'] = _shorten(flight.get('uuid_serial', ''))

          try:
            flight_info = grid_client.get_flight_info(
                args['flight_info_endpoint'], flight['uuid_operation'])
            flight['flight_info_raw'] = json.dumps(
                flight_info, indent=2, sort_keys=True)
          except Exception as e:
            flight['flight_info_error'] = (
              'Error while retrieving flight info: ' + str(e))

      except Exception as e:
        args['public_portal_error'] = (
            'Error querying public_portal_endpoint: ' + str(e))
    else:
      args['public_portal_error'] = ('Missing public_portal_endpoint parameter '
                                     'in GridCellMetaData')

    operators_args.append(args)

  return render_template(
      'listoperators.jinja', operators=operators_args, coords=coords)


def _shorten(identifier):
  if len(identifier) > 11:
    return identifier[0:4] + '...' + identifier[-4:]
  else:
    return  identifier


def _most_recent_position(history):
  recent_time = datetime.datetime.fromtimestamp(0, pytz.utc)
  recent_entry = None
  for entry in history:
    if 'timestamp' in entry:
      timestamp = iso8601.parse_date(entry['timestamp'])
      if timestamp > recent_time:
        recent_time = timestamp
        recent_entry = entry
  return recent_entry


@webapp.before_first_request
def BeforeFirstRequest():
  if grid_client is None:
    Initialize([])


def Initialize(argv):
  options = config.ParseOptions(argv)

  global grid_client
  grid_client = interuss_platform.Client(
    options.nodeurl, int(options.zoom), options.authurl, options.username,
    options.password)

  return options


def main(argv):
  options = Initialize(argv)

  log.info('Starting webserver...')
  webapp.run(host=options.server, port=int(options.port))


# This is what starts everything when run directly as an executable
if __name__ == '__main__':
  main(sys.argv)
