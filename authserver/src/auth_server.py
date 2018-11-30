"""The InterUSS Platform authorization server.

This module is an example of a server that can provide access tokens to access
an InterUSS Platform data node.


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
import csv
import datetime
import hashlib
import logging
from optparse import OptionParser
import os
import sys
import uuid
from flask import abort
from flask import Flask
from flask import jsonify
from flask import request
import jwt
from rest_framework import status

VERSION = '0.1.0.000'  # Initial draft/demo release

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_AuthServer')
webapp = Flask(__name__)  # Global object serving the authorization service


config = None
Configuration = collections.namedtuple(
    'Configuration', 'public_key private_key users issuer expiration')


_UNIX_EPOCH = datetime.datetime.utcfromtimestamp(0)
_SALT = 'InterUSS Platform'


# Environment variable names for various parameters:
_KEY_SERVER = 'INTERUSS_AUTH_SERVER'
_KEY_PORT = 'INTERUSS_AUTH_PORT'
_KEY_VERBOSE = 'INTERUSS_AUTH_VERBOSE'
_KEY_PUBLIC_KEY = 'INTERUSS_AUTH_PUBLIC_KEY'
_KEY_PRIVATE_KEY = 'INTERUSS_AUTH_PRIVATE_KEY'
_KEY_ROSTER = 'INTERUSS_AUTH_ROSTER'
_KEY_ISSUER = 'INTERUSS_AUTH_ISSUER'
_KEY_EXPIRATION = 'INTERUSS_AUTH_EXPIRATION'


@webapp.route('/', methods=['GET'])
@webapp.route('/status', methods=['GET'])
def Status():
  # Just a quick status checker.
  log.debug('Status handler instantiated...')
  return jsonify({
    'status': 'success',
    'message': 'OK',
    'version': VERSION
  })


@webapp.route('/key', methods=['GET'])
def WebKey():
  # Returns the public key that may be used to validate tokens from this server.
  log.debug('Key handler instantiated...')
  return config.public_key


@webapp.route(
  '/oauth/token',
  methods=['POST'])
def TokenHandler():
  """Handles the web service request for an access token.

  Tokens can be manually validated at https://jwt.io

  Args:
    Basic Authorization header.
    Query argument of grant_type must be "client_credentials".
  Returns:
    200 with token and associated data in JSON format, or the nominal error
    codes as necessary.
  """
  log.debug('Token handler instantiated...')

  # Verify grant_type
  if 'grant_type' not in request.args and 'grant_type' not in request.form:
    abort(status.HTTP_400_BAD_REQUEST, jsonify({
      "error": "invalid_request",
      "error_description": "Missing grant type"
    }))

  grant_type = (request.args['grant_type'] if 'grant_type' in request.args
                else request.form['grant_type'])
  if grant_type != 'client_credentials':
    abort(status.HTTP_400_BAD_REQUEST, jsonify({
      "error": "unsupported_grant_type",
      "error_description": "Unsupported grant type: " + grant_type
    }))

  # Authorize user
  auth = request.authorization
  if not auth:
    abort(status.HTTP_401_UNAUTHORIZED, jsonify({
      "error": "authorization_header_missing",
      "error_description": 'Authorization header missing'
    }))

  if auth.username not in config.users:
    abort(status.HTTP_401_UNAUTHORIZED, jsonify({
      "error": "invalid_login",
      "error_description": 'Username or password not recognized'
    }))
  user = config.users[auth.username]

  password_hash = hashlib.sha256(
    _SALT + ' ' + auth.username + ' ' + auth.password).hexdigest()
  if password_hash != user.hashed_password:
    abort(status.HTTP_401_UNAUTHORIZED, jsonify({
      "error": "invalid_login",
      "error_description": 'Username or password not recognized'
    }))

  # Create access_token
  timestamp = int((datetime.datetime.utcnow() - _UNIX_EPOCH).total_seconds())
  jti = str(uuid.uuid4())

  payload = {
    'sub': auth.username,
    'nbf': timestamp - 1,
    'scope': user.scopes,
    'iss': config.issuer,
    'exp': timestamp + config.expiration,
    'jti': jti,
    'client_id': auth.username,
  }

  access_token = jwt.encode(payload, key=config.private_key, algorithm='RS256')

  # Make sure the token validates correctly
  try:
    jwt.decode(access_token, config.public_key, algorithms='RS256')
  except jwt.ExpiredSignatureError as e:
    log.error('Generated access token has already expired: ' + str(e))
    abort(status.HTTP_500_INTERNAL_SERVER_ERROR,
          'Encountered ExpiredSignatureError when validating generated token')
  except jwt.DecodeError as e:
    log.error('Generated access token could not be decoded: ' + str(e))
    abort(status.HTTP_500_INTERNAL_SERVER_ERROR,
          'Generated access token could not be decoded for validation')

  return jsonify({
    'access_token': access_token,
    'token_type': "bearer",
    'expires_in': config.expiration - 1,
    'scope': ' '.join(payload['scope']),
    'sub': payload['sub'],
    'nbf': payload['nbf'],
    'iss': payload['iss'],
    'jti': payload['jti'],
  })


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
    '-s',
    '--server',
    dest='server',
    default=os.getenv(_KEY_SERVER, '127.0.0.1'),
    help='Specific server name to use on this machine for the web services '
         '(only applies directly from command line with Flask development '
         'server) [or use env variable %s]' % _KEY_SERVER,
    metavar='SERVER')
  parser.add_option(
    '-p',
    '--port',
    dest='port',
    default=os.getenv(_KEY_PORT, '80'),
    help='Specific port to use on this machine for the web services '
         '[or use env variable %s]' % _KEY_PORT,
    metavar='PORT')
  parser.add_option(
    '-v',
    '--verbose',
    action='store_true',
    dest='verbose',
    default=os.getenv(_KEY_VERBOSE, False),
    help='Verbose (DEBUG) logging [or env variable %s]' % _KEY_VERBOSE)
  parser.add_option(
    '-d',
    '--publickey',
    dest='publickey',
    default=os.getenv(_KEY_PUBLIC_KEY, None),
    help='Name of file containing public key for JWT verification. Given an '
         'existing private key, generate a public key with '
         '`openssl rsa -in private.pem -outform PEM -pubout -out public.pem` '
         '[or use env variable %s]' % _KEY_PUBLIC_KEY,
    metavar='FILENAME')
  parser.add_option(
    '-k',
    '--privatekey',
    dest='privatekey',
    default=os.getenv(_KEY_PRIVATE_KEY, None),
    help='Name of file containing private key for JWT generation. Generate one '
         'with `openssl genrsa -out private.pem 2048` '
         '[or use env variable %s]' % _KEY_PRIVATE_KEY,
    metavar='FILENAME')
  parser.add_option(
    '-r',
    '--roster',
    dest='roster',
    default=os.getenv(_KEY_ROSTER, None),
    help='Name of file containing lines of\nUSERNAME,HASHED PASSWORD,SCOPES\n'
         'where USERNAME is the domain name of the user, HASHED PASSWORD is '
         '(the SHA256 hash of "InterUSS Platform USERNAME PASSWORD" (so, for '
         'instance, if wing.com\'s password were "wing", its HASHED PASSWORD '
         'would be SHA256(InterUSS Platform wing.com wing) which begins '
         'b03ed6. Generate this hash at a Linux command line with `echo -n '
         '"InterUSS Platform wing.com wing" | openssl dgst -sha256` Scopes '
         'should be separated by spaces [or use env variable %s]' % _KEY_ROSTER,
    metavar='FILENAME')
  parser.add_option(
    '-i',
    '--issuer',
    dest='issuer',
    default=os.getenv(_KEY_ISSUER, '127.0.0.1'),
    help='What to populate in the JWT iss (issuer) field '
         '[or env variable %s]' % _KEY_ISSUER,
    metavar='ISSUER')
  parser.add_option(
    '-e',
    '--expiration',
    dest='expiration',
    default=os.getenv(_KEY_EXPIRATION, '3600'),
    help='Number of seconds each access_token should live for '
         '[or env variable %s]' % _KEY_EXPIRATION,
    metavar='SECONDS')
  (options, args) = parser.parse_args(argv)
  del args
  return options


def _LoadConfiguration(options=None):
  """Initializes the configuration of this server.

  The side effect of this method is to set the global variable 'config'.

  Args:
    options: Options structure with a field per option.
  """
  global config
  if (options and options.verbose) or os.environ.get(_KEY_VERBOSE):
    log.setLevel(logging.DEBUG)
  log.debug('Initializing InterUSS auth server...')

  filename = (options.publickey if options else
              os.environ.get(_KEY_PUBLIC_KEY, None))
  if filename:
    with open(filename, 'r') as f:
      public_key = f.read()
  else:
    public_key = None

  filename = options.privatekey if options else os.environ[_KEY_PRIVATE_KEY]
  with open(filename, 'r') as f:
    private_key = f.read()

  filename = options.roster if options else os.environ[_KEY_ROSTER]
  users = {}
  User = collections.namedtuple('User', 'hashed_password scopes')
  with open(filename, 'r') as f:
    reader = csv.reader(f)
    for line in reader:
      if len(line) == 3:
        users[line[0]] = User(line[1], line[2].split(' '))

  issuer = (options.issuer if options else
            os.environ.get(_KEY_ISSUER, '127.0.0.1'))

  expiration = int(options.expiration if options else
                   os.environ.get(_KEY_EXPIRATION, '3600'))

  config = Configuration(public_key, private_key, users, issuer, expiration)


@webapp.before_first_request
def BeforeFirstRequest():
  if config is None:
    # This is triggered when using a separate WSGI server (e.g., Gunicorn).
    _LoadConfiguration()


def main(argv):
  # This is only triggered when run as a standalone app using Flask's debug
  # WSGI server.
  log.debug(
    """Instantiated application, parsing commandline
  %s and initializing connection...""", str(argv))
  options = ParseOptions(argv)
  _LoadConfiguration(options)
  log.info('Starting webserver...')
  webapp.run(host=options.server, port=int(options.port))


# This is what starts everything when run directly as an executable
if __name__ == '__main__':
  main(sys.argv)
