"""The InterUSS Platform Data Node storage API server.

This flexible and distributed system is used to connect multiple USSs operating
in the same general area to share safety information while protecting the
privacy of USSs, businesses, operator and consumers. The system is focused on
facilitating communication amongst actively operating USSs with no details about
UAS operations stored or processed on the InterUSS Platform.

A data node contains all of the API, logic, and data consistency infrastructure
required to perform CRUD (Create, Read, Update, Delete) operations on specific
grid cells. Multiple data nodes can be executed to increase resilience and
availability. This is achieved by a stateless API to service USSs, an
information interface to translate grid cell USS information into the correct
data storage format, and an information consistency store to ensure data is up
to date.

This module is the stateless API to service USSs.


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
import json
# logging is our log infrastructure used for this application
import logging
import math
# OptionParser is our command line parser interface
from optparse import OptionParser
import os
import sys
# Flask is our web services infrastructure
from flask import abort
from flask import Flask
from flask import jsonify
from flask import request
# pyJWT is for decrypting JWT tokens
import jwt
# rest_framework is for HTTP status codes
from rest_framework import status
# Our main class for accessing metadata from the locking system
import storage_interface

# Initialize everything we need
# VERSION = '0.1.0'  # Initial TCL3 release
# VERSION = '0.1.1'  # Pythonized file names and modules
# VERSION = '0.1.2'  # Added OS Environment Variables in addition to cmd line
# VERSION = '0.1.3'  # Added server reconnection logic on lost session
# VERSION = '0.1.4'  # Added utility function to convert lat/lon to slippy
# VERSION = '0.2.0'  # Added OAuth access_token validation
# VERSION = '0.2.1'  # Changed uss_id to use client_id field from NASA
# VERSION = '0.2.2'  # Updated parameter parsing to support swaggerhub
# VERSION = '0.2.3'  # Update overall timestamp in locking metadata on change
# VERSION = '0.2.4'  # Fixed incorrect failed assertion with zero numbered tiles
# VERSION = '0.3.0'  # Changed to locally verifying JWT, removing NASA FIMS link
# VERSION = '0.3.1'  # Added token validation option in test mode
# VERSION = '0.4.0'  # Changed data structure to match v1 of InterUSS Platform
# VERSION = '1.0.0'  # Initial, approved release deployed on GitHub
# VERSION = 'TCL4.0.0'  # Specific branch for TCL4 only
VERSION = 'TCL4.0.1'  # Updated slippy format, added SSL

TESTID = None

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_StorageAPI')
wrapper = None  # Global object API uses for access
webapp = Flask(__name__)  # Global object serving the API


######################################################################
################    WEB SERVICE ENDPOINTS    #########################
######################################################################

@webapp.route('/', methods=['GET'])
@webapp.route('/status', methods=['GET'])
def Status():
  # just a quick status checker, not really a health check
  log.debug('Status handler instantiated...')
  return _FormatResult({'status': 'success',
                        'message': 'OK',
                        'version': VERSION})


@webapp.route('/slippy/<zoom>', methods=['GET'])
def ConvertCoordinatesToSlippy(zoom):
  """Converts an CSV of coords to slippy tile format at the specified zoom.

  Args:
    zoom: zoom level to use for encapsulating the tiles
    Plus posted webarg coords: csv of lon,lat,long,lat,etc.
  Returns:
    200 with tiles array in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  log.info('Convert coordinates to slippy instantiated for %sz...', zoom)
  tiles = []
  coords = _GetRequestParameter('coords', '')
  log.debug('Retrieved coords from web params and split to %s...', coords)
  coordinates = _ValidateCoordinates(coords)
  try:
    zoom = int(zoom)
    if zoom < 0 or zoom > 20:
      raise ValueError
  except ValueError:
    log.error('Invalid parameters for zoom %s, must be integer 0-20...', zoom)
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid parameters for zoom, must be integer 0-20.')
  if not coordinates:
    log.error('Invalid coords %s, must be a CSV of lat,lon...', zoom)
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid coords, must be a CSV of lat,lon,lat,lon...')
  for c in coordinates:
    x, y = _ConvertPointToTile(zoom, c[1], c[0])
    link = 'http://tile.openstreetmap.org/%d/%d/%d.png' % (zoom, x, y)
    tile = {'link': link, 'zoom': zoom, 'x': x, 'y': y}
    if tile not in tiles:
      tiles.append(tile)
  return jsonify({
      'status': 'success',
      'data': {
          'zoom': zoom,
          'grid_cells': tiles,
      }
  })


