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
# logging is our log infrastructure used for this application
import logging
# OptionParser is our command line parser interface
from optparse import OptionParser
import os
import sys
# Flask is our web services infrastructure
from flask import abort
from flask import Flask
from flask import jsonify
from flask import request
# rest_framework is for HTTP status codes
from rest_framework import status

# Tools for checking client authorization
import authorization
# Our main class for accessing metadata from the locking system
import storage_interface
# Tools for slippy conversion
import slippy_util

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
# VERSION = '1.0.1.001'  # Bug fixes for slippy, dates, and OAuth key
# VERSION = '1.0.2.001'  # Refactored to run with gunicorn
# VERSION = '1.0.2.002'  # Standardize OAuth Authorization header, docker fix
# VERSION = '1.0.2.003'  # slippy utility updates to support point/path/polygon
# VERSION = '1.0.2.004'  # slippy non-breaking api changes to support path/polygon
# VERSION = '1.1.0.005'  # api changes to support multi-grid GET/PUT/DEL
# VERSION = 'PublicPortal1.1.1.006'  # Added public portal support
VERSION = 'PublicPortal1.1.1.007'  # Fixed multi-cell bug

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_StorageAPI')
wrapper = None  # Global object API uses for access
webapp = Flask(__name__)  # Global object serving the API
auth = None  # Global object providing authorization


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


@webapp.route('/introspect', methods=['GET'])
def Introspect():
  #  status checker of FIMS authorization token (access_token)
  log.debug('Status handler instantiated...')
  uss_id = auth.ValidateAccessToken()
  return _FormatResult({
      'status': 'success',
      'message': 'ACCESS TOKEN VALID',
      'data': {
          'uss_id': uss_id
      }
  })


@webapp.route('/introspect/<zoom>/<x>/<y>', methods=['GET'])
def IntrospectTile(zoom, x, y):
  """Status checker of FIMS authorization token (access_token).
  Args:
    zoom: zoom level to use for encapsulating the tile
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
  """
  log.debug('Status handler instantiated...')
  try:
    zoom = int(zoom)
    tiles = ((int(x), int(y)), )
  except ValueError as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)
  uss_id = auth.ValidateAccessToken(authorization.JoinZoom(zoom, tiles))
  return _FormatResult({
    'status': 'success',
    'message': 'ACCESS TOKEN VALID',
    'data': {
      'uss_id': uss_id
    }
  })


@webapp.route('/introspect/<zoom>', methods=['GET'])
def IntrospectTiles(zoom):
  """Status checker of FIMS authorization token (access_token).
  Args:
    zoom: zoom level to use for encapsulating the tiles
    Plus posted webargs
     coords: csv of lon,lat,long,lat,etc.
     coord_type: (optional) type of coords - point (default), path, polygon
  """
  log.debug('Status handler instantiated...')
  try:
    zoom = int(zoom)
    tiles = _ConvertRequestToTiles(zoom)
  except (ValueError, TypeError, OverflowError) as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)
  uss_id = auth.ValidateAccessToken(authorization.JoinZoom(zoom, tiles))
  return _FormatResult({
    'status': 'success',
    'message': 'ACCESS TOKEN VALID',
    'data': {
      'uss_id': uss_id
    }
  })


@webapp.route('/slippy/<zoom>', methods=['GET'])
def ConvertCoordinatesToSlippy(zoom):
  """Converts an CSV of coords to slippy tile format at the specified zoom.
  Args:
    zoom: zoom level to use for encapsulating the tiles
    Plus posted webargs
     coords: csv of lon,lat,long,lat,etc.
     coord_type: (optional) type of coords - point (default), path, polygon
  Returns:
    200 with tiles array in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  log.info('Convert coordinates to slippy instantiated for %sz...', zoom)
  try:
    zoom = int(zoom)
    tiles = _ConvertRequestToTiles(zoom)
    result = []
    for x, y in tiles:
      link = 'http://tile.openstreetmap.org/%d/%d/%d.png' % (zoom, x, y)
      tile = {'link': link, 'zoom': zoom, 'x': x, 'y': y}
      if tile not in result:
        result.append(tile)
  except (ValueError, TypeError, OverflowError) as e:
    log.error('/slippy error: %s...', e.message)
    abort(status.HTTP_400_BAD_REQUEST, e.message)

  return jsonify({
    'status': 'success',
    'data': {
      'zoom': zoom,
      'grid_cells': result,
    }
  })


@webapp.route(
    '/GridCellMetaData/<zoom>/<x>/<y>',
    methods=['GET', 'PUT', 'POST', 'DELETE'])
def GridCellMetaDataHandler(zoom, x, y):
  """Handles the web service request and routes to the proper function.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    OAuth access_token as part of the header
    Plus posted webargs as needed for PUT/POST and DELETE methods
      (see internal functions Get/Put/Delete metadata below)
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  result = {}
  try:
    zoom = int(zoom)
    x = int(x)
    y = int(y)
  except ValueError:
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid parameters for slippy tile coordinates, must be integers.')
  uss_id = auth.ValidateAccessToken(((zoom, x, y),))
  if request.method == 'GET':
    result = _GetGridCellMetaData(zoom, x, y)
  elif request.method in ('PUT', 'POST'):
    result = _PutGridCellMetaData(zoom, x, y, uss_id)
  elif request.method == 'DELETE':
    result = _DeleteGridCellMetaData(zoom, x, y, uss_id)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)


