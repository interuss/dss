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

# NOTE: A zookeeper instance must be available for these tests to succeed.
# To host a suitable zookeeper instance on your local machine, run:
#   docker run --net=host --rm zookeeper

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

  def testIntrospectWithBadTokens(self):
    result = self.app.get('/introspect', headers={'Authorization': None})

    # If you receive a 403 here, check that your FIMS_AUTH and
    # INTERUSS_PUBLIC_KEY environment variables are set appropriately.
    self.assertEqual(400, result.status_code)

    result = self.app.get('/introspect?token=NOTVALID',
                          headers={'Authorization': None})
    self.assertEqual(400, result.status_code)
    result = self.app.get('/introspect?access_token=NOTVALID',
                          headers={'Authorization': None})
    self.assertEqual(400, result.status_code)
    result = self.app.get('/introspect', headers={'access_token': 'NOTVALID'})
    self.assertEqual(400, result.status_code)

  def testIntrospectWithExpiredToken(self):
    result = self.app.get(
      '/introspect',
      headers={
        'access_token':
          '1/fFAGRNJru1FTz70BzhT3Zg'
      })

    # If you receive a 403 here, check that your FIMS_AUTH and
    # INTERUSS_PUBLIC_KEY environment variables are set appropriately.
    self.assertEqual(400, result.status_code)

  def testValidAuthorizationTokensInTest(self):
    for field in ('access_token', 'Authorization'):
      for token in (TESTID, TESTID + 'a', '123' + TESTID):
        result = self.app.get('/GridCellMetaData/1/1/1',
                              headers={field: token})
        self.assertEqual(200, result.status_code)

  def testInvalidAuthorizationTokensInTest(self):
    for field in ('Authorization', 'access_token'):
      for token in ('not_valid', '', None):
        result = self.app.get('/GridCellMetaData/1/1/1',
                              headers={field: token})
        self.assertAlmostEqual(400, result.status_code, delta=3)

  def testIntrospectWithValidToken(self):
    self.assertIsNotNone(os.environ.get('FIMS_AUTH'))
    self.assertIsNotNone(os.environ.get('INTERUSS_PUBLIC_KEY'))
    endpoint = 'https://utmalpha.arc.nasa.gov//fimsAuthServer/oauth/token?grant_type=client_credentials&scope=utm.nasa.gov_write.conflictmanagement'
    headers = {'Authorization': 'Basic ' + os.environ.get('FIMS_AUTH', '')}
    r = requests.post(endpoint, headers=headers)
    self.assertEqual(200, r.status_code)
    token = r.json()['access_token']
    result = self.app.get('/introspect', headers={'access_token': token})
    self.assertEqual(200, result.status_code)
    result = self.app.get('/introspect/10/223/355', headers={'access_token': token})
    self.assertEqual(200, result.status_code)
    result = self.app.get(
        '/introspect/10?coords=48.832,-101.832,47.954,-101.832,47.954,-100.501,'
        '48.832,-100.501,48.832,-101.832&coord_type=polygon',
        headers={'access_token': token})
    self.assertEqual(200, result.status_code)

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
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/1/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/19/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/10/100/100'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/15/1/1'))
    self.CheckEmptyGridCell(self.app.get('/GridCellMetaData/15/9132/1425'))

  def testIncorrectGetsOnGridCells(self):
    self.assertEqual(404, self.app.get('/GridCellMetaDatas/1/1/1').status_code)
    self.assertEqual(404, self.app.get('/GridCellMetaData').status_code)
    self.assertEqual(404, self.app.get('/GridCellMetaData/admin').status_code)
    self.assertEqual(404,
                     self.app.get('/GridCellMetaData/1/1/1/admin').status_code)
    self.assertEqual(400, self.app.get('/GridCellMetaData/1a/1/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellMetaData/99/1/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellMetaData/1/99/1').status_code)
    self.assertEqual(400, self.app.get('/GridCellMetaData/1/1/99').status_code)

  def testIncorrectPutsOnGridCells(self):
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(404, self.app.put(
      '/GridCellMetaDatas/1/1/1',
      query_string=dict(
        sync_token=s,
        flight_endpoint='https://g.co/f1',
        priority_flight_callback='https://g.co/r')).status_code)
    self.assertEqual(404, self.app.put(
      '/GridCellMetaData',
      query_string=dict(
        ssync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1a/1/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/99/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        # sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token=s,
        # scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        # operation_endpoint='https://g.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        # operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        # minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token=s,
        scope='https://g.co/r',
        operation_endpoint='https://g.co/f',
        operation_format='NASA',
        # maximum_operation_timestamp='2018-01-02'
        minimum_operation_timestamp='2018-01-01')).status_code)
    self.assertEqual(409, self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        sync_token='NOT_VALID',
        scope='https://g.co/r',
        # operation_endpoint='https://g.co/f',
        # operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02',
        public_portal_endpoint='https://g.co/pp',
        flight_info_endpoint='https://g.co/fi')).status_code)
    self.assertEqual(400, self.app.put(
      '/GridCellMetaData/1/1/1', data={
        'sync_token': 'NOT_VALID'
      }).status_code)
    self.assertEqual(400, self.app.put('/GridCellMetaData/1/1/1').status_code)

  def testIncorrectDeletesOnGridCells(self):
    self.assertEqual(404,
                     self.app.delete('/GridCellMetaDatas/1/1/1').status_code)
    self.assertEqual(404, self.app.delete('/GridCellMetaData').status_code)
    self.assertEqual(404,
                     self.app.delete('/GridCellMetaData/admin').status_code)
    self.assertEqual(404, self.app.delete(
      '/GridCellMetaData/1/1/1/admin').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellMetaData/1a/1/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellMetaData/99/1/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellMetaData/1/99/1').status_code)
    self.assertEqual(400,
                     self.app.delete('/GridCellMetaData/1/1/99').status_code)

  def CheckEmptyGridCell(self, result):
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual('success', j['status'])
    self.assertEqual(0, j['data']['version'])
    self.assertEqual(0, len(j['data']['operators']))
    return True

  def testFullValidSequenceOfGetPutDelete(self):
    # Make sure it is empty
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
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
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    # Delete the record
    result = self.app.delete('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    # Make sure it is gone
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(0, len(j['data']['operators']))

  def testValidPublicPortalGetPutDeleteSequence(self):
    # Make sure it is empty
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in there
    result = self.app.put(
      '/GridCellMetaData/1/1/1',
      query_string=dict(
        scope='https://g.co/r',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02',
        public_portal_endpoint='https://g.co/pp',
        flight_info_endpoint='https://g.co/fi'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])
    s = j['sync_token']
    # Delete the record
    result = self.app.delete('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])
    # Make sure it is gone
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(0, len(j['data']['operators']))

  def testMultipleUpdates(self):
    # Make sure it is empty
    result = self.app.get('/GridCellMetaData/1/1/1')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
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
    self.assertEqual(409, result.status_code)
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
    self.assertEqual(200, result.status_code)
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
    self.assertEqual(409, result.status_code)

  def testMultipleGridCellGets(self):
    # for this zoom level (10), the points refer to the following tiles:
    # (512, 512), (512, 509), (514, 509), (514, 512)
    # Path includes the following (in addition to points):
    # (512, 510), (512, 511), (513, 509), (514, 511), (514, 510)
    # Polygon includes the following (in addition to path):
    # (513, 512), (513, 510),(513, 511)
    # Make sure it is empty, try points first
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    multisync = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Now write to one and make sure the sync token changes
    result = self.app.get('/GridCellMetaData/10/512/512')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in one of the cells
    self.app.put(
      '/GridCellMetaData/10/512/512',
      query_string=dict(
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(multisync, j['sync_token'])
    self.assertEqual(1, len(j['data']['operators']))
    # Now do it with a path
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(1, len(j['data']['operators']))
    # Put a record in one of the cells that only applies to the path
    result = self.app.get('/GridCellMetaData/10/512/510')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    self.app.put(
      '/GridCellMetaData/10/512/510',
      query_string=dict(
        sync_token=s,
        scope='https://g2.co/r',
        operation_endpoint='https://g2.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(2, len(j['data']['operators']))
    # and make sure only one still in the point method
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))
    # and a polygon, add a record only applies to the polygon grid
    result = self.app.get('/GridCellMetaData/10/513/511')
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    self.app.put(
      '/GridCellMetaData/10/513/511',
      query_string=dict(
        sync_token=s,
        scope='https://g3.co/r',
        operation_endpoint='https://g3.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(3, len(j['data']['operators']))
    # and make sure only one still in the point method
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))

  def testMultipleGridCellDeletes(self):
    # Put a record in two of the cells
    result = self.app.get('/GridCellMetaData/10/512/512')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    self.app.put(
      '/GridCellMetaData/10/512/512',
      query_string=dict(
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    result = self.app.get('/GridCellMetaData/10/512/510')
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellMetaData/10/512/510',
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
      '/GridCellMetaData/10/512/510',
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
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1',
                                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(3, len(j['data']['operators']))
    result = self.app.delete('/GridCellsMetaData/10',
                             query_string=dict(coords='0,0,1,0,1,1,0,1',
                                               coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(1, len(j['data']['operators']))

  def testMultipleGridCellPut(self):
    # for this zoom level (10), the points refer to the following tiles:
    # (512, 512), (512, 509), (514, 509), (514, 512)
    # Path includes the following (in addition to points):
    # (512, 510), (512, 511), (513, 509), (514, 511), (514, 510)
    # Polygon includes the following (in addition to path):
    # (513, 512), (513, 510),(513, 511)
    # Make sure it is empty, try points first
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(0, len(j['data']['operators']))
    # Put a record in all of the point cells
    result = self.app.put(
      '/GridCellsMetaData/10',
      headers={'access_token': TESTID + '1'},
      query_string=dict(
        coords='0,0,1,0,1,1,0,1',
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])
    s = j['sync_token']
    self.assertEqual(4, len(j['data']['operators']))
    # Now do it with a path
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='path'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(4, len(j['data']['operators']))
    # Put a record in all of the path cells
    result = self.app.put(
      '/GridCellsMetaData/10',
      headers={'access_token': TESTID + '2'},
      query_string=dict(
        coords='0,0,1,0,1,1,0,1',
        coord_type='path',
        sync_token=s,
        scope='https://g2.co/r',
        operation_endpoint='https://g2.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 9, len(j['data']['operators']))
    # and make sure eight in the point method
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 4, len(j['data']['operators']))
    # and a polygon, add records that applies to the polygon grid
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(
                            coords='0,0,1,0,1,1,0,1',
                            coord_type='polygon'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(4 + 9, len(j['data']['operators']))
    result = self.app.put(
      '/GridCellsMetaData/10',
      headers={'access_token': TESTID + '3'},
      query_string=dict(
        coords='0,0,1,0,1,1,0,1',
        coord_type='polygon',
        sync_token=s,
        scope='https://g3.co/r',
        operation_endpoint='https://g3.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertEqual(4 + 9 + 12, len(j['data']['operators']))
    # Put a public portal record without a sync_token
    result = self.app.put(
      '/GridCellsMetaData/10',
      headers={'access_token': TESTID + '1'},
      query_string=dict(
        coords='0,0,1,0,1,1,0,1',
        scope='https://g1.co/r',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02',
        public_portal_endpoint='https://g.co/pp',
        flight_info_endpoint='https://g.co/fi'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])
    s = j['sync_token']
    # Put a public portal record with a sync_token
    result = self.app.put(
      '/GridCellsMetaData/10',
      headers={'access_token': TESTID + '1'},
      query_string=dict(
        coords='0,0,1,0,1,1,0,1',
        sync_token=s,
        scope='https://g1.co/r',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02',
        public_portal_endpoint='https://g.co/pp',
        flight_info_endpoint='https://g.co/fi'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    self.assertNotEqual(s, j['sync_token'])

  def testMultipleGridCellFailures(self):
    self.assertEqual(400, self.app.get('/GridCellsMetaData/10',
                                       query_string=dict(
                                         coords='0,0,1')).status_code)
    self.assertEqual(400, self.app.get('/GridCellsMetaData/10',
                                       query_string=dict(
                                         coords='0,0,1,0,1,1,0,1',
                                         coord_type='rainbows')).status_code)
    self.assertEqual(413, self.app.get('/GridCellsMetaData/18',
                                       query_string=dict(
                                         coords='0,0,1,0,1,1,0,1',
                                         coord_type='polygon')).status_code)
    result = self.app.get('/GridCellsMetaData/10',
                          query_string=dict(coords='0,0,1,0,1,1,0,1'))
    self.assertEqual(200, result.status_code)
    j = json.loads(result.data)
    s = j['sync_token']
    self.assertEqual(400, self.app.put(
      '/GridCellsMetaData/10',
      query_string=dict(
        coords='0,0,1',
        sync_token=s,
        scope='https://g1.co/r',
        operation_endpoint='https://g1.co/f',
        operation_format='NASA',
        minimum_operation_timestamp='2018-01-01',
        maximum_operation_timestamp='2018-01-02')).status_code)

  def testVerbose(self):
    options = storage_api.ParseOptions([
      '-z', ZK_TEST_CONNECTION_STRING, '-t', TESTID,
      '-v'
    ])
    storage_api.InitializeConnection(options)


if __name__ == '__main__':
  unittest.main()
