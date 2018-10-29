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
# Tools for slippy conversion
import slippy_util

import uvrs

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
# VERSION = 'TCL4.0.1'  # Updated slippy format, added SSL
# VERSION = 'TCL4.0.1.002'  # Updated public key format for env variables
# VERSION = 'TCL4.0.1.003'  # Updated docker startup script
# VERSION = 'TCL4.0.1.004'  # Fixed slippy validation error
# VERSION = 'TCL4.0.1.005'  # Added Authorization as a valid JWT field
# VERSION = 'TCL4.0.1.006'  # Corrected error codes and text
# VERSION = 'TCL4.0.1.007'  # Fixed non-tz aware dates and version in operations
VERSION = 'TCL4.0.2.008'  # sync with master branch for multi-grid and docker updates


TESTID = None
DEFAULT_SCOPE = 'utm.nasa.gov_write.conflictmanagement'

UVR_WRITE_SCOPE = 'utm.nasa.gov_write.constraint'
UVR_TILE_LIMIT = 500  # Number of tiles to allow for multi get/put/del for UVRs

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_StorageAPI')
log.setLevel(logging.DEBUG)
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
  uss_id = _ValidateAccessToken()
  result = {}
  try:
    zoom = int(zoom)
    x = int(x)
    y = int(y)
  except ValueError:
    abort(status.HTTP_400_BAD_REQUEST,
          'Invalid parameters for slippy tile coordinates, must be integers.')
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


@webapp.route(
  '/GridCellsOperator/<zoom>',
  methods=['GET', 'PUT', 'POST', 'DELETE'])
def GridCellsOperatorDataHandler(zoom):
  """Handles the web service request for multi-grid operations.

  Args:
    zoom: zoom level in slippy tile format
    OAuth access_token as part of the header
    Plus posted webargs:
      coords: csv of lon,lat,long,lat,etc.
      coord_type: (optional) type of coords - point (default), path, polygon
      and additional as needed for PUT/POST and DELETE methods (see below)
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  uss_id = _ValidateAccessToken()
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
  if request.method == 'GET':
    result = _GetGridCellsOperator(zoom, tiles)
  elif request.method in ('PUT', 'POST'):
    result = _PutGridCellsOperator(zoom, tiles, uss_id)
  elif request.method == 'DELETE':
    result = _DeleteGridCellsOperator(zoom, tiles, uss_id)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)


@webapp.route(
  '/GridCellsOperation/<zoom>/<gufi>',
  methods=['PUT', 'POST', 'DELETE'])
def GridCellsOperationDataHandler(zoom, gufi):
  """Handles the web service request for multi-grid operations.

  Args:
    zoom: zoom level in slippy tile format
    gufi: flight identifier to remove
    OAuth access_token as part of the header
    Plus posted webargs:
      coords: csv of lon,lat,long,lat,etc.
      coord_type: (optional) type of coords - point (default), path, polygon
      and additional as needed for PUT/POST and DELETE methods (see below)
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  uss_id = _ValidateAccessToken()
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
  if request.method in ('PUT', 'POST'):
    result = _PutGridCellsOperation(zoom, tiles, uss_id, gufi)
  elif request.method == 'DELETE':
    result = _DeleteGridCellsOperation(zoom, tiles, uss_id, gufi)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)


@webapp.route(
  '/UVR/<zoom>/<message_id>',
  methods=['PUT', 'POST', 'DELETE'])
def UvrDataHandler(zoom, message_id):
  """Handles the web service request for emplacing or removing a UVR.

  Args:
    zoom: zoom level in slippy tile format
    message_id: unique message ID of constraint
    OAuth access_token as part of the header
  Returns:
    200 with token and metadata in JSON format,
    or the nominal 4xx error codes as necessary.
  """
  uss_id = _ValidateAccessToken(UVR_WRITE_SCOPE)

  if not uss_id:
    return {
      'status': 'fail',
      'code': status.HTTP_400_BAD_REQUEST,
      'message': 'uss_id must be provided in the request to modify a UVR.'}

  uvr_json = _GetRequestParameter('uvr', None)
  if not uvr_json:
    return {
      'status': 'error',
      'code': status.HTTP_400_BAD_REQUEST,
      'message': 'uvr field must be provided in the form data request'}

  result = {}
  try:
    unvalidated_uvr = json.loads(uvr_json)
  except ValueError as e:
    abort(status.HTTP_400_BAD_REQUEST, 'Error parsing UVR JSON: ' + e.message)

  try:
    uvr = uvrs.Uvr(unvalidated_uvr)

    if uvr['message_id'] != message_id:
      raise ValueError('message_id "%s" in request does not match message_id '
                       '"%s" in UVR' % (message_id, uvr['message_id']))
    if uvr['originator_id'] != uss_id:
      raise ValueError('authenticated client ID "%s" does not match '
                       'originator_id "%s" in UVR' % (uss_id,
                                                      uvr['originator_id']))

    zoom = int(zoom)
    tiles = uvr.get_tiles(zoom)
    if len(tiles) > UVR_TILE_LIMIT:
      raise OverflowError('Limit of %d tiles impacted exceeded (%d)'
                          % (UVR_TILE_LIMIT, len(tiles)))
  except (KeyError, ValueError, TypeError) as e:
    abort(status.HTTP_400_BAD_REQUEST, e.message)
  except OverflowError as e:
    abort(status.HTTP_413_REQUEST_ENTITY_TOO_LARGE, e.message)
  if request.method in ('PUT', 'POST'):
    result = wrapper.insert_uvr(zoom, tiles, uvr)
  elif request.method == 'DELETE':
    result = wrapper.delete_uvr(zoom, tiles, uvr)
  else:
    abort(status.HTTP_405_METHOD_NOT_ALLOWED, 'Request method not supported.')
  return _FormatResult(result)