@webapp.route(
  '/GridCellsMetaData/<zoom>',
  methods=['GET', 'PUT', 'POST', 'DELETE'])
def GridCellsMetaDataHandler(zoom):
  """Handles the web service request for multi-grid operations.

  Args:
    zoom: zoom level in slippy tile format
    OAuth access_token as part of the header
    Plus posted webargs:
      coords: csv of lon,lat,long,lat,etc.
      coord_type: (optional) type of coords - point (default), path, polygon
       and additional as needed for PUT/POST and DELETE methods
       (see internal functions Get/Put/Delete metadata below)

  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  result = {}
  try:
    zoom = int(zoom)
    tiles = _ConvertRequestToTiles(zoom)
    if len(tiles) > slippy_util.TILE_LIMIT:
      raise OverflowError('Limit of %d tiles impacted exceeded (%d)'
                          % (slippy_util.TILE_LIMIT, len(tiles)))
  except (ValueError, TypeError) as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)
  except OverflowError as e:
    abort(status.HTTP_413_REQUEST_ENTITY_TOO_LARGE, e.message)
  uss_id = auth.ValidateAccessToken(authorization.JoinZoom(zoom, tiles))
  if request.method == 'GET':
    result = _GetGridCellsMetaData(zoom, tiles)
  elif request.method in ('PUT', 'POST'):
    result = _PutGridCellsMetaData(zoom, tiles, uss_id)
  elif request.method == 'DELETE':
    result = _DeleteGridCellsMetaData(zoom, tiles, uss_id)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)


######################################################################
################       INTERNAL FUNCTIONS    #########################
######################################################################

def _GetGridCellMetaData(zoom, x, y):
  """Provides an instantaneous snapshot of the metadata for a specific GridCell.

  GridCellMetaData provides an instantaneous snapshot of the metadata stored
  in a specific GridCell, along with a token to be used when updating.

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
  Returns:
    200 with token and JSON metadata,
    or the nominal 4xx error codes as necessary.
  """
  log.info('Grid cell metadata request instantiated for %sz, %s,%s...', zoom, x,
           y)
  result = wrapper.get(zoom, x, y)
  return result


def _PutGridCellMetaData(zoom, x, y, uss_id):
  """Updates the metadata stored in a specific slippy GridCell.

    Updates the metadata stored in a specific GridCell using optimistic locking
    behavior. Operation fails if the metadata has been updated since
    GET GridCellMetadata was originally called (based on token).

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
  Plus posted webargs:
    sync_token: the token retrieved in the original GET GridCellMetadata,
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
  log.info('Grid cell metadata submit instantiated for %sz, %s,%s...', zoom, x,
           y)
  sync_token = _GetRequestParameter('sync_token', None)
  if not sync_token and 'sync_token' in request.headers:
    sync_token = request.headers['sync_token']
  scope = _GetRequestParameter('scope', None)
  operation_endpoint = _GetRequestParameter('operation_endpoint', '')
  operation_format = _GetRequestParameter('operation_format', '')
  minimum_operation_timestamp = _GetRequestParameter(
      'minimum_operation_timestamp', None)
  maximum_operation_timestamp = _GetRequestParameter(
      'maximum_operation_timestamp', None)
  public_portal_endpoint = _GetRequestParameter('public_portal_endpoint', '')
  flight_info_endpoint = _GetRequestParameter('flight_info_endpoint', '')
  errorfield = errormsg = None
  if operation_endpoint and not sync_token:
    errorfield = 'sync_token'
  elif not uss_id:
    errorfield = 'uss_id'
    errormsg = 'USS identifier not received from OAuth token check.'
  elif not scope:
    errorfield = 'scope'
  elif operation_endpoint and not operation_format:
    errorfield = 'operation_format'
  elif not minimum_operation_timestamp:
    errorfield = 'minimum_operation_timestamp'
  elif not maximum_operation_timestamp:
    errorfield = 'maximum_operation_timestamp'
  elif (not operation_endpoint and
        not public_portal_endpoint and
        not flight_info_endpoint):
    errorfield = ('operation_endpoint, public portal_endpoint, or '
                  'flight_info_endpoint')
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
    result = wrapper.set(zoom, x, y, sync_token, uss_id, scope,
                         operation_format, operation_endpoint,
                         minimum_operation_timestamp,
                         maximum_operation_timestamp, public_portal_endpoint,
                         flight_info_endpoint)
  return result