@webapp.route(
    '/GridCellOperator/<zoom>/<x>/<y>',
    methods=['GET', 'PUT', 'POST', 'DELETE'])
def GridCellOperatorHandler(zoom, x, y):
  """Handles the web service request and routes to the proper function.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    OAuth access_token as part of the header
    Plus posted webargs as needed for PUT/POST and DELETE methods (see below)
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  if ('access_token' in request.headers and TESTID and
      TESTID in request.headers['access_token']):
    uss_id = request.headers['access_token']
  elif TESTID and 'access_token' not in request.headers:
    uss_id = TESTID
  else:
    uss_id = _ValidateAccessToken()
  result = {}
  try:
    zoom = int(zoom)
    x = int(x)
    y = int(y)
  except ValueError:
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid parameters for slippy tile coordinates, must be integers.')
  if not wrapper:
    InitializeConnection(None)
  # Check the request method
  if request.method == 'GET':
    result = _GetGridCellOperator(zoom, x, y)
  elif request.method in ('PUT', 'POST'):
    result = _PutGridCellOperator(zoom, x, y, uss_id)
  elif request.method == 'DELETE':
    result = _DeleteGridCellOperator(zoom, x, y, uss_id)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)


@webapp.route(
    '/GridCellOperation/<zoom>/<x>/<y>/<gufi>',
    methods=['PUT', 'POST', 'DELETE'])
def GridCellOperationHandler(zoom, x, y, gufi):
  """Handles the web service request and routes to the proper function.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    gufi: flight identifier to remove
    OAuth access_token as part of the header
    Plus posted webargs as needed for PUT/POST and DELETE methods (see below)
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  if ('access_token' in request.headers and TESTID and
      TESTID in request.headers['access_token']):
    uss_id = request.headers['access_token']
  elif TESTID and 'access_token' not in request.headers:
    uss_id = TESTID
  else:
    uss_id = _ValidateAccessToken()
  result = {}
  try:
    zoom = int(zoom)
    x = int(x)
    y = int(y)
  except ValueError:
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid parameters for slippy tile coordinates, must be integers.')
  if not wrapper:
    InitializeConnection(None)
  # Check the request method
  if request.method in ('PUT', 'POST'):
    result = _PutGridCellOperation(zoom, x, y, uss_id, gufi)
  elif request.method == 'DELETE':
    result = _DeleteGridCellOperation(zoom, x, y, uss_id, gufi)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)

######################################################################
################       INTERNAL FUNCTIONS    #########################
######################################################################
def _ValidateAccessToken():
  """Checks the access token, aborting if it does not pass.

  Uses an Oauth public key to validate an access token.

  Returns:
    USS identification from OAuth client_id field
  """
  # TODO(hikevin): Replace with OAuth Discovery and JKWS
  secret = os.getenv('INTERUSS_PUBLIC_KEY')
  token = None
  if 'access_token' in request.headers:
    token = request.headers['access_token']
    # ENV variables sometimes don't pass newlines, spec says white space
    # doesn't matter, but pyjwt cares about it, so fix it
    token = token.replace(' PUBLIC ', '_PLACEHOLDER_')
    token = token.replace(' ', '\n')
    token = token.replace('_PLACEHOLDER_', ' PUBLIC ')
  if secret and token:
    try:
      r = jwt.decode(token, secret, algorithms='RS256')
      return r['client_id']
    except jwt.ExpiredSignatureError:
      log.error('Access token has expired.')
      abort(status.HTTP_401_UNAUTHORIZED,
            'OAuth access_token is invalid: token has expired.')
    except jwt.DecodeError:
      log.error('Access token is invalid and cannot be decoded.')
      abort(status.HTTP_400_BAD_REQUEST,
            'OAuth access_token is invalid: token cannot be decoded.')
  else:
    log.error('Attempt to access resource without access_token in header.')
    abort(status.HTTP_403_FORBIDDEN,
          'Valid OAuth access_token must be provided in header.')


