'''Test of the InterUSS Platform Data Node storage API server.

Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the 'License');
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an 'AS IS' BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
'''
import json
import unittest
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

  def testSyncTokenInHeader(self):
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    # Put a record in there with an invalid token
    qs = dict(
        uss_baseurl='https://g.co/f',
        announcement_level=False,
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=qs,
        headers={'sync_token': 'arbitrary_and_NOT_VALID'})
    self.assertEqual(result.status_code, 409)
    # Put a record in there with an invalid token
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=qs,
        headers={'sync_token': s})
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
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/1/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/19/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/10/100/100'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/15/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/15/9132/1425'))

  def testIncorrectGetsOnGridCells(self):
    self.assertEqual(self.app.get('/GridCellOperators/1/1/1').status_code, 404)
    self.assertEqual(self.app.get('/GridCellOperator').status_code, 404)
    self.assertEqual(self.app.get('/GridCellOperator/admin').status_code, 404)
    self.assertEqual(self.app.get('/GridCellOperator/1/1/1/admin').status_code,
                     404)
    self.assertEqual(self.app.get('/GridCellOperator/1a/1/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellOperator/99/1/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellOperator/1/99/1').status_code, 400)
    self.assertEqual(self.app.get('/GridCellOperator/1/1/99').status_code, 400)

  def testIncorrectPutsOnGridCells(self):
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(self.app.put(
        '/GridCellOperators/1/1/1',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code, 404)
    self.assertEqual(self.app.put(
        '/GridCellOperator',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code, 404)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1a/1/1',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/99/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            # sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            # scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            # operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            # operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # maximum_operation_timestamp='2018-01-02'
            minimum_operation_timestamp='2018-01-01')).status_code, 400)
    self.assertEqual(self.app.put(
        '/GridCellOperator/1/1/1', data={
          'sync_token': 'NOT_VALID'
        }).status_code, 400)
    self.assertEqual(self.app.put('/GridCellOperator/1/1/1').status_code, 400)

  def testIncorrectDeletesOnGridCells(self):
    self.assertEqual(
        self.app.delete('/GridCellOperators/1/1/1').status_code, 404)
    self.assertEqual(self.app.delete('/GridCellOperator').status_code, 404)
    self.assertEqual(
        self.app.delete('/GridCellOperator/admin').status_code, 404)
    self.assertEqual(self.app.delete(
        '/GridCellOperator/1/1/1/admin').status_code, 404)
    self.assertEqual(
        self.app.delete('/GridCellOperator/1a/1/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellOperator/99/1/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellOperator/1/99/1').status_code, 400)
    self.assertEqual(
        self.app.delete('/GridCellOperator/1/1/99').status_code, 400)

  def CheckEmptyGridCell(self, result):
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 0)
    self.assertEqual(len(j['data']['operators']), 0)
    return True

  def testFullValidSequenceOfGetPutDelete(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            uss_baseurl='https://g.co/r',
            announcement_level=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    # Delete the record
    result = self.app.delete('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    # Make sure it is gone
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 0)

  def testPutWithJSONFormattedBody(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with the bare minimum fields
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01',
      'maximum_operation_timestamp': '2018-01-02',
      'announcement_level': 'NONE'
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 1)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(o['uss_baseurl'], 'https://g.co/r')
    self.assertEqual(o['announcement_level'], 'NONE')
    self.assertEqual(len(o['operations']), 0)
    # Put a record in there with lots of data
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': 'YES',
      'operations': [
        {'gufi': 'G00F1', 'operation_signature': 'signed4.1',
          'effective_time_begin': '2018-01-01T01:00:00-05:00',
          'effective_time_end': '2018-01-01T02:00:00-05:00'},
        {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
         'effective_time_begin': '2018-01-01T02:00:00-05:00',
         'effective_time_end': '2018-01-01T03:00:00-05:00'},
        {'gufi': 'G00F3', 'operation_signature': 'signed4.3',
         'effective_time_begin': '2018-01-01T03:00:00-05:00',
         'effective_time_end': '2018-01-01T04:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 2)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(o['uss_baseurl'], 'https://g.co/r')
    self.assertEqual(o['announcement_level'], 'YES')
    self.assertEqual(len(o['operations']), 3)
    # Put a record in there with too much data
    joper = {
      'version': 99,
      'timestamp': '2018-01-01T01:00:00-05:00',
      'some_unknown_field': True,
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': True,
      'operations': [
        {'version': 99, 'timestamp': '2018-01-01T01:00:00-05:00',
         'some_unknown_field': 'True',
         'gufi': 'G00F1', 'operation_signature': 'signed4.1',
         'effective_time_begin': '2018-01-01T01:00:00-05:00',
         'effective_time_end': '2018-01-01T02:00:00-05:00'},
        {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
         'effective_time_begin': '2018-01-01T02:00:00-05:00',
         'effective_time_end': '2018-01-01T03:00:00-05:00'},
        {'gufi': 'G00F3', 'operation_signature': 'signed4.3',
         'effective_time_begin': '2018-01-01T03:00:00-05:00',
         'effective_time_end': '2018-01-01T04:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        data=json.dumps(joper),
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 3)
    self.assertEqual(len(j['data']['operators']), 1)
    self.assertNotEqual(j['data']['timestamp'], '2018-01-01T01:00:00-05:00')
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 3)
    self.assertEqual(o['operations'][0]['version'], 3)
    self.assertNotEqual(o['operations'][0]['timestamp'],
                        '2018-01-01T01:00:00-05:00')


  def testValidDeleteOperations(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with operations
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': True,
      'operations': [
        {'gufi': 'G00F1', 'operation_signature': 'signed4.1',
         'effective_time_begin': '2018-01-01T01:00:00-05:00',
         'effective_time_end': '2018-01-01T02:00:00-05:00'},
        {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
         'effective_time_begin': '2018-01-01T02:00:00-05:00',
         'effective_time_end': '2018-01-01T03:00:00-05:00'},
        {'gufi': 'G00F3', 'operation_signature': 'signed4.3',
         'effective_time_begin': '2018-01-01T03:00:00-05:00',
         'effective_time_end': '2018-01-01T04:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 3)
    # Delete an operation
    result = self.app.delete('/GridCellOperation/1/1/1/G00F1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 2)
    # Make sure it is gone
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 2)

  def testInvalidDeleteOperations(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with operations
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': True,
      'operations': [
        {'gufi': 'G00F1', 'operation_signature': 'signed4.1',
         'effective_time_begin': '2018-01-01T01:00:00-05:00',
         'effective_time_end': '2018-01-01T02:00:00-05:00'},
        {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
         'effective_time_begin': '2018-01-01T02:00:00-05:00',
         'effective_time_end': '2018-01-01T03:00:00-05:00'},
        {'gufi': 'G00F3', 'operation_signature': 'signed4.3',
         'effective_time_begin': '2018-01-01T03:00:00-05:00',
         'effective_time_end': '2018-01-01T04:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 3)
    # Delete the wrong operation
    result = self.app.delete('/GridCellOperation/1/1/1/G00FAA')
    self.assertEqual(result.status_code, 404)
    result = self.app.delete('/GridCellOperation/1/1/1')
    self.assertEqual(result.status_code, 404)

  def testMultipleOperators(self):
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with operations
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': True,
      'operations': [
        {'gufi': 'G00F1', 'operation_signature': 'signed4.1',
         'effective_time_begin': '2018-01-01T01:00:00-05:00',
         'effective_time_end': '2018-01-01T02:00:00-05:00'},
        {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
         'effective_time_begin': '2018-01-01T02:00:00-05:00',
         'effective_time_end': '2018-01-01T03:00:00-05:00'},
        {'gufi': 'G00F3', 'operation_signature': 'signed4.3',
         'effective_time_begin': '2018-01-01T03:00:00-05:00',
         'effective_time_end': '2018-01-01T04:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s,
                 'access_token': 'InterUSSStorageAPITestCase-1'})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 1)
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s,
                 'access_token': 'InterUSSStorageAPITestCase-2'})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 2)
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s,
                 'access_token': 'InterUSSStorageAPITestCase-3'})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 3)

  def testMultipleUpdates(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with the wrong sequence token
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token='arbitrary_and_NOT_VALID',
            uss_baseurl='https://g.co/r',
            announcement_level=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 409)
    # Put a record in there with the right sequence token
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            uss_baseurl='https://g.co/r',
            announcement_level=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 200)
    # Try to put a record in there again with the old sequence token
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            uss_baseurl='https://g.co/r',
            announcement_level=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(result.status_code, 409)

  def testValidUpsertOperations(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with operations
    joper = {
      'uss_baseurl': 'https://g.co/r',
      'minimum_operation_timestamp': '2018-01-01T01:00:00-05:00',
      'maximum_operation_timestamp': '2018-01-01T04:00:00-05:00',
      'announcement_level': True,
      'operations': [
        {'gufi': 'G00F1', 'operation_signature': 'signed4.1',
         'effective_time_begin': '2018-01-01T01:00:00-05:00',
         'effective_time_end': '2018-01-01T02:00:00-05:00'}
      ]
    }
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        json=joper,
        headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 1)
    # Add a new operation
    joper = {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put('/GridCellOperation/1/1/1/G00F2',
                          json=joper, headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 2)
    self.assertIn('G00F1', [d['gufi'] for d in o['operations']])
    self.assertIn('G00F2', [d['gufi'] for d in o['operations']])
    # Update the old operation
    joper = {'gufi': 'G00F2', 'operation_signature': 'UNSIGNED',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put('/GridCellOperation/1/1/1/G00F2',
                          json=joper, headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(len(j['data']['operators']), 1)
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(len(o['operations']), 2)
    self.assertIn('UNSIGNED',
                  [d['operation_signature'] for d in o['operations']])
    self.assertIn(1, [d['version'] for d in o['operations']])
    self.assertIn(3, [d['version'] for d in o['operations']])
    print(j)

  def testVerbose(self):
    storage_api.InitializeConnection([
      '-z', ZK_TEST_CONNECTION_STRING, '-t', 'InterUSSStorageAPITestCase',
      '-v'
    ])


if __name__ == '__main__':
  unittest.main()