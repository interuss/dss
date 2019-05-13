"""Utilities for interacting with TCL4 InterUSS Platform Data Nodes.


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
import datetime
import json
import os
import sys

import requests
import requests.exceptions

try:
  from urllib.request import Request, urlopen, HTTPError  # Python 3
except ImportError:
  from urllib2 import Request, urlopen, HTTPError  # Python 2

DEFAULT_AUTH_URL = 'https://utmalpha.arc.nasa.gov/fimsAuthServer/oauth/token?grant_type=client_credentials'
DEFAULT_HOST = 'https://node4.tcl4.interussplatform.com:8121/'
DEFAULT_REQUEST_PATH = 'GridCellsOperator/10?coords=48.832,-101.832,47.954,-101.832,47.954,-100.501,48.832,-100.501,48.832,-101.832&coord_type=polygon'
EXPIRATION_BUFFER = 5  # seconds

# Access token scope for accessing the InterUSS Platform.
INTERUSS_SCOPE = 'utm.nasa.gov_write.conflictmanagement'

# Access token scope for access individual USSs' operations.
USS_SCOPE = 'utm.nasa.gov_write.operation'


def get_metadata(token, url):
  """Retrieve metadata from TCL4 InterUSS Platform data node.

  Args:
    token: Access token acquired from OAuth server.
    url: GridCellOperators endpoint of data node.

  Returns:
    USS Metadata JSON string returned from GridCellOperators endpoint.

  Raises:
    ValueError: When result does not contain valid JSON.
  """
  r = requests.get(url, headers={
    'Cache-Control': 'no-cache', 'access_token': token})
  try:
    json.loads(r.content)
  except ValueError:
    raise ValueError('Error getting metadata: ' + r.content)
  except HTTPError as e:
    raise ValueError('HTTPError getting metadata: ' + str(e))
  except requests.exceptions.ConnectionError as e:
    raise ValueError('ConnectionError getting metadata: ' + str(e))
  return r.content


def get_operation(token, uss_baseurl, gufi):
  """Retrieve details of a single operation from a USS."""
  url = os.path.join(uss_baseurl, 'operations', gufi)
  try:
    response = requests.get(url, headers={
      'Cache-Control': 'no-cache', 'Authorization': 'Bearer ' + token})
    return response.content, response.status_code
  except HTTPError as e:
    return 'HTTPError: ' + str(e), 500
  except ValueError as e:
    return 'ValueError: ' + str(e), 400
  except requests.exceptions.ConnectionError as e:
    return 'Client ConnectionError: ' + str(e), 600


def add_auth_arguments(parser):
  """Add arguments relating to OAuth authentication.

  Args:
    parser: argparser to mutate with additional arguments.
  """
  parser.add_argument(
    '-k',
    '--auth_key',
    dest='auth_key',
    default=os.environ.get('AUTH_KEY', None),
    help='Base64-encoded username and password to pass to the OAuth server as '
         'Basic XXX in the Authentication header. Defaults to AUTH_KEY '
         'environment variable if defined',
    metavar='AUTH_KEY')
  parser.add_argument(
    '-a',
    '--auth_url',
    dest='auth_url',
    default=DEFAULT_AUTH_URL,
    help='URL from which to obtain an access token',
    metavar='URL')


def add_node_arguments(parser):
  """Add arguments relating to TCL4 InterUSS Platform data node access.

  Args:
    parser: argparser to mutate with additional arguments.
  """
  parser.add_argument(
    '-n',
    '--node',
    dest='node',
    default=DEFAULT_HOST,
    help='Host name of TCL4 InterUSS Platform node to poll',
    metavar='HOST')
  parser.add_argument(
    '-r',
    '--request_path',
    dest='request_path',
    default=DEFAULT_REQUEST_PATH,
    help='Path and query to use on the TCL4 InterUSS Platform node',
    metavar='REQUESTPATH')


def make_node_url(options):
  """Combine option values from node arguments into a URL for that node.

  Args:
    options: Parsed options from argparser which included arguments from
      add_node_arguments.

  Returns:
    URL to TCL4 InterUSS Platform data node endpoint.
  """
  return (options.node + ('' if options.node.endswith('/') else '/') +
          options.request_path)



CachedToken = collections.namedtuple('CachedToken', ('value', 'expiration'))


class TokenManager(object):
  """Transparently provides access tokens, from cache when possible."""

  def __init__(self, auth_url, auth_key):
    """Create a TokenManager.

    Args:
      auth_url: URL the provides an access token.
      auth_key: Base64-encoded username and password.
    """
    if '&scope=' in auth_url:
      print('USAGE ERROR: The auth URL should now be provided without a scope '
            'specified in its GET parameters.')
      sys.exit(1)

    self._auth_url = auth_url
    self._auth_key = auth_key
    self._tokens = {}

  def _retrieve_token(self, scope):
    """Call the specified OAuth server to retrieve an access token.

    Args:
      scope: Access token scope to request.

    Returns:
      CachedToken with requested scope.

    Raises:
      ValueError: When access_token was not returned properly.
    """
    url = self._auth_url + '&scope=' + scope
    r = requests.post(url, headers={'Authorization': 'Basic ' + self._auth_key})
    result = r.content
    result_json = json.loads(result)
    if 'access_token' in result_json:
      token = result_json['access_token']
      expires_in = int(result_json.get('expires_in', 0))
      expiration = (datetime.datetime.utcnow() +
                    datetime.timedelta(seconds=expires_in - EXPIRATION_BUFFER))
      return CachedToken(token, expiration)
    else:
      raise ValueError('Error getting token: ' + r.content)

  def get_token(self, scope):
    """Retrieve a current access token with the requested scope.

    Args:
      scope: Access token scope to request.

    Returns:
      Access token content.

    Raises:
      ValueError: When access_token was not returned properly.
    """
    if scope in self._tokens:
      if self._tokens[scope].expiration > datetime.datetime.utcnow():
        return self._tokens[scope].value

    print('')
    print('Getting access token for %s...' % scope)
    self._tokens[scope] = self._retrieve_token(scope)
    return self._tokens[scope].value