def _GetGridCellOperator(zoom, x, y):
  """Provides an instantaneous snapshot of operators for a specific GridCell.

  GridCellOperator provides an instantaneous snapshot of the operators stored
  in a specific GridCell, along with a token to be used when updating. For
  TCL3, this will support a single cell.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
  Returns:
    200 with token and JSON metadata,
    or the nominal 4xx error codes as necessary.
  """
  log.info('Grid cell operators request instantiated for %sz, %s,%s...', zoom, x,
           y)
  result = wrapper.get(zoom, x, y)
  return result


def _PutGridCellOperator(zoom, x, y, uss_id):
  """Updates the operator info stored in a specific slippy GridCell.

    Updates the metadata stored in a specific GridCell using optimistic locking
    behavior, which acquires and releases the lock for the specific GridCell.
    Operation fails if unable to acquire the locks or if the lock has been
    updated since GET GridCellOperator was originally called (based on token).
  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
  Plus posted webargs:
    sync_token: the token retrieved in the original GET GridCellOperator,
    scope: The submitting USS scope for the web service endpoint (used for OAuth
      access),
    operation_endpoint: the submitting USS endpoint where all flights in this
      cell can be retrieved from,
    operation_format: The output format for the USS web service endpoint (i.e.
      NASA, GUTMA),
    minimum_operation_timestamp: the lower time bound of all of the USSs flights
      in this grid cell.
    maximum_operation_timestamp: the upper time bound of all of the USSs flights
      in this grid cell.

  Returns:
    200 and a new token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cell operator submit instantiated for %sz, %s, %s...',
           zoom, x, y)
  sync_token = _GetRequestParameter('sync_token', None)
  if not sync_token and 'sync_token' in request.headers:
    sync_token = request.headers['sync_token']
  baseurl = _GetRequestParameter('uss_baseurl', None)
  announce = _GetRequestParameter('announcement_level', None)
  operations = _GetRequestParameter('operations', None)
  minimum_operation_timestamp = _GetRequestParameter(
      'minimum_operation_timestamp', None)
  maximum_operation_timestamp = _GetRequestParameter(
      'maximum_operation_timestamp', None)
  errorfield = errormsg = None
  if not sync_token:
    errorfield = 'sync_token'
  elif not uss_id:
    errorfield = 'uss_id'
    errormsg = 'USS identifier not received from OAuth token check.'
  elif not baseurl:
    errorfield = 'uss_baseurl'
  elif not announce:
    errorfield = 'announcement_level'
  elif not minimum_operation_timestamp:
    errorfield = 'minimum_operation_timestamp'
  elif not maximum_operation_timestamp:
    errorfield = 'maximum_operation_timestamp'
  if errorfield:
    if not errormsg:
      errormsg = errorfield + (
          ' must be provided in the form data request to add to a '
          'GridCell.')
    result = {
        'status': 'error',
        'code': status.HTTP_400_BAD_REQUEST,
        'message': errormsg
    }
  else:
    result = wrapper.set(zoom, x, y, sync_token, uss_id, baseurl,
                         announce, minimum_operation_timestamp,
                         maximum_operation_timestamp, operations)
  return result


def _DeleteGridCellOperator(zoom, x, y, uss_id):
  """Removes the USS entry in the metadata stored in a specific GridCell.

  Removes the USS entry in the metadata using optimistic locking behavior, which
  acquires and releases the lock for the specific GridCell. Operation fails if
  unable to acquire the locks.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
  Returns:
    200 and a new sync_token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cell operator delete instantiated for %sz, %s, %s...',
           zoom, x, y)
  if uss_id:
    result = wrapper.delete(zoom, x, y, uss_id)
  else:
    result = {
        'status':
            'fail',
        'code':
            status.HTTP_400_BAD_REQUEST,
        'message':
            """uss_id must be provided in the request to
              delete a USS from a GridCell."""
    }
  return result

