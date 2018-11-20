"""Test of theInterUSS Platform Data Node storage API server.


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
import threading
import unittest
from dateutil import parser
from kazoo.handlers.threading import KazooTimeoutError

import storage_interface
import uss_metadata

# NOTE: A zookeeper instance must be available for these tests to succeed.
# To host a suitable zookeeper instance on your local machine, run:
#   docker run --net=host --rm zookeeper

ZK_TEST_CONNECTION_STRING = 'localhost:2181'
TESTID = 'storage-interface-test'
PARALLEL_WORKERS = 10


class InterUSSStorageInterfaceTestCase(unittest.TestCase):

  def setUp(self):
    # IMPORTANT: Puts us in a test data location
    self.mm = storage_interface.USSMetadataManager(
        ZK_TEST_CONNECTION_STRING, testgroupid=TESTID)
    self.mm.set_verbose()

  def tearDown(self):
    # IMPORTANT: Clean out your test data when you are done
    self.mm.delete_testdata()
    self.mm = None

  def testBadConnectionStrings(self):
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          'terrible:connection:1459231232133_string-#$%@',
          testgroupid=TESTID)
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          '__init__%password%', testgroupid=TESTID)
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          '\'printf();\'', testgroupid=TESTID)
    with self.assertRaises(KazooTimeoutError):
      storage_interface.USSMetadataManager(
          '123456789101112', testgroupid=TESTID)
    with self.assertRaises(KazooTimeoutError):
      storage_interface.USSMetadataManager(
          'google.com:2424,gmail.com:14566',
          testgroupid=TESTID)

  def testGetCellNegativeCases(self):
    self.assertEqual('fail', self.mm.get(2, 1, 2**2)['status'])
    # x, y, z are ints
    self.assertEqual('fail', self.mm.get(1, '1a', 1)['status'])
    self.assertEqual('fail', self.mm.get(1, 1, 'aa')['status'])
    self.assertEqual('fail', self.mm.get(None, 1, 1)['status'])
    # x and y tiles max are 2^zoom - 1
    self.assertEqual('fail', self.mm.get(1, 2, 1)['status'])
    self.assertEqual('fail', self.mm.get(2, 5478118, 1)['status'])
    self.assertEqual('fail', self.mm.get(2, 2**2, 1)['status'])
    self.assertEqual('fail', self.mm.get(12, 2**12, 1)['status'])
    self.assertEqual('fail', self.mm.get(1, 17, 1)['status'])
    self.assertEqual('fail', self.mm.get(1, 1, 11)['status'])
    self.assertEqual('fail', self.mm.get(9, 2**8, 2**11)['status'])

  def testGetCellPositiveEmptyCases(self):
    # simple 1,1,1
    r = self.mm.get(1, 1, 1)
    self.assertEqual('success', r['status'])
    self.assertEqual(r['data']['version'], 0)
    # zero case
    r = self.mm.get(0, 0, 0)
    self.assertEqual('success', r['status'])
    self.assertEqual(r['data']['version'], 0)
    r = self.mm.get(11, 0, 5)
    self.assertEqual('success', r['status'])
    self.assertEqual(r['data']['version'], 0)
    # limit in the y direction
    r = self.mm.get(10, 1, 2**10 - 1)
    self.assertEqual('success', r['status'])
    self.assertEqual(r['data']['version'], 0)
    # limit in the x direction
    r = self.mm.get(18, 2**18 - 1, 2**10 - 1)
    self.assertEqual('success', r['status'])
    self.assertEqual(r['data']['version'], 0)

  def testPositiveGetSetDeleteCycle(self):
    # 2,1,1 get empty
    g = self.mm.get(2, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 0)
    self.assertEqual(len(g['data']['operators']), 0)
    # simple set with basic values
    s = self.mm.set(2, 1, 1, g['sync_token'], 'uss', 'uss-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    self.assertEqual(s['data']['version'], 1)
    self.assertEqual(len(s['data']['operators']), 1)
    o = s['data']['operators'][0]
    self.assertEqual(o['uss'], 'uss')
    self.assertEqual(o['operation_endpoint'], 'https://g.co/flight')
    self.assertEqual(o['operation_format'], 'GUTMA')
    self.assertEqual(o['version'], 1)
    self.assertEqual(o['minimum_operation_timestamp'],
                     '2018-01-01T00:00:00+00:00')
    self.assertEqual(o['maximum_operation_timestamp'],
                     '2018-01-01T01:00:00+00:00')
    # simple delete
    d = self.mm.delete(2, 1, 1, 'uss')
    self.assertEqual('success', d['status'])
    self.assertEqual(d['data']['version'], 2)
    self.assertEqual(len(d['data']['operators']), 0)
    # simple confirm get is empty
    g = self.mm.get(2, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 2)
    self.assertEqual(len(g['data']['operators']), 0)
    # simple set with public portal values
    s = self.mm.set(2, 1, 1, None, 'uss', 'uss-scope', '',
                    '', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', 'https://g.co/pp',
                    'https://g.co/fi')
    self.assertEqual('success', s['status'])
    self.assertEqual(s['data']['version'], 3)
    self.assertEqual(len(s['data']['operators']), 1)
    o = s['data']['operators'][0]
    self.assertEqual(o['uss'], 'uss')
    self.assertEqual(o['operation_endpoint'], '')
    self.assertEqual(o['operation_format'], '')
    self.assertEqual(o['version'], 3)
    self.assertEqual(o['minimum_operation_timestamp'],
                     '2018-01-01T00:00:00+00:00')
    self.assertEqual(o['maximum_operation_timestamp'],
                     '2018-01-01T01:00:00+00:00')
    self.assertEqual(o['public_portal_endpoint'], 'https://g.co/pp')
    self.assertEqual(o['flight_info_endpoint'], 'https://g.co/fi')
    # simple delete
    d = self.mm.delete(2, 1, 1, 'uss')
    self.assertEqual('success', d['status'])
    self.assertEqual(d['data']['version'], 4)
    self.assertEqual(len(d['data']['operators']), 0)
    # simple confirm get is empty
    g = self.mm.get(2, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 4)
    self.assertEqual(len(g['data']['operators']), 0)

  def testSetCellWithOutdatedsync_token(self):
    # 3,1,1 get empty
    g = self.mm.get(3, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 0)
    self.assertEqual(len(g['data']['operators']), 0)
    # simple set with basic values
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    self.assertEqual(s['data']['version'], 1)
    self.assertEqual(len(s['data']['operators']), 1)
    # now try to do a set with the original sync token
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss2', 'uss2-scope', 'GUTMA',
                    'https://h.com/f/3/1/1', '2018-01-01T11:00:00+00:00',
                    '2018-01-01T12:00:00+00:00', '', '')
    self.assertEqual(s['status'], 'fail')
    # confirm version is still the first write
    g = self.mm.get(3, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 1)
    self.assertEqual(len(g['data']['operators']), 1)

  def testSetCellsInParallelWithSamesync_token(self):
    # 4,1,1 get empty
    g = self.mm.get(4, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 0)
    self.assertEqual(len(g['data']['operators']), 0)
    threads = []
    for i in range(PARALLEL_WORKERS):
      t = threading.Thread(
          target=self.SetCellWorker, args=(
              i,
              g['sync_token'],
          ))
      threads.append(t)
      t.start()
    for t in threads:
      t.join()
    # confirm there is only one update
    g = self.mm.get(4, 1, 1)
    self.assertEqual('success', g['status'])
    self.assertEqual(g['data']['version'], 1)
    self.assertEqual(len(g['data']['operators']), 1)

  def SetCellWorker(self, num, sync_token):
    self.mm.set(4, 1, 1, sync_token, 'uss' + str(num), 'uss-scope' + str(num),
                'GUTMA', 'https://' + str(num) + '.io/flight',
                '2018-01-01T00:00:00+00:00', '2018-01-01T01:00:00+00:00', '',
                '')
    return

  def testSetCellsWithInvalidTimestamps(self):
    # 5,1,1 get empty
    s = self.mm.get(5, 1, 1)
    token = s['sync_token']
    testsets = [('Not a valid timestamp', '215664892128621657566'),
                ('2018-01-01H00:00:00+00:00', '2019-01-01!00:00:00'),
                ('2018-01-01T00:00:00+00:00', '215664892128621657566')]
    for test in testsets:
      s = self.mm.set(5, 1, 1, token, 'uss', 'uss-scope', 'GUTMA',
                      'https://g.co/flight', test[0], test[1], '', '')
      self.assertEqual(s['status'], 'fail')

  def testSetCellsWithValidTimestamps(self):
    # 5,1,1 get empty
    s = self.mm.get(5, 1, 1)
    token = s['sync_token']
    testsets = [('2018-01-01T00:00+00', '2019-01-01T01:02:03.12345+00:00'),
                ('2018-01-01T00:00:00', '2019-01-01T01:02:03.12345'),
                ('2018-02-28T23:59:59-07:00', '2018-03-02T23:59:59+08:00'),
                ('2018-01-01T00:00:00', '2019-01-01')
               ]
    for test in testsets:
      s = self.mm.set(5, 1, 1, token, 'uss', 'uss-scope', 'GUTMA',
                      'https://g.co/flight', test[0], test[1], '', '')
      token = s['sync_token']
      self.assertEqual('success', s['status'])
      o = s['data']['operators'][0]
      # Fix up the test cases to compare, this isn't what is sent to the api
      mintest = test[0]
      maxtest = test[1]
      if len(maxtest) <= 10:
        maxtest = maxtest + 'T00:00:00Z'
      if not ('+' in mintest[-6:] or '-' in mintest[-6:] or 'Z' in mintest[-6:]):
        mintest += 'Z'
      if not ('+' in maxtest[-6:] or '-' in maxtest[-6:] or 'Z' in maxtest[-6:]):
        maxtest += 'Z'
      self.assertAlmostEqual(
        0, (parser.parse(mintest) -
            parser.parse(o['minimum_operation_timestamp'])).total_seconds(), 0)
      self.assertAlmostEqual(
        0, (parser.parse(maxtest) -
            parser.parse(o['maximum_operation_timestamp'])).total_seconds(), 0)

  def testGetCellMultipleCellCases(self):
    r1 = self.mm.get_multi(6, [(0, 0), (0, 1), (1, 1)])
    self.assertEqual('success', r1['status'])
    self.assertEqual(0, r1['data']['version'])
    # now do a write to a single cell
    g = self.mm.get(6, 0, 1)
    s = self.mm.set(6, 0, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight/0/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    # confirm the sync_token has changed on multi
    r2 = self.mm.get_multi(6, [(0, 0), (0, 1), (1, 1)])
    self.assertNotEqual(r1['sync_token'], r2['sync_token'])
    # add a few more writes
    g = self.mm.get(6, 1, 1)
    s = self.mm.set(6, 1, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight/1/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    g = self.mm.get(6, 1, 1)
    s = self.mm.set(6, 1, 1, g['sync_token'], 'uss2', 'uss2-scope', 'GUTMA',
                    'https://g2.co/uss2/1/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    # Now get them all and confirm you have the right amount
    r3 = self.mm.get_multi(6, [(0, 0), (0, 1), (1, 1)])
    self.assertEqual('success', r3['status'])
    self.assertNotEqual(r1['sync_token'], r3['sync_token'])
    self.assertNotEqual(r2['sync_token'], r3['sync_token'])
    self.assertEqual(2, r3['data']['version'])
    self.assertEqual(3, len(r3['data']['operators']))

  def testDeleteCellMultipleCellCases(self):
    grids = [(0, 0), (0, 1), (1, 1)]
    r1 = self.mm.get_multi(7, grids)
    self.assertEqual('success', r1['status'])
    self.assertEqual(0, r1['data']['version'])
    g = self.mm.get(7, 0, 1)
    s = self.mm.set(7, 0, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight/0/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    # add a few more writes
    g = self.mm.get(7, 1, 1)
    s = self.mm.set(7, 1, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight/1/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    g = self.mm.get(7, 1, 1)
    s = self.mm.set(7, 1, 1, g['sync_token'], 'uss2', 'uss2-scope', 'GUTMA',
                    'https://g2.co/uss2/1/1', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    # Now try deleting uss1, which would delete from two different cells
    r2 = self.mm.delete_multi(7, grids, 'uss1')
    self.assertEqual('success', r2['status'])
    self.assertNotEqual(r1['sync_token'], r2['sync_token'])
    self.assertEqual(3, r2['data']['version'])
    self.assertEqual(1, len(r2['data']['operators']))

  def testInvalidMultipleCellCases(self):
    r = self.mm.get_multi(6, [])
    self.assertEqual('fail', r['status'])
    r = self.mm.get_multi(6, '0,0,1,1,0,1')
    self.assertEqual('fail', r['status'])
    r = self.mm.delete_multi(6, [(0, 0), (0, 1), (1, 1)], None)
    self.assertEqual('fail', r['status'])
    r = self.mm.delete_multi(21, [(0, 0), (0, 1), (1, 1)], 'uss')
    self.assertEqual('fail', r['status'])

  def testSetMultipleCellCases(self):
    grids = [(0, 0), (0, 1), (0, 2),
             (1, 0), (1, 1), (1, 2),
             (2, 0), (2, 1), (2, 2),
             (3, 0), (3, 1), (3, 2)]
    g = self.mm.get_multi(8, grids)
    self.assertEqual('success', g['status'])
    self.assertEqual(0, g['data']['version'])
    # now do a write to multiple cells
    s = self.mm.set_multi(8, grids, g['sync_token'], 'uss1', 'uss1-scope',
                          'GUTMA', 'https://g.co/flight/{z}/{x}/{y}',
                          '2018-01-01T00:00:00+00:00',
                          '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    self.assertEqual(1, s['data']['version'])
    self.assertEqual(len(grids), len(s['data']['operators']))
    g = self.mm.get_multi(8, grids)
    self.assertEqual('success', s['status'])
    self.assertEqual(1, s['data']['version'])
    self.assertEqual(len(grids), len(s['data']['operators']))

  def testFullCycleMultipleCellCases(self):
    grids = [(0, 0), (0, 1), (1, 1)]
    g = self.mm.get_multi(9, grids)
    self.assertEqual('success', g['status'])
    self.assertEqual(0, g['data']['version'])
    # now do a write to multiple cells
    s = self.mm.set_multi(9, grids, g['sync_token'], 'uss1', 'uss1-scope',
                          'GUTMA', 'https://g1.co/flight/{z}/{x}/{y}',
                          '2018-01-01T00:00:00+00:00',
                          '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    self.assertNotEqual(g['sync_token'], s['sync_token'])
    self.assertEqual(1, s['data']['version'])
    self.assertEqual(len(grids), len(s['data']['operators']))
    # Do a write to two other cells
    grids = [(0, 0), (1, 1)]
    g = self.mm.get_multi(9, grids)
    self.assertEqual('success', g['status'])
    self.assertEqual(1, g['data']['version'])
    s = self.mm.set_multi(9, grids, g['sync_token'], 'uss2', 'uss2-scope',
                          'GUTMA', 'https://g2.co/flight/{z}/{x}/{y}',
                          '2018-01-01T00:00:00+00:00',
                          '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    self.assertNotEqual(g['sync_token'], s['sync_token'])
    self.assertEqual(2, s['data']['version'])
    self.assertEqual(4, len(s['data']['operators']))
    grids = [(0, 0), (0, 1), (1, 1)]
    s = self.mm.get_multi(9, grids)
    self.assertEqual('success', s['status'])
    self.assertEqual(2, s['data']['version'])
    self.assertEqual(5, len(s['data']['operators']))
    multi_token = s['sync_token']
    # Now write to another cell singly and then update using old token
    g = self.mm.get(9, 0, 1)
    s = self.mm.set(9, 0, 1, g['sync_token'], 'uss3', 'uss3-scope', 'GUTMA',
                    'https://g3.co/f/{z}/{x}/{y}', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('success', s['status'])
    s = self.mm.set_multi(9, grids, multi_token, 'ussXX', 'ussXX-scope',
                          'GUTMA', 'https://gXX.co/flight/{z}/{x}/{y}',
                          '2018-01-01T00:00:00+00:00',
                          '2018-01-01T01:00:00+00:00', '', '')
    self.assertEqual('fail', s['status'])
    grids = [(0, 0), (0, 1), (0, 2),
             (1, 0), (1, 1), (1, 2),
             (2, 0), (2, 1), (2, 2),
             (3, 0), (3, 1), (3, 2)]
    g = self.mm.get_multi(9, grids)
    s = self.mm.set_multi(9, grids, g['sync_token'], 'uss4', 'uss4-scope',
                          'GUTMA', 'https://g4.co/flight/{z}/{x}/{y}',
                          '2018-01-01T00:00:00+00:00',
                          '2018-01-01T01:00:00+00:00', '', '')

  def testUSSmetadatAddition(self):
    a = uss_metadata.USSMetadata()
    a.upsert_operator('uss-a', 'scope-a', 'NASA', 'http://a.com/uss',
                      '2018-01-01', '2018-01-02', '', '', 10, 1, 1)
    b1 = uss_metadata.USSMetadata()
    b1.upsert_operator('uss-b', 'scope-b', 'NASA', 'http://b.com/uss',
                      '2018-01-01', '2018-01-02', '', '', 10, 1, 1)
    b2 = uss_metadata.USSMetadata()
    b2.upsert_operator('uss-b', 'scope-b', 'NASA', 'http://b.com/uss',
                       '2018-01-01', '2018-01-02', '', '', 10, 1, 2)
    usss = a + b1 + b2
    self.assertEqual(3, len(usss.operators))
    with self.assertRaises(ValueError):
      usss = a + a
    ax = uss_metadata.USSMetadata()
    ax.upsert_operator('uss-a', 'scope-ax', 'NASA', 'http://ax.com/uss',
                      '2018-01-03', '2018-01-04', '', '', 10, 1, 1)
    with self.assertRaises(ValueError):
      usss = a + ax


if __name__ == '__main__':
  unittest.main()
