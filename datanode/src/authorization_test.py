"""Test of the InterUSS Platform Data Node authorization module.

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
import json
import os
import unittest
import tempfile

from cryptography.hazmat.primitives import serialization as crypto_serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.backends import default_backend as crypto_default_backend
import jwt
from werkzeug.exceptions import HTTPException

import authorization

KeyPair = collections.namedtuple('KeyPair', 'private_key public_key')

def MakeKeyPair():
  """Create a public/private RSA key pair.

  Returns:
    Public and private keys in KeyPair using PEM text format.
  """
  key = rsa.generate_private_key(
    backend=crypto_default_backend(),
    public_exponent=65537,
    key_size=2048
  )
  private_key = key.private_bytes(
    crypto_serialization.Encoding.PEM,
    crypto_serialization.PrivateFormat.PKCS8,
    crypto_serialization.NoEncryption())
  public_key = key.public_key().public_bytes(
    encoding=crypto_serialization.Encoding.PEM,
    format=crypto_serialization.PublicFormat.SubjectPublicKeyInfo
  )
  return KeyPair(private_key, public_key)


KEY_PAIRS = {
  'Global': MakeKeyPair(),
  'pinecreek.com': MakeKeyPair(),
  'zion.gov': MakeKeyPair()
}

CONFIG_CUSTOM_AUTH = json.dumps([
  {
    'public_key': KEY_PAIRS['Global'].public_key,
    'constraints': [
      {
        'type': 'range',
        'tiles': [[14, '*', '*']]
      }
    ]
  }, {
    'name': 'Pine Creek area authority',
    'public_key': KEY_PAIRS['pinecreek.com'].public_key,
    'constraints': [
      {
        'type': 'issuer',
        'issuer': 'pinecreek.com'
      }, {
        'type': 'range',
        'tiles': [[14, '*', '*']],
        'inclusive': True
      }, {
        'type': 'area',
        'outline': [[37.2167, -112.9584],
                    [37.2095, -112.9581],
                    [37.2091, -112.9406],
                    [37.2172, -112.9311]],
        'inside': True
      }]
  }, {
    'name': 'Zion canyon authority',
    'public_key': KEY_PAIRS['zion.gov'].public_key,
    'constraints': [
      {
        'type': 'issuer',
        'issuer': 'zion.gov'
      }, {
        'type': 'range',
        'tiles': [
          [14, 3050, '6364:6366'],
          [14, 3051, '6365:6369'],
          [14, '3048:3055', '6362:6364']
        ],
        'inclusive': True
      }, {
        'type': 'range',
        'tiles': [[14, 3049, 6364]],
        'inclusive': False
      }
    ]
  }
])


class InterUSSStorageAPITestCase(unittest.TestCase):

  def setUp(self):
    self.global_auth = authorization.Authorizer(
        KEY_PAIRS['Global'].public_key, '')

    self.custom_auth_config = tempfile.NamedTemporaryFile(delete=False)
    self.custom_auth_config.write(CONFIG_CUSTOM_AUTH)
    self.custom_auth_config.close()
    self.custom_auth = authorization.Authorizer(
        '', 'file://' + self.custom_auth_config.name)

  def tearDown(self):
    os.unlink(self.custom_auth_config.name)

  def assertFails(self, auth, headers, tiles):
    def validate():
      auth.ValidateAccessToken(headers, tiles)
    self.assertRaises(HTTPException, validate)

  def assertSucceeds(self, auth, headers, tiles):
    uss_id = auth.ValidateAccessToken(headers, tiles)
    self.assertIsNotNone(uss_id)

  def testGlobalNoToken(self):
    # Expect failure when no access token is provided
    self.assertFails(self.global_auth, {}, None)
    self.assertFails(self.global_auth, {}, ((14, 123, 456),))
    self.assertFails(self.global_auth, {}, ((14, 123, 456), (14, 124, 456)))

  def testGlobalSuccess(self):
    # Expect success with a valid access token
    token = jwt.encode(
        {'client_id': 'client'}, KEY_PAIRS['Global'].private_key, 'RS256')
    self.assertSucceeds(self.global_auth, {'access_token': token}, None)
    self.assertSucceeds(
        self.global_auth, {'access_token': token}, ((14, 123, 456),))
    self.assertSucceeds(
        self.global_auth, {'access_token': token},
        ((14, 123, 456), (14, 124, 456)))

  def testGlobalCorruptToken(self):
    # Expect failure with a corrupted access token
    token = jwt.encode(
      {'client_id': 'client'}, KEY_PAIRS['Global'].private_key, 'RS256')
    for i in range(token.find('\n') + 10, len(token)):
      if token[i] != '0':
        token = token[0:i] + '0' + token[i+1:]
        break
    self.assertFails(self.global_auth, {'access_token': token}, None)
    self.assertFails(
        self.global_auth, {'access_token': token}, ((14, 123, 456),))
    self.assertFails(
        self.global_auth, {'access_token': token},
        ((14, 123, 456), (14, 124, 456)))

  def testCustomNoToken(self):
    # Expect failure when no access token is provided
    self.assertFails(self.custom_auth, {}, None)
    self.assertFails(self.custom_auth, {}, ((14, 3051, 6365),))
    self.assertFails(self.custom_auth, {}, ((14, 3050, 6365), (14, 3051, 6365)))

  def testCustomGlobalSuccess(self):
    # Expect success authorizing with global authority
    token = jwt.encode({'client_id': 'client'},
                       KEY_PAIRS['Global'].private_key, 'RS256')
    self.assertSucceeds(self.custom_auth, {'access_token': token}, None)
    self.assertSucceeds(
      self.custom_auth, {'access_token': token}, ((14, 3051, 6365),))
    self.assertSucceeds(
      self.custom_auth, {'access_token': token}, ((14, 3050, 6365),))
    self.assertSucceeds(
      self.custom_auth, {'access_token': token},
      ((14, 3050, 6365), (14, 3051, 6365)))
    self.assertSucceeds(
      self.custom_auth, {'access_token': token},
      ((14, 3051, 6365), (14, 3052, 6365)))
    self.assertSucceeds(
      self.custom_auth, {'access_token': token},
      ((14, 3050, 6365), (14, 3051, 6365), (14, 3052, 6365)))

  def testCustomGlobalFailure(self):
    # Expect failure authorizing with global authority at wrong zoom
    token = jwt.encode({'client_id': 'client'},
                       KEY_PAIRS['Global'].private_key, 'RS256')
    self.assertFails(
      self.custom_auth, {'access_token': token},
      ((13, int(3051/2), int(6365/2)),))
    self.assertFails(
      self.custom_auth, {'access_token': token}, ((15, 3051, 6365),))
    self.assertFails(
      self.custom_auth, {'access_token': token},
      ((15, 3050*2, 6365*2), (14, 3051, 6365)))

  def testCustomAreaSuccess(self):
    # Expect success authorizing with area-based authority
    token = jwt.encode({'client_id': 'client', 'iss': 'pinecreek.com'},
                       KEY_PAIRS['pinecreek.com'].private_key, 'RS256')
    self.assertSucceeds(self.custom_auth, {'access_token': token}, None)
    self.assertSucceeds(
        self.custom_auth, {'access_token': token}, ((14, 3051, 6365),))
    self.assertSucceeds(
        self.custom_auth, {'access_token': token},
        ((14, 3051, 6365), (14, 3052, 6365), (14, 3052, 6364)))

  def testCustomAreaFailure(self):
    # Expect failure authorizing with area-based authority outside area
    token = jwt.encode({'client_id': 'client', 'iss': 'pinecreek.com'},
                       KEY_PAIRS['pinecreek.com'].private_key, 'RS256')
    self.assertFails(
      self.custom_auth, {'access_token': token}, ((14, 3050, 6365),))
    self.assertFails(
      self.custom_auth, {'access_token': token},
      ((14, 3050, 6365), (14, 3051, 6365)))
    self.assertFails(
      self.custom_auth, {'access_token': token},
      ((14, 3052, 6365), (14, 3053, 6365)))

    # Expect failure with wrong issuer
    token = jwt.encode({'client_id': 'client', 'iss': 'wrong iss'},
                       KEY_PAIRS['pinecreek.com'].private_key, 'RS256')
    self.assertFails(self.custom_auth, {'access_token': token}, None)

  def testCustomRangeSuccess(self):
    # Expect success authorizing with range-based authority
    token = jwt.encode({'client_id': 'client', 'iss': 'zion.gov'},
                       KEY_PAIRS['zion.gov'].private_key, 'RS256')
    self.assertSucceeds(self.custom_auth, {'access_token': token}, None)
    self.assertSucceeds(
      self.custom_auth, {'access_token': token}, ((14, 3051, 6365),))
    self.assertSucceeds(
      self.custom_auth, {'access_token': token},
      ((14, 3050, 6365), (14, 3051, 6365), (14, 3050, 6366)))

  def testCustomRangeFailure(self):
    # Expect failure authorizing with range-based authority outside area
    token = jwt.encode({'client_id': 'client', 'iss': 'zion.gov'},
                       KEY_PAIRS['zion.gov'].private_key, 'RS256')
    self.assertFails(
      self.custom_auth, {'access_token': token}, ((14, 3052, 6365),))
    self.assertFails(
      self.custom_auth, {'access_token': token},
      ((14, 3051, 6365), (14, 3052, 6365)))

    # Expect failure with wrong issuer
    token = jwt.encode({'client_id': 'client', 'iss': 'wrong iss'},
                       KEY_PAIRS['zion.gov'].private_key, 'RS256')
    self.assertFails(self.custom_auth, {'access_token': token}, None)


if __name__ == '__main__':
  unittest.main()