def _PutGridCellOperation(zoom, x, y, uss_id, gufi):
  """Puts a single operation in the metadata stored in a specific GridCell.

  Upserts the operation entry in the metadata using optimistic locking behavior,
  which acquires and releases the lock for the specific GridCell. Operation
  fails if unable to acquire the locks.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
    gufi: flight identifier to add/update
  Returns:
    200 and a new sync_token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cell operation upsert instantiated for %sz, %s, %s, %s...',
           zoom, x, y, gufi)
  sync_token = _GetRequestParameter('sync_token', None)
  if not sync_token and 'sync_token' in request.headers:
    sync_token = request.headers['sync_token']
  gufi = _GetRequestParameter('gufi', None)
  signature = _GetRequestParameter('operation_signature', None)
  begin = _GetRequestParameter('effective_time_begin', None)
  end = _GetRequestParameter('effective_time_end', None)
  errorfield = errormsg = None
  if not sync_token:
    errorfield = 'sync_token'
  elif not uss_id:
    errorfield = 'uss_id'
    errormsg = 'USS identifier not received from OAuth token check.'
  elif not gufi:
    errorfield = 'gufi'
  elif not signature:
    errorfield = 'operation_signature'
  elif not begin:
    errorfield = 'effective_time_begin'
  elif not end:
    errorfield = 'effective_time_end'
  if errorfield:
    if not errormsg:
      errormsg = errorfield + (
        ' must be provided in the form data request to add to a '
        'GridCell.')
    result = {
      'status': 'error',
      'code': status.HTTP_400_BAD_REQUEST,
      'message': errormsg
    }
  else:
    result = wrapper.set_operation(zoom, x, y, sync_token, uss_id, gufi,
                                   signature, begin, end)
  return result

def _DeleteGridCellOperation(zoom, x, y, uss_id, gufi):
  """Removes a single operation in the metadata stored in a specific GridCell.

  Removes the operation entry in the metadata using optimistic locking behavior,
  which acquires and releases the lock for the specific GridCell. Operation
  fails if unable to acquire the locks.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
    gufi: flight identifier to remove
  Returns:
    200 and a new sync_token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cell operation delete instantiated for %sz, %s, %s, %s...',
           zoom, x, y, gufi)
  if uss_id:
    result = wrapper.delete_operation(zoom, x, y, uss_id, gufi)
  else:
    result = {
      'status':
        'fail',
      'code':
        status.HTTP_400_BAD_REQUEST,
      'message':
        """uss_id must be provided in the request to
          delete a USS from a GridCell."""
    }
  return result


def _GetRequestParameter(name, default):
  """Parses a web request parameter, regardless of how it was passed in.

  Args:
    name: request parameter name
    default: default value to return if not found
  Returns:
    Value if found, default if not found
  """
  if request.json:
    r = default if name not in request.json else request.json[name]
  elif request.args:
    r = request.args.get(name, default)
  elif request.form:
    r = default if name not in request.form else request.form[name]
  elif request.data:
    rjson = json.loads(request.data)
    r = default if name not in rjson else rjson[name]
  else:
    log.error('Request is in an unknown format: %s', str(request))
    r = default
  return r


def _ValidateCoordinates(csv):
  """Converts and validates string of CSV coords into array of coords."""
  result = []
  try:
    coords = csv.split(',')
    if len(coords) % 2 != 0:
      raise ValueError
  except ValueError:
    return None
  log.debug('Split coordinates to %s and passed early validation...', coords)
  for a, b in _Pairwise(coords):
    try:
      lon = float(a)
      lat = float(b)
      if lat >= 90 or lat <= -90 or lon >= 180 or lon <= -180:
        raise ValueError
    except ValueError:
      return None
    result.append((lon, lat))
  return result


def _Pairwise(it):
  """Iterator for sets of lon,lat in an array."""
  it = iter(it)
  while True:
    yield next(it), next(it)