######################################################################
################       INTERNAL FUNCTIONS    #########################
######################################################################
def _ValidateAccessToken(required_scope=None):
  """Checks the access token, aborting if it does not pass.

  Uses an Oauth public key to validate an access token.

  Args:
    required_scope: If specified, an explicit scope that must be in the list of
      granted scopes for authorization to succeed. Defaults to DEFAULT_SCOPE.

  Returns:
    USS identification from OAuth client_id field
  """
  if not required_scope:
    required_scope = DEFAULT_SCOPE
  uss_id = None
  if ('access_token' in request.headers and TESTID and
    TESTID in request.headers['access_token']) :
    uss_id = request.headers['access_token']
  elif ('Authorization' in request.headers and TESTID and
        TESTID in request.headers['Authorization']):
    uss_id = request.headers['Authorization']
  elif (TESTID and 'access_token' not in request.headers and
        'Authorization' not in request.headers):
    uss_id = TESTID
  else:
    # TODO(hikevin): Replace with OAuth Discovery and JKWS
    secret = os.getenv('INTERUSS_PUBLIC_KEY')
    token = None
    if 'Authorization' in request.headers:
      token = request.headers['Authorization'].replace('Bearer ', '')
    elif 'access_token' in request.headers:
      token = request.headers['access_token']
    if secret and token:
      # ENV variables sometimes don't pass newlines, spec says white space
      # doesn't matter, but pyjwt cares about it, so fix it
      secret = secret.replace(' PUBLIC ', '_PLACEHOLDER_')
      secret = secret.replace(' ', '\n')
      secret = secret.replace('_PLACEHOLDER_', ' PUBLIC ')
      try:
        r = jwt.decode(token, secret, algorithms='RS256')
        if required_scope in r['scope']:
          uss_id = r['sub']
        else:
          log.error('%s not in scope %s', required_scope, r['scope'])
          abort(status.HTTP_401_UNAUTHORIZED,
                'OAuth access_token is invalid: '
                'scope not valid for conflict management.')
      except jwt.ExpiredSignatureError:
        log.error('Access token has expired.')
        abort(status.HTTP_401_UNAUTHORIZED,
              'OAuth access_token is invalid: token has expired.')
      except jwt.DecodeError:
        log.error('Access token is invalid and cannot be decoded.')
        abort(status.HTTP_400_BAD_REQUEST,
              'OAuth access_token is invalid: token cannot be decoded.')
      except jwt.ImmatureSignatureError:
        log.error('Access token nbf claim represents a time in the future.')
        abort(status.HTTP_400_BAD_REQUEST,
              'OAuth access_token is invalid: nbf claim represents a time in the future.')
      except:
        log.error('Unknown error processing JWT.')
        abort(status.HTTP_400_BAD_REQUEST,
              'OAuth access_token is invalid: unknown error.')
    else:
      log.error('Attempt to access resource without access_token in header.')
      abort(status.HTTP_403_FORBIDDEN,
            'Valid OAuth access_token must be provided in header.')
  return uss_id


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
    behavior. Operation fails if the metadata has been updated since
    GET GridCellOperator was originally called (based on token).

  Args:
    zoom: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
    uss_id: the plain text identifier for the USS from OAuth
  Plus posted webargs:
    sync_token: the token retrieved in the original GET GridCellOperator,
    announce: Boolean announcement request from other USSs,
    minimum_operation_timestamp: the lower time bound of all of the USSs flights
      in this grid cell.
    maximum_operation_timestamp: the upper time bound of all of the USSs flights
      in this grid cell.
    operations: a list of operations in this cell,

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
  Plus posted webargs:
    operation_signature: encoded signatrue for this version of the operation
    effective_time_begin: tile the operation begins
    effective_time_end: time the operation ends
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
  gufi = _GetRequestParameter('gufi', gufi)
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


