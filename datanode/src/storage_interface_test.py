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
ZK_TEST_CONNECTION_STRING = 'localhost:2181'
TESTID = 'storage-interface-test'
PARALLEL_WORKERS = 10


class InterUSSStorageInterfaceTestCase(unittest.TestCase):

  def setUp(self):
    # IMPORTANT: Puts us in a test data location
    self.mm = storage_interface.USSMetadataManager(
        ZK_TEST_CONNECTION_STRING, testgroupid=TESTID)

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
    self.assertEqual(self.mm.get(2, 1, 2**2)['status'], 'fail')
    # x, y, z are ints
    self.assertEqual(self.mm.get(1, '1a', 1)['status'], 'fail')
    self.assertEqual(self.mm.get(1, 1, 'aa')['status'], 'fail')
    self.assertEqual(self.mm.get(None, 1, 1)['status'], 'fail')
    # x and y tiles max are 2^zoom - 1
    self.assertEqual(self.mm.get(1, 2, 1)['status'], 'fail')
    self.assertEqual(self.mm.get(2, 5478118, 1)['status'], 'fail')
    self.assertEqual(self.mm.get(2, 2**2, 1)['status'], 'fail')
    self.assertEqual(self.mm.get(12, 2**12, 1)['status'], 'fail')
    self.assertEqual(self.mm.get(1, 17, 1)['status'], 'fail')
    self.assertEqual(self.mm.get(1, 1, 11)['status'], 'fail')
    self.assertEqual(self.mm.get(9, 2**8, 2**11)['status'], 'fail')

  def testGetCellPositiveEmptyCases(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # simple 1,1,1
    r = self.mm.get(1, 1, 1)
    self.assertEqual(r['status'], 'success')
    self.assertEqual(r['data']['version'], 0)
    # zero case
    r = self.mm.get(0, 0, 0)
    self.assertEqual(r['status'], 'success')
    self.assertEqual(r['data']['version'], 0)
    r = self.mm.get(11, 0, 5)
    self.assertEqual(r['status'], 'success')
    self.assertEqual(r['data']['version'], 0)
    # limit in the y direction
    r = self.mm.get(10, 1, 2**10 - 1)
    self.assertEqual(r['status'], 'success')
    self.assertEqual(r['data']['version'], 0)
    # limit in the x direction
    r = self.mm.get(18, 2**18 - 1, 2**10 - 1)
    self.assertEqual(r['status'], 'success')
    self.assertEqual(r['data']['version'], 0)
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testPositiveGetSetDeleteCycle(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 2,1,1 get empty
    g = self.mm.get(2, 1, 1)
    self.assertEqual(g['status'], 'success')
    self.assertEqual(g['data']['version'], 0)
    self.assertEqual(len(g['data']['operators']), 0)
    # simple set with basic values
    s = self.mm.set(2, 1, 1, g['sync_token'], 'uss', 'uss-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00')
    self.assertEqual(s['status'], 'success')
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
    self.assertEqual(d['status'], 'success')
    self.assertEqual(d['data']['version'], 2)
    self.assertEqual(len(d['data']['operators']), 0)
    # simple confirm get is empty
    g = self.mm.get(2, 1, 1)
    self.assertEqual(g['status'], 'success')
    self.assertEqual(g['data']['version'], 2)
    self.assertEqual(len(g['data']['operators']), 0)
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testSetCellWithOutdatedsync_token(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 3,1,1 get empty
    g = self.mm.get(3, 1, 1)
    self.assertEqual(g['status'], 'success')
    self.assertEqual(g['data']['version'], 0)
    self.assertEqual(len(g['data']['operators']), 0)
    # simple set with basic values
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00')
    self.assertEqual(s['status'], 'success')
    self.assertEqual(s['data']['version'], 1)
    self.assertEqual(len(s['data']['operators']), 1)
    # now try to do a set with the original sync token
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss2', 'uss2-scope', 'GUTMA',
                    'https://h.com/f/3/1/1', '2018-01-01T11:00:00+00:00',
                    '2018-01-01T12:00:00+00:00')
    self.assertEqual(s['status'], 'fail')
    # confirm version is still the first write
    g = self.mm.get(3, 1, 1)
    self.assertEqual(g['status'], 'success')
    self.assertEqual(g['data']['version'], 1)
    self.assertEqual(len(g['data']['operators']), 1)
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testSetCellsInParallelWithSamesync_token(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 4,1,1 get empty
    g = self.mm.get(4, 1, 1)
    self.assertEqual(g['status'], 'success')
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
    t.join()
    # confirm there is only one update
    g = self.mm.get(4, 1, 1)
    self.assertEqual(g['status'], 'success')
    self.assertEqual(g['data']['version'], 1)
    self.assertEqual(len(g['data']['operators']), 1)
    # Make sure everything is clean
    self.mm.delete_testdata()

  def SetCellWorker(self, num, sync_token):
    self.mm.set(4, 1, 1, sync_token, 'uss' + str(num), 'uss-scope' + str(num),
                'GUTMA', 'https://' + str(num) + '.io/flight',
                '2018-01-01T00:00:00+00:00', '2018-01-01T01:00:00+00:00')
    return

  def testSetCellsWithInvalidTimestamps(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 5,1,1 get empty
    s = self.mm.get(5, 1, 1)
    token = s['sync_token']
    testsets = [('Not a valid timestamp', '215664892128621657566'),
                ('2018-01-01H00:00:00+00:00', '2019-01-01!00:00:00'),
                ('2018-01-01T00:00:00+00:00', '215664892128621657566')]
    for test in testsets:
      s = self.mm.set(5, 1, 1, token, 'uss', 'uss-scope', 'GUTMA',
                      'https://g.co/flight', test[0], test[1])
      self.assertEqual(s['status'], 'fail')
    # Make sure everything is clean
    self.mm.delete_testdata()
    return

  def testSetCellsWithValidTimestamps(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
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
                      'https://g.co/flight', test[0], test[1])
      token = s['sync_token']
      self.assertEqual(s['status'], 'success')
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
    # Make sure everything is clean
    self.mm.delete_testdata()
    return


if __name__ == '__main__':
  unittest.main()
