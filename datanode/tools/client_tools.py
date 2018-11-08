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

import json
import os

import requests

try:
  from urllib.request import Request, urlopen, HTTPError  # Python 3
except ImportError:
  from urllib2 import Request, urlopen, HTTPError  # Python 2

DEFAULT_AUTH_URL = 'https://utmalpha.arc.nasa.gov/fimsAuthServer/oauth/token?grant_type=client_credentials'
DEFAULT_HOST = 'http://18.216.177.215:8120/'
DEFAULT_REQUEST_PATH = 'GridCellsOperator/10?coords=48.3379,-103.4582,47.6191,-103.1149,47.5672,-102.4530,48.3525,-102.4722,48.3379,-103.4582&coord_type=polygon'

def get_token(auth_key, auth_url):
  """Call the specified OAuth server to retrieve an access token.

  Args:
    auth_key: Base64-encoded username and password.
    auth_url: URL the provides an access token.

  Returns:
    Access token for TCL4 InterUSS Platform data node.

  Raises:
    ValueError: When access_token was not returned properly.
  """
  r = requests.post(auth_url, headers={'Authorization': 'Basic ' + auth_key})
  result = r.content
  result_json = json.loads(result)
  if 'access_token' in result_json:
    return result_json['access_token']
  else:
    raise ValueError('Error getting token: ' + r.content)

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
  return r.content

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