def _GetGridCellsOperator(zoom, tiles):
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


def _PutGridCellsOperator(zoom, tiles, uss_id):
  """Updates the metadata stored in multiple GridCells.

    Updates the metadata stored in a multiple GridCell using optimistic locking
    behavior. Operation fails if the metadata has been updated since
    GET GridCellsMetadata was originally called (based on sync_token).

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to retrieve
    uss_id: the plain text identifier for the USS from OAuth
  Plus posted webargs:
    sync_token: the token retrieved in the original GET GridCellOperator,
    announce: Boolean announcement request from other USSs,
    minimum_operation_timestamp: the lower time bound of all of the USSs flights
      in this grid cell.
    maximum_operation_timestamp: the upper time bound of all of the USSs flights
      in this grid cell.
    operations: a list of operations in this cell,

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
    result = wrapper.set_multi(zoom, tiles, sync_token, uss_id, baseurl,
                               announce, minimum_operation_timestamp,
                               maximum_operation_timestamp, operations)
  return result


def _DeleteGridCellsOperator(zoom, tiles, uss_id):
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


def _PutGridCellsOperation(zoom, tiles, uss_id, gufi):
  """Updates the metadata for an operation stored in multiple GridCells.

    Updates the metadata stored in a multiple GridCell using optimistic locking
    behavior. Operation fails if the metadata has been updated since
    GET GridCellsMetadata was originally called (based on sync_token).

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to retrieve
    uss_id: the plain text identifier for the USS from OAuth
    gufi: flight identifier to write
  Plus posted webargs:
    operation_signature: encoded signatrue for this version of the operation
    effective_time_begin: tile the operation begins
    effective_time_end: time the operation ends

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
  gufi = _GetRequestParameter('gufi', gufi)
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
    result = wrapper.set_multi_operation(zoom, tiles, sync_token, uss_id, gufi,
                                         signature, begin, end)
  return result


def _DeleteGridCellsOperation(zoom, tiles, uss_id, gufi):
  """Removes the USS entry in multiple GridCells.

  Args:
    zoom: zoom level in slippy tile format
    tiles: array of x,y tiles to delete the uss from
    uss_id: the plain text identifier for the USS from OAuth
    gufi: flight identifier to remove
  Returns:
    200 and a new sync_token if updated successfully,
    409 if there is a locking conflict that could not be resolved, or
    the other nominal 4xx error codes as necessary.
  """
  log.info('Grid cells delete operation instantiated for %s, %sz, %s...',
           uss_id, zoom, str(tiles))
  if uss_id:
    result = wrapper.delete_multi_operation(zoom, tiles, uss_id, gufi)
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
  elif request.data:
    rjson = json.loads(request.data)
    r = default if name not in rjson else rjson[name]
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


def _VerifyPublicKey():
  if not os.environ.get('INTERUSS_PUBLIC_KEY'):
    log.error('INTERUSS_PUBLIC_KEY environment variable must be set.')
    sys.exit(-1)


def ParseOptions(argv):
  """Parses desired options from the command line.

  Uses the command line parameters as argv, which can be altered as needed for
  testing.

  Args:
    argv: Command line parameters
  Returns:
    Options structure
  """
  global TESTID
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
  (options, args) = parser.parse_args(argv)
  del args
  return options


def InitializeConnection(options=None):
  """Initializes the wrapper and the connection to the zookeeper servers.

  The side effects of this method are to set the global variables 'wrapper' and
  'TESTID'.

  Args:
    options: Options structure with a field per option.
  """
  global wrapper, TESTID
  if (options and options.verbose) or os.environ.get('INTERUSS_VERBOSE'):
    log.setLevel(logging.DEBUG)
  log.debug('Initializing USS metadata manager...')
  wrapper = storage_interface.USSMetadataManager(
      os.getenv('INTERUSS_CONNECTIONSTRING',
                options.connectionstring if options else None))
  if (options and options.verbose) or os.environ.get('INTERUSS_VERBOSE'):
    wrapper.set_verbose()
  if (options and options.testid) or os.environ.get('INTERUSS_TESTID'):
    TESTID = os.getenv('INTERUSS_TESTID', options.testid if options else None)
    log.info('Authorization set to test mode with TESTID=%s' % TESTID)
    wrapper.set_testmode(TESTID)
    wrapper.delete_testdata(TESTID)


def TerminateConnection():
  global wrapper
  wrapper = None


@webapp.before_first_request
def BeforeFirstRequest():
  if wrapper is None:
    _VerifyPublicKey()
    InitializeConnection()


def main(argv):
  _VerifyPublicKey()
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
