"""Test of the InterUSS Platform Data Node storage API server.

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
import unittest
import requests

import storage_api
ZK_TEST_CONNECTION_STRING = '35.224.64.48:2181,35.188.14.39:2181,35.224.180.72:2181'


class InterUSSStorageAPITestCase(unittest.TestCase):

  def setUp(self):
    storage_api.webapp.testing = True
    self.app = storage_api.webapp.test_client()
    storage_api.InitializeConnection(
        ['-z', ZK_TEST_CONNECTION_STRING, '-t', 'InterUSSStorageAPITestCase'])

  def tearDown(self):
    storage_api.TerminateConnection()
    storage_api.webapp.testing = False

  def testStatus(self):
    result = self.app.get('/status')
    self.assertEqual(result.status_code, 200)
    self.assertIn(b'OK', result.data)
    result = self.app.get('/')
    self.assertEqual(result.status_code, 200)
    self.assertIn(b'OK', result.data)

  def testIntrospectWithBadTokens(self):
    result = self.app.get('/introspect')
    self.assertEqual(result.status_code, 403)
    result = self.app.get('/introspect?token=NOTVALID')
    self.assertEqual(result.status_code, 403)
    result = self.app.get('/introspect?access_token=NOTVALID')
    self.assertEqual(result.status_code, 403)
    result = self.app.get('/introspect', headers={'access_token': 'NOTVALID'})
    self.assertEqual(result.status_code, 403)

  def testIntrospectWithExpiredToken(self):
    self.assertIsNotNone(os.environ.get('FIMS_AUTH'))
    result = self.app.get(
        '/introspect',
        headers={
            'access_token':
            '1/fFAGRNJru1FTz70BzhT3Zg'
        })
    self.assertEqual(result.status_code, 400)

  def testIntrospectWithValidToken(self):
    # pylint: disable=line-too-long
    self.assertIsNotNone(os.environ.get('FIMS_AUTH'))
    self.assertIsNotNone(os.environ.get('INTERUSS_PUBLIC_KEY'))
    endpoint = 'https://utmbeta.arc.nasa.gov//fimsAuthServer/oauth/token?grant_type=client_credentials'
    headers = {'Authorization': 'Basic ' + os.environ.get('FIMS_AUTH', '')}
    r = requests.post(endpoint, headers=headers)
    self.assertEqual(r.status_code, 200)
    token = r.json()['access_token']
    result = self.app.get('/introspect', headers={'access_token': token})
    self.assertEqual(result.status_code, 200)

  def testSlippyConversionWithInvalidData(self):
    result = self.app.get('/slippy')
    self.assertEqual(result.status_code, 404)
    result = self.app.get('/slippy/11')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11a')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=1')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=1a,1')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=1,1a')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=181,1')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=1,91')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/21?coords=1,1')
    self.assertEqual(result.status_code, 400)
    result = self.app.get('/slippy/11?coords=1,1,2')
    self.assertEqual(result.status_code, 400)

  def testSlippyConversionWithValidData(self):
    r = self.app.get('/slippy/11?coords=1,1')
    self.assertEqual(r.status_code, 200)
    j = json.loads(r.data)
    self.assertEqual(j['data']['tiles'][0], [11, 1029, 1018])
    self.assertEqual(j['data']['links'][0],
                    'http://tile.openstreetmap.org/11/1029/1018.png')
    r = self.app.get('/slippy/10?coords=37.203335,-80.599481')
    r = self.app.get('/slippy/10?coords=37.203335,-80.599481')
    self.assertEqual(r.status_code, 200)
    j = json.loads(r.data)
    self.assertEqual(j['data']['tiles'][0], [10, 617, 919])
    self.assertEqual(j['data']['links'][0],
                    'http://tile.openstreetmap.org/10/617/919.png')
    r = self.app.get('/slippy/11?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [11, 1235, 1838])
    r = self.app.get('/slippy/12?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [12, 2471, 3676])
    r = self.app.get('/slippy/13?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [13, 4942, 7353])
    r = self.app.get('/slippy/14?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [14, 9885, 14706])
    r = self.app.get('/slippy/15?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [15, 19770, 29413])
    r = self.app.get('/slippy/16?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [16, 39540, 58826])
    r = self.app.get('/slippy/17?coords=37.203335,-80.599481')
    self.assertEqual(json.loads(r.data)['data']['tiles'][0], [17, 79081, 117653])
    r = self.app.get('/slippy/11?coords=0,0,1,1')
    self.assertEqual(r.status_code, 200)
    j = json.loads(r.data)
    self.assertEqual(len(j['data']['tiles']), 2)
    self.assertEqual(len(j['data']['links']), 2)

  def testMultipleSuccessfulEmptyRandomGets(self):
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/1/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/19/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/10/100/100'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/15/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/15/9132/1425'))

  def testIncorrectGetsOnGridCells(self):
    self.assertEqual(self.app.get('/GridCellMetaDatas/1/1/1').status_code, 404)
    self.assertEqual(self.app.get('/GridCellMetaData').status_code, 404)
    self.assertEqual(self.app.get('/GridCellMetaData/admin').status_code, 404)
    self.assertEqual(self.app.get('/GridCellMetaData/1/1/1/admin').status_code,
                    404)
    self.assertEqual(self.app.get('/GridCellMetaData/1a/1/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellMetaData/99/1/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellMetaData/1/99/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellMetaData/1/1/99').status_code, 400)

  def testIncorrectPutsOnGridCells(self):
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(self.app.put(
        '/GridCellMetaDatas/1/1/1',
        query_string=dict(
            sync_token=s,
            flight_endpoint='https://g.co/f1',
            priority_flight_callback='https://g.co/r')).status_code, 404)
    self.assertEqual(self.app.put(
        '/GridCellMetaData',
        query_string=dict(
            ssync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 404)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1a/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/99/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            # sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            # scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            # operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            # operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # maximum_operation_timestamp='2018-01-02'
            minimum_operation_timestamp='2018-01-01')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellMetaData/1/1/1', data={
            'sync_token': 'NOT_VALID'
        }).status_code, 400)
    self.assertEqual(self.app.put('/GridCellMetaData/1/1/1').status_code, 400)

  def testIncorrectDeletesOnGridCells(self):
    self.assertEqual(
        self.app.delete('/GridCellMetaDatas/1/1/1').status_code, 404)
    self.assertEqual(self.app.delete('/GridCellMetaData').status_code, 404)
    self.assertEqual(
        self.app.delete('/GridCellMetaData/admin').status_code, 404)
    self.assertEqual(self.app.delete(
        '/GridCellMetaData/1/1/1/admin').status_code, 404)
    self.assertEqual(
        self.app.delete('/GridCellMetaData/1a/1/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellMetaData/99/1/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellMetaData/1/99/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellMetaData/1/1/99').status_code, 400)

  def CheckEmptyGridCell(self, result):
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 0)
    self.assertEqual(len(j['data']['operators']), 0)
    return True

  def testFullValidSequenceOfGetPutDelete(self):
    # Make sure it is empty
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there
    result = self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    # Delete the record
    result = self.app.delete('/GridCellMetaData/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    # Make sure it is gone
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 0)

  def testMultipleUpdates(self):
    # Make sure it is empty
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with the wrong sequence token
    result = self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token='arbitrary_and_NOT_VALID',
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 409)
    # Put a record in there with the right sequence token
    result = self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 200)
    # Try to put a record in there again with the old sequence token
    result = self.app.put(
        '/GridCellMetaData/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 409)

  def testVerbose(self):
    storage_api.InitializeConnection([
        '-z', ZK_TEST_CONNECTION_STRING, '-t', 'InterUSSStorageAPITestCase',
        '-v'
    ])


if __name__ == '__main__':
  unittest.main()