def _DeleteGridCellMetaData(zoom, x, y, uss_id):
  """Removes the USS entry in the metadata stored in a specific GridCell.

  Removes the USS entry in the metadata using optimistic locking behavior.

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
  log.info('Grid cell metadata delete instantiated for %sz, %s,%s...', zoom, x,
           y)
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


def _GetGridCellsMetaData(zoom, tiles):
  """Provides an instantaneous snapshot of the metadata for a multiple GridCells

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to retrieve
  Returns:
    200 with token and JSON metadata,
    or the nominal 4xx error codes as necessary.
  """
  log.info('Grid cells metadata request instantiated for %sz, %s...',
           zoom, str(tiles))
  result = wrapper.get_multi(zoom, tiles)
  return result

def _PutGridCellsMetaData(zoom, tiles, uss_id):
  """Updates the metadata stored in multiple GridCells.

    Updates the metadata stored in a multiple GridCell using optimistic locking
    behavior. Operation fails if the metadata has been updated since
    GET GridCellsMetadata was originally called (based on sync_token).

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to retrieve
    uss_id: the plain text identifier for the USS from OAuth
  Plus posted webargs:
    sync_token: the composite sync_token retrieved in the
      original GET GridCellsMetadata,
    scope: The submitting USS scope for the web service endpoint (used for OAuth
      access),
    operation_endpoint: the submitting USS endpoint where all flights in these
      cells can be retrieved from (variables {zoom}, {x}, and {y} can be used in
      the endpoint, and will be replaced with the actual grid values),
    operation_format: The output format for the USS web service endpoint (i.e.
      NASA, GUTMA),
    minimum_operation_timestamp: the lower time bound of all of the USSs flights
      in these grid cells.
    maximum_operation_timestamp: the upper time bound of all of the USSs flights
      in these grid cells.

  Returns:
    200 and a new composite token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cells metadata submit instantiated for %s at %sz, %s...',
           uss_id, zoom, str(tiles))
  sync_token = _GetRequestParameter('sync_token', None)
  if not sync_token and 'sync_token' in request.headers:
    sync_token = request.headers['sync_token']
  scope = _GetRequestParameter('scope', None)
  operation_endpoint = _GetRequestParameter('operation_endpoint', '')
  operation_format = _GetRequestParameter('operation_format', '')
  minimum_operation_timestamp = _GetRequestParameter(
    'minimum_operation_timestamp', None)
  maximum_operation_timestamp = _GetRequestParameter(
    'maximum_operation_timestamp', None)
  public_portal_endpoint = _GetRequestParameter('public_portal_endpoint', '')
  flight_info_endpoint = _GetRequestParameter('flight_info_endpoint', '')
  errorfield = errormsg = None
  if operation_endpoint and not sync_token:
    errorfield = 'sync_token'
  elif not uss_id:
    errorfield = 'uss_id'
    errormsg = 'USS identifier not received from OAuth token check.'
  elif not scope:
    errorfield = 'scope'
  elif operation_endpoint and not operation_format:
    errorfield = 'operation_format'
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
    result = wrapper.set_multi(zoom, tiles, sync_token, uss_id, scope,
                               operation_format, operation_endpoint,
                               minimum_operation_timestamp,
                               maximum_operation_timestamp,
                               public_portal_endpoint,
                               flight_info_endpoint)
  return result