def _ConvertPointToTile(zoom, longitude, latitude):
  """Actual calculation from lat/lon to tile at specific zoom."""
  latitude_rad = math.radians(latitude)
  n = 2.0**zoom
  xtile = int((longitude + 180.0) / 360.0 * n)
  ytile = int(
      (1.0 - math.log(math.tan(latitude_rad) +
                      (1 / math.cos(latitude_rad))) / math.pi) / 2.0 * n)
  return xtile, ytile


def _FormatResult(result):
  """Formats the result for returning via the web service.

  Args:
    result: JSend version of the result, complete with code if in error
  Returns:
    Aborts if the code is not 200, otherwise returns JSON formatted response
  """
  if 'code' in result and str(result['code']) != '200':
    abort(result['code'], result['message'])
  else:
    return jsonify(result)


def InitializeConnection(argv):
  """Initializes the wrapper and the connection to the zookeeper servers.

  Uses the command line parameters as argv, which can be altered as needed for
  testing.

  Args:
    argv: Command line parameters
  Returns:
    Host and port to use for the server
  """
  global wrapper, TESTID
  log.debug('Parsing command line arguments...')
  parser = OptionParser(
      usage='usage: %prog [options]', version='%prog ' + VERSION)
  parser.add_option(
      '-z',
      '--zookeeper-servers',
      dest='connectionstring',
      help='Specific zookeeper server connection string, '
      'server:port,server:port...'
      '[or env variable INTERUSS_CONNECTIONSTRING]',
      metavar='CONNECTIONSTRING')
  parser.add_option(
      '-s',
      '--server',
      dest='server',
      default=os.getenv('INTERUSS_API_SERVER', '127.0.0.1'),
      help='Specific server name to use on this machine for the web services '
      '[or env variable INTERUSS_API_SERVER]',
      metavar='SERVER')
  parser.add_option(
      '-p',
      '--port',
      dest='port',
      default=os.getenv('INTERUSS_API_PORT', '5000'),
      help='Specific port to use on this machine for the web services '
      '[or env variable INTERUSS_API_PORT]',
      metavar='PORT')
  parser.add_option(
      '-v',
      '--verbose',
      action='store_true',
      dest='verbose',
      default=False,
      help='Verbose (DEBUG) logging [or env variable INTERUSS_VERBOSE]')
  parser.add_option(
      '-t',
      '--testid',
      dest='testid',
      default=False,
      help='Force testing mode with test data located in specific test id  '
      '[or env variable INTERUSS_TESTID]',
      metavar='TESTID')
  parser.add_option(
      '-a',
      '--ssladhoc',
      action='store_true',
      dest='ssladhoc',
      default=False,
      help='Enable ad-hoc TLS encryption')
  (options, args) = parser.parse_args(argv)
  del args
  if options.verbose or os.environ.get('INTERUSS_VERBOSE'):
    log.setLevel(logging.DEBUG)
  log.debug('Initializing USS metadata manager...')
  wrapper = storage_interface.USSMetadataManager(
      os.getenv('INTERUSS_CONNECTIONSTRING', options.connectionstring))
  if options.verbose or os.environ.get('INTERUSS_VERBOSE'):
    wrapper.set_verbose()
  if options.testid or os.environ.get('INTERUSS_TESTID'):
    TESTID = os.getenv('INTERUSS_TESTID', options.testid)
    wrapper.set_testmode(TESTID)
    wrapper.delete_testdata(TESTID)
  return options.server, options.port, options.ssladhoc


def TerminateConnection():
  global wrapper
  wrapper = None


def main(argv):
  if not os.environ.get('INTERUSS_PUBLIC_KEY'):
    log.error('INTERUSS_PUBLIC_KEY environment variable must be set.')
    sys.exit(-1)
  else:
    log.debug(
        """Instantiated application, parsing commandline
      %s and initializing connection...""", str(argv))
    host, port, ssl_adhoc = InitializeConnection(argv)
    log.info('Starting webserver...')
    webapp.run(host=host, port=int(port),
               ssl_context='adhoc' if ssl_adhoc else None)


# this is what starts everything
if __name__ == '__main__':
  main(sys.argv)
