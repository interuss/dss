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

import datetime
import json
import jwt
import logging
import math
import sys

from flask import abort
from flask import Flask
from flask import jsonify
from flask import request
import requests
from rest_framework import status

import config
import formatting
import interuss_platform
import simulation

EARTH_CIRCUMFERENCE = 40.075e6  # meters
MAX_QUERY_DIAGONAL = 3600  # meters

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('DizzySim')
webapp = Flask(__name__)  # Global object serving the API


flightsim = None
public_key = None


@webapp.route('/', methods=['GET'])
@webapp.route('/status', methods=['GET'])
def Status():
  log.debug('Status requested')
  return jsonify({'status': 'success',
                  'flights': flightsim.get_flights_info()})


@webapp.route('/launch', methods=['POST'])
def Launch():
  log.debug('Launch requested')
  _ValidateAccessToken()
  flightsim.launch()
  return jsonify({'status': 'success',
                  'flights': flightsim.get_flights_info()})


@webapp.route('/land/<i>', methods=['POST'])
def Land(i):
  log.debug('Land %s requested', i)
  _ValidateAccessToken()
  try:
    flightsim.land(int(i))
  except Exception as e:
    abort(status.HTTP_400_BAD_REQUEST, str(e))

  return jsonify({'status': 'success',
                  'flights': flightsim.get_flights_info()})


@webapp.route('/public_portal/<coords>', methods=['GET'])
@webapp.route('/public_portal/USS_public_portal_endpoint/<coords>', methods=['GET'])
def PublicPortal(coords):
  log.debug('Public portal queried')

  _ValidateAccessToken()

  # Retrieve and validate request parameters
  try:
    values = [float(v) for v in coords.split(',')]
    if len(values) % 2 != 0:
      raise ValueError('CSV of coordinates must in lat,lon pairs')
    pts = zip(values[0::2], values[1::2])
    if len(pts) < 3:
      raise ValueError('Must specify at least 3 points as an area boundary')
  except ValueError as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)

  if request.args and 'history' in request.args:
    history = datetime.timedelta(seconds=float(request.args['history']))
    if history > datetime.timedelta(seconds=60):
      abort(status.HTTP_400_BAD_REQUEST,
            'History duration may not exceed 60 seconds')
  else:
    history = datetime.timedelta(seconds=60)

  # Make sure the query area isn't too big
  ll = simulation.LatLng(min(p[0] for p in pts), min(p[1] for p in pts))
  ur = simulation.LatLng(max(p[0] for p in pts), max(p[1] for p in pts))
  dy = (ur.lat - ll.lat) * EARTH_CIRCUMFERENCE / 360
  dx = ((ur.lng - ll.lng) * EARTH_CIRCUMFERENCE *
        math.cos(math.radians(ll.lat)) / 360)
  if math.sqrt(dx*dx + dy*dy) > MAX_QUERY_DIAGONAL:
    abort(status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
          'Bounding area exceeds maximum for privacy')

  try:
    return jsonify({
      'status': 'success',
      'data': {
        'timestamp': formatting.timestamp(datetime.datetime.utcnow()),
        'telemetries': flightsim.get_telemetries(history),
        'volumes': []}})
  except Exception as e:
    abort(status.HTTP_500_INTERNAL_SERVER_ERROR, str(e))


@webapp.route('/flight_info/<uuid_operation>', methods=['GET'])
@webapp.route('/flight_info/USS_flight_info_endpoint/<uuid_operation>', methods=['GET'])
def FlightInfo(uuid_operation):
  _ValidateAccessToken()

  flight_info = flightsim.get_flight_info(uuid_operation)
  if not flight_info:
    abort(status.HTTP_404_NOT_FOUND,
          'No active flight found with UUID ' + uuid_operation)
  else:
    return jsonify({
      'status': 'success',
      'data': flight_info})


@webapp.before_first_request
def BeforeFirstRequest():
  if flightsim is None:
    Initialize([])


def _ValidateAccessToken():
  if 'Authorization' in request.headers:
    token = request.headers['Authorization'].replace('Bearer ', '')
  elif 'access_token' in request.headers:
    token = request.headers['access_token']
  else:
    abort(status.HTTP_401_UNAUTHORIZED,
          'access_token was not included in request')

  try:
    jwt.decode(token, public_key, algorithms='RS256')
  except jwt.ExpiredSignatureError:
    log.error('Access token has expired.')
    abort(status.HTTP_401_UNAUTHORIZED,
          'access_token is invalid: token has expired.')
  except jwt.DecodeError:
    log.error('Access token is invalid and cannot be decoded.')
    abort(status.HTTP_400_BAD_REQUEST,
          'access_token is invalid: token cannot be decoded.')


def Initialize(argv):
  options = config.ParseOptions(argv)

  latlng = options.origin.split(',')
  if len(latlng) != 2:
    raise ValueError('Invalid origin: ' + options.origin)

  with open(options.hanger, 'r') as f:
    hanger = json.loads(f.read())

  global public_key
  if options.authpublickey.startswith('http'):
    log.info('Downloading auth public key from ' + options.authpublickey)
    response = requests.get(options.authpublickey)
    response.raise_for_status()
    public_key = response.content
  else:
    public_key = options.authpublickey
  public_key = public_key.replace(' PUBLIC ', '_PLACEHOLDER_')
  public_key = public_key.replace(' ', '\n')
  public_key = public_key.replace('_PLACEHOLDER_', ' PUBLIC ')

  grid_client = interuss_platform.Client(
    options.nodeurl, int(options.zoom), options.authurl, options.username,
    options.password, options.baseurl + '/public_portal',
                      options.baseurl + '/flight_info')

  global flightsim
  flightsim = simulation.FlightSim(
    origin=simulation.LatLng(lat=float(latlng[0]), lng=float(latlng[1])),
    radius=float(options.radius),
    period=datetime.timedelta(seconds=float(options.flightperiod)),
    interval=datetime.timedelta(seconds=float(options.flightinterval)),
    min_altitude=float(options.minaltitude),
    max_altitude=float(options.maxaltitude),
    hanger=hanger,
    grid_client=grid_client)

  return options


def main(argv):
  options = Initialize(argv)

  log.info('Starting webserver...')
  webapp.run(host=options.server, port=int(options.port))


# This is what starts everything when run directly as an executable
if __name__ == '__main__':
  main(sys.argv)