def _DeleteGridCellsMetaData(zoom, tiles, uss_id):
  """Removes the USS entry in multiple GridCells.

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to delete the uss from
    uss_id: the plain text identifier for the USS from OAuth
  Returns:
    200 and a new sync_token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cells metadata delete instantiated for %s, %sz, %s...',
           uss_id, zoom, str(tiles))
  if uss_id:
    result = wrapper.delete_multi(zoom, tiles, uss_id)
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


def _ConvertRequestToTiles(zoom):
  """Converts an CSV of coords into slippy tile format at the specified zoom
      and the specified coordinate type (path, polygon, point) """
  tiles = []
  coords = _GetRequestParameter('coords', '')
  coord_type = _GetRequestParameter('coord_type', 'point').lower()
  log.debug('Retrieved coords from web params and split to %s...', coords)
  coordinates = slippy_util.convert_csv_to_coordinates(coords)
  if not coordinates:
    log.error('Invalid coords %s, must be a CSV of lat,lon...', coords)
    raise ValueError('Invalid coords, must be a CSV of lat,lon,lat,lon...')
  if coord_type == 'point':
    for c in coordinates:
      tiles.append((slippy_util.convert_point_to_tile(zoom, c[0], c[1])))
  elif coord_type == 'path':
    tiles = slippy_util.convert_path_to_tiles(zoom, coordinates)
  elif coord_type == 'polygon':
    tiles = slippy_util.convert_polygon_to_tiles(zoom, coordinates)
  else:
    raise ValueError('Invalid coord_type, must be point/path/polygon')
  return tiles


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
  else:
    log.error('Request is in an unknown format: %s', str(request))
    r = default
  return r


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


def ParseOptions(argv):
  """Parses desired options from the command line.

  Uses the command line parameters as argv, which can be altered as needed for
  testing.

  Args:
    argv: Command line parameters
  Returns:
    Options structure
  """
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
      default=int(os.getenv('INTERUSS_API_PORT', '5000')),
      help='Specific port to use on this machine for the web services '
      '[or env variable INTERUSS_API_PORT]',
      metavar='PORT')
  parser.add_option(
      '-v',
      '--verbose',
      action='store_true',
      dest='verbose',
      default=(os.environ.get('INTERUSS_VERBOSE', 'false').lower() == 'true'),
      help='Verbose (DEBUG) logging [or env variable INTERUSS_VERBOSE]')
  parser.add_option(
      '-t',
      '--testid',
      dest='testid',
      default=os.environ.get('INTERUSS_TESTID'),
      help='Force testing mode with test data located in specific test id  '
      '[or env variable INTERUSS_TESTID]',
      metavar='TESTID')
  parser.add_option(
      '-k',
      '--public_key',
      dest='public_key',
      default=os.environ.get('INTERUSS_PUBLIC_KEY'),
      help='Public key of global authorization authority [or env variable '
      'INTERUSS_PUBLIC_KEY]',
      metavar='RSAKEY')
  parser.add_option(
      '-a',
      '--auth_config',
      dest='auth_config',
      default=os.environ.get('INTERUSS_AUTH_CONFIG'),
      help='JSON describing authorization configuration, or path to JSON '
      'resource [or env variable INTERUSS_AUTH_CONFIG]',
      metavar='CONFIG')
  (options, args) = parser.parse_args(argv)
  del args
  return options


def InitializeConnection(options):
  """Initializes the wrapper and the connection to the zookeeper servers.

  The side effects of this method are to set the global variable 'wrapper' and
  call authorization.set_test_id if appropriate.

  Args:
    options: Options structure with a field per option.
  """
  global wrapper, auth

  if not options.public_key and not options.auth_config:
    log.error('Public key or auth config must be provided.')
    sys.exit(-1)

  if options.verbose:
    log.setLevel(logging.DEBUG)
  log.debug('Initializing USS metadata manager...')
  wrapper = storage_interface.USSMetadataManager(options.connectionstring)
  if options.verbose:
    wrapper.set_verbose()
  auth = authorization.Authorizer(options.public_key, options.auth_config)
  if options.testid:
    auth.SetTestId(options.testid)
    wrapper.set_testmode(options.testid)
    wrapper.delete_testdata(options.testid)


def TerminateConnection():
  global wrapper
  wrapper = None


@webapp.before_first_request
def BeforeFirstRequest():
  if wrapper is None:
    InitializeConnection(ParseOptions([]))


def main(argv):
  log.debug(
      """Instantiated application, parsing commandline
    %s and initializing connection...""", str(argv))
  options = ParseOptions(argv)
  InitializeConnection(options)
  log.info('Starting webserver...')
  webapp.run(host=options.server, port=int(options.port))


# this is what starts everything when run directly as an executable
if __name__ == '__main__':
  main(sys.argv)
