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
import copy
import json
import uuid

import unittest

import storage_api
import test_utils

ZK_TEST_CONNECTION_STRING = 'localhost:2181'
TESTID = 'storage-api-test'


class InterUSSStorageAPITestCase(unittest.TestCase):

  def setUp(self):
    storage_api.webapp.testing = True
    self.app = storage_api.webapp.test_client()
    options = storage_api.ParseOptions(
      ['-z', ZK_TEST_CONNECTION_STRING, '-t', TESTID])
    storage_api.InitializeConnection(options)

  def tearDown(self):
    storage_api.TerminateConnection()
    storage_api.webapp.testing = False

  def testStatus(self):
    result = self.app.get('/status')
    self.assertEqual(200, result.status_code)
    self.assertIn(b'OK', result.data)
    result = self.app.get('/')
    self.assertEqual(200, result.status_code)
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

  def testValidAuthorizationTokensInTest(self):
    for field in ('access_token', 'Authorization'):
      for token in (TESTID, TESTID + 'a', '123' + TESTID):
        result = self.app.get('/GridCellOperator/1/1/1',
                              headers={field: token})
        self.assertEqual(200, result.status_code)

  def testInvalidAuthorizationTokensInTest(self):
    for field in ('Authorization', 'access_token'):
      for token in ('not_valid', '', None):
        result = self.app.get('/GridCellOperator/1/1/1',
                              headers={field: token})
        self.assertAlmostEqual(400, result.status_code, delta=3)

  def testSlippyConversionWithInvalidData(self):
    result = self.app.get('/slippy')
    self.assertEqual(404, result.status_code)
    result = self.app.get('/slippy/11')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11a')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=1')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=1a,1')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=1,1a')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=91,1')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=1,181')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/21?coords=1,1')
    self.assertEqual(400, result.status_code)
    result = self.app.get('/slippy/11?coords=1,1,2')
    self.assertEqual(400, result.status_code)

  def testSlippyConversionWithValidPoints(self):
    r = self.app.get('/slippy/11?coords=1,1')
    self.assertEqual(200, r.status_code)
    j = json.loads(r.data)
    self.assertEqual(11, j['data']['grid_cells'][0]['zoom'])
    self.assertEqual(1029, j['data']['grid_cells'][0]['x'])
    self.assertEqual(1018, j['data']['grid_cells'][0]['y'])
    self.assertEqual('http://tile.openstreetmap.org/11/1029/1018.png',
                     j['data']['grid_cells'][0]['link'])
    r = self.app.get('/slippy/10?coord_type=point&coords=37.408959,-122.053834')
    self.assertEqual(200, r.status_code)
    j = json.loads(r.data)
    self.assertEqual(10, j['data']['grid_cells'][0]['zoom'])
    self.assertEqual(164, j['data']['grid_cells'][0]['x'])
    self.assertEqual(397, j['data']['grid_cells'][0]['y'])
    self.assertEqual('http://tile.openstreetmap.org/10/164/397.png',
                     j['data']['grid_cells'][0]['link'])
    r = self.app.get('/slippy/11?coords=37.203335,-80.599481')
    self.assertEqual('http://tile.openstreetmap.org/11/565/795.png',
                     json.loads(r.data)['data']['grid_cells'][0]['link'])
    r = self.app.get('/slippy/12?coords=37.203335,-80.599481')
    self.assertEqual('http://tile.openstreetmap.org/12/1130/1591.png',
                     json.loads(r.data)['data']['grid_cells'][0]['link'])
    r = self.app.get('/slippy/13?coords=37.203335,-80.599481')
    self.assertEqual('http://tile.openstreetmap.org/13/2261/3182.png',
                     json.loads(r.data)['data']['grid_cells'][0]['link'])
    j = json.loads(r.data)
    self.assertEqual(
      j, json.loads(self.app.get(
        '/slippy/13?coords=37.203335,-80.599481,37.20334,-80.59948').data))
    r = self.app.get('/slippy/11?coords=0,0,1,1,2,2,3,3')
    self.assertEqual(200, r.status_code)
    self.assertEqual(4, len(json.loads(r.data)['data']['grid_cells']))

  def testSlippyConversionWithValidPaths(self):
    r = self.app.get('/slippy/0?coord_type=path&coords=0,0,1,1.5')
    self.assertEqual(200, r.status_code)
    j = json.loads(r.data)
    self.assertEqual(0, j['data']['grid_cells'][0]['zoom'])
    self.assertEqual(0, j['data']['grid_cells'][0]['x'])
    self.assertEqual(0, j['data']['grid_cells'][0]['y'])
    self.assertEqual('http://tile.openstreetmap.org/0/0/0.png',
                     j['data']['grid_cells'][0]['link'])
    self.assertEqual(
      self.app.get('/slippy/5?coord_type=path&coords=0,0,1,1.5').data,
      self.app.get('/slippy/5?coord_type=path&coords=0,0,1,1.5,0,0').data)

  def testSlippyConversionWithInvalidPaths(self):
    r = self.app.get('/slippy/0?coord_type=path&coords=0')
    self.assertEqual(400, r.status_code)
    r = self.app.get('/slippy/0?coord_type=path&coords=0,1,2')
    self.assertEqual(400, r.status_code)
    r = self.app.get('/slippy/15?coord_type=path&coords=0,0,1,1.5')
    self.assertEqual(400, r.status_code)

  def testSlippyConversionWithValidPolygons(self):
    r = self.app.get('/slippy/0?coord_type=polygon&coords=0,0,1,1.5,0,1.5')
    self.assertEqual(200, r.status_code)
    j = json.loads(r.data)
    self.assertEqual(0, j['data']['grid_cells'][0]['zoom'])
    self.assertEqual(0, j['data']['grid_cells'][0]['x'])
    self.assertEqual(0, j['data']['grid_cells'][0]['y'])
    self.assertEqual('http://tile.openstreetmap.org/0/0/0.png',
                     j['data']['grid_cells'][0]['link'])
    self.assertEqual(
      self.app.get('/slippy/5?coord_type=polygon&coords=0,0,1,1.5,0,1.5').data,
      self.app.get(
        '/slippy/5?coord_type=polygon&coords=0,0,1,1.5,0,0,0,1.5,0,0').data)
    s = '47.5,-103,47.5,-101.8,48,-101.8,48,-103,47.5,-103'
    r = self.app.get('/slippy/9?coord_type=polygon&coords=' + s)
    self.assertEqual(200, r.status_code)
    j = json.loads(r.data)
    self.assertEqual(9, j['data']['grid_cells'][0]['zoom'])
    self.assertEqual(9, len(j['data']['grid_cells']))

  def testSlippyConversionWithInvalidPolygons(self):
    r = self.app.get('/slippy/0?coord_type=polygon&coords=0')
    self.assertEqual(400, r.status_code)
    r = self.app.get('/slippy/0?coord_type=polygon&coords=0,1,2')
    self.assertEqual(400, r.status_code)
    r = self.app.get('/slippy/0?coord_type=polygon&coords=0,0,1,1.5')
    self.assertEqual(400, r.status_code)

  def testMultipleSuccessfulEmptyRandomGets(self):
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/1/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/19/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/10/100/100'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/15/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellOperator/15/9132/1425'))

  def testIncorrectGetsOnGridCells(self):
    self.assertEqual(404, self.app.get('/GridCellOperators/1/1/1').status_code)
    self.assertEqual(404, self.app.get('/GridCellOperator').status_code)
    self.assertEqual(404, self.app.get('/GridCellOperator/admin').status_code)
    self.assertEqual(404,
                     self.app.get('/GridCellOperator/1/1/1/admin').status_code)
    self.assertEqual(400, self.app.get('/GridCellOperator/1a/1/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellOperator/99/1/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellOperator/1/99/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellOperator/1/1/99').status_code)

  def testIncorrectPutsOnGridCells(self):
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(404, self.app.put(
        '/GridCellOperators/1/1/1',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code)
    self.assertEqual(404, self.app.put(
        '/GridCellOperator',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1a/1/1',
        query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False)).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/99/1',
          query_string=dict(
            sync_token=s,
            baseurl='https://g.co/f',
            announce=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            # sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            # scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            # operation_endpoint='https://g.co/f',
            operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            # operation_format='NASA',
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            scope='https://g.co/r',
            operation_endpoint='https://g.co/f',
            operation_format='NASA',
            # maximum_operation_timestamp='2018-01-02'
            minimum_operation_timestamp='2018-01-01')).status_code)
    self.assertEqual(400, self.app.put(
        '/GridCellOperator/1/1/1', data={
          'sync_token': 'NOT_VALID'
        }).status_code, 400)
    self.assertEqual(400, self.app.put('/GridCellOperator/1/1/1').status_code)

  def testIncorrectDeletesOnGridCells(self):
    self.assertEqual(404,
                     self.app.delete('/GridCellOperators/1/1/1').status_code)
    self.assertEqual(404, self.app.delete('/GridCellOperator').status_code)
    self.assertEqual(404,
                     self.app.delete('/GridCellOperator/admin').status_code)
    self.assertEqual(404, self.app.delete(
      '/GridCellOperator/1/1/1/admin').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellOperator/1a/1/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellOperator/99/1/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellOperator/1/99/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellOperator/1/1/99').status_code)

  def CheckEmptyGridCell(self, result):
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual('success', j['status'])
    self.assertEqual(0, j['data']['version'])
    self.assertEqual(0, len(j['data']['operators']))
    return True

  def testFullValidSequenceOfGetPutDelete(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in there
    result = self.app.put(
        '/GridCellOperator/1/1/1',
        query_string=dict(
            sync_token=s,
            uss_baseurl='https://g.co/r',
            announcement_level=False,
            minimum_operation_timestamp='2018-01-01',
            maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    # Delete the record
    result = self.app.delete('/GridCellOperator/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    # Make sure it is gone
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(0, len(j['data']['operators']))

  def testMultipleUpdates(self):
    # Make sure it is empty
    result = self.app.get('/GridCellOperator/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
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
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(j['status'], 'success')
    self.assertEqual(j['data']['version'], 1)
    self.assertEqual(len(j['data']['operators']), 1)
    o = j['data']['operators'][0]
    self.assertEqual(o['uss_baseurl'], 'https://g.co/r')
    self.assertEqual(o['announcement_level'], 'NONE')
    self.assertEqual(len(o['operations']), 0)
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
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual('success', j['status'])
    self.assertEqual(2, j['data']['version'])
    self.assertEqual(1, len(j['data']['operators']))
    self.assertNotEqual('2018-01-01T01:00:00-05:00', j['data']['timestamp'])
    o = j['data']['operators'][0]
    self.assertEqual(3, len(o['operations']))
    self.assertEqual(2, o['operations'][0]['version'])
    self.assertNotEqual('2018-01-01T01:00:00-05:00',
                        o['operations'][0]['timestamp'])


  def testMultipleGridCellDeletes(self):
    # Put a record in two of the cells
    result = self.app.get('/GridCellOperator/10/512/512')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    self.app.put(
      '/GridCellOperator/10/512/512',
      query_string=dict(
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    result = self.app.get('/GridCellOperator/10/512/510')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellOperator/10/512/510',
      query_string=dict(
        sync_token=s,
        scope='https://g2.co/r',
        operation_endpoint='https://g2.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    # Put a record for a different USS in one of the cells
    self.assertEqual(1, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellOperator/10/512/510',
      headers={'access_token': TESTID + '3'},
      query_string=dict(
        sync_token=s,
        scope='https://g3.co/r',
        operation_endpoint='https://g3.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    # Now delete the first USS from all cells, leaving just the uss#3
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(3, len(j['data']['operators']))
    result = self.app.delete('/GridCellsOperator/10',
                             query_string=dict(coords='0,0,1,0,1,1,0,1',
                                               coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))

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
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(1, len(o['operations']))
    # Add a new operation
    joper = {'gufi': 'G00F2', 'operation_signature': 'signed4.2',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put('/GridCellOperation/1/1/1/G00F2',
                          json=joper, headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(2, len(o['operations']))
    self.assertIn('G00F1', [d['gufi'] for d in o['operations']])
    self.assertIn('G00F2', [d['gufi'] for d in o['operations']])
    # Update the old operation
    joper = {'gufi': 'G00F2', 'operation_signature': 'UNSIGNED',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put('/GridCellOperation/1/1/1/G00F2',
                          json=joper, headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(2, len(o['operations']))
    self.assertIn('UNSIGNED',
                  [d['operation_signature'] for d in o['operations']])
    self.assertIn(1, [d['version'] for d in o['operations']])
    self.assertIn(3, [d['version'] for d in o['operations']])

  def testValidMultiGridUpsertOperations(self):
    # Make sure it is empty
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(len(j['data']['operators']), 0)
    # Put a record in there with operations
    joper = {
      'coords': '0,0,1,0,1,1,0,1',
      'coord_type': 'point',
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
      '/GridCellsOperator/10',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4, len(j['data']['operators']))
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(1, len(o['operations']))
    o = j['data']['operators'][1]
    self.assertEqual(1, len(o['operations']))
    # Add a new operation
    joper = {'coords': '0,0,1,0,1,1,0,1', 'coord_type': 'point',
             'gufi': 'G00F2', 'operation_signature': 'signed4.2',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put(
      '/GridCellsOperation/10/G00F2',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(result.status_code, 200)
    j = json.loads(result.data)
    self.assertEqual(4, len(j['data']['operators']))
    s = j['sync_token']
    o = j['data']['operators'][0]
    self.assertEqual(2, len(o['operations']))
    o = j['data']['operators'][1]
    self.assertEqual(2, len(o['operations']))
    self.assertIn('G00F1', [d['gufi'] for d in o['operations']])
    self.assertIn('G00F2', [d['gufi'] for d in o['operations']])
    # Update the old operation
    joper = {'coords': '0,0,1,0,1,1,0,1', 'coord_type': 'point',
             'gufi': 'G00F2', 'operation_signature': 'UNSIGNED',
             'effective_time_begin': '2018-01-01T02:00:00-05:00',
             'effective_time_end': '2018-01-01T03:00:00-05:00'}
    result = self.app.put(
      '/GridCellsOperation/10/G00F2',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4, len(j['data']['operators']))
    o = j['data']['operators'][0]
    self.assertEqual(2, len(o['operations']))
    self.assertIn('UNSIGNED',
                  [d['operation_signature'] for d in o['operations']])
    self.assertIn(1, [d['version'] for d in o['operations']])
    self.assertIn(3, [d['version'] for d in o['operations']])
    # Now delete one of the operations
    joper = {'coords': '0,0,1,0,1,1,0,1', 'coord_type': 'point'}
    result = self.app.delete('/GridCellsOperation/10/G00F1',
      json=joper)
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4, len(j['data']['operators']))
    o = j['data']['operators'][0]
    self.assertEqual(1, len(o['operations']))
    o = j['data']['operators'][1]
    self.assertEqual(1, len(o['operations']))



  def testMultipleGridCellPuts(self):
    # for this zoom level (10), the points refer to the following tiles:
    # (512, 512), (512, 509), (514, 509), (514, 512)
    # Path includes the following (in addition to points):
    # (512, 510), (512, 511), (513, 509), (514, 511), (514, 510)
    # Polygon includes the following (in addition to path):
    # (513, 512), (513, 510),(513, 511)
    # Make sure it is empty, try points first
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in all of the point cells
    joper = {
      'coords': '0,0,1,0,1,1,0,1',
      'coord_type': 'point',
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
      '/GridCellsOperator/10',
      json=joper,
      headers={'access_token': TESTID + '1', 'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])
    s = j['sync_token']
    self.assertEqual(4, len(j['data']['operators']))
    # Now do it with a path
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(4, len(j['data']['operators']))
    # Put a record in all of the path cells
    joper['coord_type'] = 'path'
    result = self.app.put(
      '/GridCellsOperator/10',
      json=joper,
      headers={'access_token': TESTID + '2', 'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 9, len(j['data']['operators']))
    # and make sure eight in the point method
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 4, len(j['data']['operators']))
    # and a polygon, add records that applies to the polygon grid
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(4 + 9, len(j['data']['operators']))
    joper['coord_type'] = 'polygon'
    result = self.app.put(
      '/GridCellsOperator/10',
      json=joper,
      headers={'access_token': TESTID + '3', 'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 9 + 12, len(j['data']['operators']))

  def testMultipleGridCellFailures(self):
    self.assertEqual(400, self.app.get('/GridCellsOperator/10',
                                       query_string=dict(
                                         coords='0,0,1')).status_code)
    self.assertEqual(400, self.app.get('/GridCellsOperator/10',
                                       query_string=dict(
                                         coords='0,0,1,0,1,1,0,1',
                                         coord_type='rainbows')).status_code)
    self.assertEqual(413, self.app.get('/GridCellsOperator/18',
                                       query_string=dict(
                                         coords='0,0,1,0,1,1,0,1',
                                         coord_type='polygon')).status_code)
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(400, self.app.put(
      '/GridCellsOperator/10',
      query_string=dict(
        coords='0,0,1',
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)

  def testMultipleGridCellGets(self):
    # for this zoom level (10), the points refer to the following tiles:
    # (512, 512), (512, 509), (514, 509), (514, 512)
    # Path includes the following (in addition to points):
    # (512, 510), (512, 511), (513, 509), (514, 511), (514, 510)
    # Polygon includes the following (in addition to path):
    # (513, 512), (513, 510),(513, 511)
    # Make sure it is empty, try points first
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    multisync = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Now write to one and make sure the sync token changes
    result = self.app.get('/GridCellOperator/10/512/512')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in one of the cells
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
      '/GridCellOperator/10/512/512',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(multisync, j['sync_token'])
    self.assertEqual(1, len(j['data']['operators']))
    # Now do it with a path
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(1, len(j['data']['operators']))
    # Put a record in one of the cells that only applies to the path
    result = self.app.get('/GridCellOperator/10/512/510')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellOperator/10/512/510',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(2, len(j['data']['operators']))
    # and make sure only one still in the point method
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))
    # and a polygon, add a record only applies to the polygon grid
    result = self.app.get('/GridCellOperator/10/513/511')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellOperator/10/513/511',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(3, len(j['data']['operators']))
    # and make sure only one still in the point method
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))

  def testMultipleGridCellDeletes(self):
    # Put a record in two of the cells
    result = self.app.get('/GridCellOperator/10/512/512')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
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
      '/GridCellOperator/10/512/512',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    result = self.app.get('/GridCellOperator/10/512/510')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellOperator/10/512/510',
      json=joper,
      headers={'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(1, len(j['data']['operators']))
    # Put a record for a different USS in one of the cells
    result = self.app.put(
      '/GridCellOperator/10/512/510',
      json=joper,
      headers={'access_token': TESTID + '3', 'sync_token': s})
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    # Now delete the first USS from all cells, leaving just the uss#3
    result = self.app.get('/GridCellsOperator/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(3, len(j['data']['operators']))
    result = self.app.delete('/GridCellsOperator/10',
                             query_string=dict(coords='0,0,1,0,1,1,0,1',
                                               coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))

  def testUvrs(self):
    uss_id = TESTID
    message_id = str(uuid.uuid4())
    zoom = 11
    uvr = test_utils.make_uvr(uss_id, message_id)
    uvr_json = json.dumps(uvr.to_json())

    def verify_uvr_count(uvr, n):
      s = self.app.get(
          '/GridCellsOperator/%d' % zoom,
          query_string=dict(coords=test_utils.csv_coords_of_uvr(uvr),
                            coords_type='polygon'))
      self.assertEqual(200, s.status_code)
      j = json.loads(s.data)
      if n > 0:
        self.assertEqual(n, len(j['data']['uvrs']))
      else:
        self.assertFalse(j['data']['uvrs'])

    # Make sure grid is empty
    verify_uvr_count(uvr, 0)

    # Make some invalid UVR PUTs and ensure they didn't emplace a UVR
    self.assertEqual(400, self.app.put(
      '/UVR/%d/%s' % (zoom, message_id),
      data='invalid uvr',
      headers={'access_token': uss_id}
    ).status_code)

    self.assertEqual(400, self.app.put(
      '/UVR/%d/%s' % (zoom, message_id),
      json={},
      headers={'access_token': uss_id}
    ).status_code)

    self.assertEqual(400, self.app.put(
      '/UVR/%d/%s' % (zoom, message_id),
      data=uvr_json[0:-1],
      headers={'access_token': uss_id}
    ).status_code)

    self.assertEqual(400, self.app.put(
      '/UVR/%d/%s' % (zoom, 'wrong'),
      data=uvr_json,
      headers={'access_token': uss_id}
    ).status_code)

    self.assertEqual(403, self.app.put(
      '/UVR/%d/%s' % (zoom, message_id),
      data=uvr_json,
      headers={'access_token': 'wrong'}
    ).status_code)

    self.assertEqual(400, self.app.put(
      '/UVR/%d/%s' % (zoom, message_id),
      json={},
      headers={'access_token': uss_id}
    ).status_code)

    uvr_too_big = test_utils.make_uvr(uss_id, coords='too_big')
    self.assertEqual(413, self.app.put(
      '/UVR/%d/%s' % (zoom, uvr_too_big['message_id']),
      json=uvr_too_big.to_json(),
      headers={'access_token': uss_id}
    ).status_code)

    verify_uvr_count(uvr, 0)

    # Correctly emplace a UVR and verify its presence
    self.assertEqual(200, self.app.put(
        '/UVR/%d/%s' % (zoom, message_id),
        data=uvr_json,
        headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_count(uvr, 1)

    # Try to delete the UVR as a different USS
    storage_api.TESTID = 'uss2'
    self.assertEqual(400, self.app.delete(
      '/UVR/%d/%s' % (zoom, message_id),
      data=uvr_json,
      headers={'access_token': 'uss2'}
    ).status_code)
    storage_api.TESTID = uss_id
    verify_uvr_count(uvr, 1)

    # Incorrectly delete the UVR and make sure it's still there
    self.assertEqual(400, self.app.delete(
      '/UVR/%d/%s' % (zoom, 'wrong'),
      data=uvr_json,
      headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_count(uvr, 1)

    bad_uvr = copy.deepcopy(uvr)
    bad_uvr._core['origin'] = 'FIMS'
    self.assertEqual(400, self.app.delete(
      '/UVR/%d/%s' % (zoom, message_id),
      json=bad_uvr.to_json(),
      headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_count(uvr, 1)

    bad_uvr = copy.deepcopy(uvr)
    bad_uvr['geography']['coordinates'][0][1][0] = -122.05187
    self.assertEqual(400, self.app.delete(
      '/UVR/%d/%s' % (zoom, message_id),
      json=bad_uvr.to_json(),
      headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_count(uvr, 1)

    # Correctly delete the UVR and verify its absence
    self.assertEqual(200, self.app.delete(
        '/UVR/%d/%s' % (zoom, message_id),
        data=uvr_json,
        headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_count(uvr, 0)

    def verify_uvr_cell_count(cell, n):
      s = self.app.get(
        '/GridCellOperator/%d/%d/%d' % (zoom, cell[0], cell[1]))
      self.assertEqual(200, s.status_code)
      j = json.loads(s.data)
      if n > 0:
        self.assertEqual(n, len(j['data']['uvrs']))
      else:
        self.assertFalse(j['data']['uvrs'])

    # Emplace overlapping UVRs and check individual cells
    uvr_west = test_utils.make_uvr(uss_id, coords='corner_triangle')
    uvr_east = test_utils.make_uvr(uss_id, coords='800box')
    self.assertEqual(200, self.app.put(
      '/UVR/%d/%s' % (zoom, uvr_west['message_id']),
      json=uvr_west.to_json(),
      headers={'access_token': uss_id}
    ).status_code)
    self.assertEqual(200, self.app.put(
      '/UVR/%d/%s' % (zoom, uvr_east['message_id']),
      json=uvr_east.to_json(),
      headers={'access_token': uss_id}
    ).status_code)

    verify_uvr_cell_count((500, 800), 2)
    verify_uvr_cell_count((499, 800), 1)
    verify_uvr_cell_count((500, 799), 1)
    verify_uvr_cell_count((499, 799), 1)
    verify_uvr_cell_count((501, 800), 1)

    # Remove overlapping UVRs
    self.assertEqual(200, self.app.delete(
      '/UVR/%d/%s' % (zoom, uvr_west['message_id']),
      json=uvr_west.to_json(),
      headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_cell_count((500, 800), 1)
    verify_uvr_cell_count((499, 800), 0)
    verify_uvr_cell_count((500, 799), 0)
    verify_uvr_cell_count((499, 799), 0)
    verify_uvr_cell_count((501, 800), 1)

    self.assertEqual(200, self.app.delete(
      '/UVR/%d/%s' % (zoom, uvr_east['message_id']),
      json=uvr_east.to_json(),
      headers={'access_token': uss_id}
    ).status_code)
    verify_uvr_cell_count((500, 800), 0)
    verify_uvr_cell_count((499, 800), 0)
    verify_uvr_cell_count((500, 799), 0)
    verify_uvr_cell_count((499, 799), 0)
    verify_uvr_cell_count((501, 800), 0)

    # Make sure repeated deletes don't fail
    self.assertEqual(200, self.app.delete(
      '/UVR/%d/%s' % (zoom, uvr_east['message_id']),
      json=uvr_east.to_json(),
      headers={'access_token': uss_id}
    ).status_code)

  def testVerbose(self):
    options = storage_api.ParseOptions([
      '-z', ZK_TEST_CONNECTION_STRING, '-t', TESTID,
      '-v'
    ])
    storage_api.InitializeConnection(options)


if __name__ == '__main__':
  unittest.main()
