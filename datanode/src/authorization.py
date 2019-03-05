"""The InterUSS Platform Data Node authorization tools.

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

import logging
import os
import sys
from flask import abort
from flask import Flask
from flask import jsonify
from flask import request
import jwt
from rest_framework import status

TESTID = None

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_Authorization')


def SetTestId(testid):
  global TESTID
  TESTID = testid
  log.info('Authorization set to test mode with TESTID=%s' % TESTID)


def ValidateAccessToken(zoom=None, tiles=None):
  """Checks the access token, aborting if it does not pass.

  Uses an Oauth public key to validate an access token.

  Args:
    zoom: slippy zoom level user is attempting to access
    tiles: collection of (x,y) slippy tiles user is attempting to access

  Returns:
    USS identification from OAuth client_id field

  Raises:
    HTTPException: when the access token is invalid or inappropriate
  """
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
        #TODO(hikevin): Check scope is valid for InterUSS Platform
        uss_id = r['client_id'] if 'client_id' in r else r['sub']
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
  return uss_id


def _VerifyPublicKey():
  if not os.environ.get('INTERUSS_PUBLIC_KEY'):
    log.error('INTERUSS_PUBLIC_KEY environment variable must be set.')
    sys.exit(-1)


# Verify that the public key was provided upon loading this library.
_VerifyPublicKey()
