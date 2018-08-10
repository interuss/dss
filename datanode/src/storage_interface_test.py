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
ZK_TEST_CONNECTION_STRING = '35.224.64.48:2181,35.188.14.39:2181,35.224.180.72:2181'
PARALLEL_WORKERS = 10


class InterUSSStorageInterfaceTestCase(unittest.TestCase):

  def setUp(self):
    # IMPORTANT: Puts us in a test data location
    self.mm = storage_interface.USSMetadataManager(
        ZK_TEST_CONNECTION_STRING, testgroupid='InterUSSStorageInterfaceTest')

  def tearDown(self):
    # IMPORTANT: Clean out your test data when you are done
    self.mm.delete_testdata()
    self.mm = None

  def testBadConnectionStrings(self):
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          'terrible:connection:1459231232133_string-#$%@',
          testgroupid='InterUSSStorageInterfaceTest')
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          '__init__%password%', testgroupid='InterUSSStorageInterfaceTest')
    with self.assertRaises(ValueError):
      storage_interface.USSMetadataManager(
          '\'printf();\'', testgroupid='InterUSSStorageInterfaceTest')
    with self.assertRaises(KazooTimeoutError):
      storage_interface.USSMetadataManager(
          '123456789101112', testgroupid='InterUSSStorageInterfaceTest')
    with self.assertRaises(KazooTimeoutError):
      storage_interface.USSMetadataManager(
          'google.com:2424,gmail.com:14566',
          testgroupid='InterUSSStorageInterfaceTest')

  def testGetCellNegativeCases(self):
    assert self.mm.get(2, 1, 2**2)['status'] == 'fail'
    # x, y, z are ints
    assert self.mm.get(1, '1a', 1)['status'] == 'fail'
    assert self.mm.get(1, 1, 'aa')['status'] == 'fail'
    assert self.mm.get(None, 1, 1)['status'] == 'fail'
    # x and y tiles max are 2^zoom - 1
    assert self.mm.get(1, 2, 1)['status'] == 'fail'
    assert self.mm.get(2, 5478118, 1)['status'] == 'fail'
    assert self.mm.get(2, 2**2, 1)['status'] == 'fail'
    assert self.mm.get(12, 2**12, 1)['status'] == 'fail'
    assert self.mm.get(1, 17, 1)['status'] == 'fail'
    assert self.mm.get(1, 1, 11)['status'] == 'fail'
    assert self.mm.get(9, 2**8, 2**11)['status'] == 'fail'

  def testGetCellPositiveEmptyCases(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # simple 1,1,1
    r = self.mm.get(1, 1, 1)
    assert r['status'] == 'success'
    assert r['data']['version'] == 0
    # zero case
    r = self.mm.get(0, 0, 0)
    assert r['status'] == 'success'
    assert r['data']['version'] == 0
    r = self.mm.get(11, 0, 5)
    assert r['status'] == 'success'
    assert r['data']['version'] == 0
    # limit in the y direction
    r = self.mm.get(10, 1, 2**10 - 1)
    assert r['status'] == 'success'
    assert r['data']['version'] == 0
    # limit in the x direction
    r = self.mm.get(18, 2**18 - 1, 2**10 - 1)
    assert r['status'] == 'success'
    assert r['data']['version'] == 0
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testPositiveGetSetDeleteCycle(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 2,1,1 get empty
    g = self.mm.get(2, 1, 1)
    assert g['status'] == 'success'
    assert g['data']['version'] == 0
    assert not g['data']['operators']
    # simple set with basic values
    s = self.mm.set(2, 1, 1, g['sync_token'], 'uss', 'uss-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00')
    assert s['status'] == 'success'
    assert s['data']['version'] == 1
    assert len(s['data']['operators']) == 1
    o = s['data']['operators'][0]
    assert o['uss'] == 'uss'
    assert o['operation_endpoint'] == 'https://g.co/flight'
    assert o['operation_format'] == 'GUTMA'
    assert o['version'] == 1
    assert o['minimum_operation_timestamp'] == '2018-01-01T00:00:00+00:00'
    assert o['maximum_operation_timestamp'] == '2018-01-01T01:00:00+00:00'
    # simple delete
    d = self.mm.delete(2, 1, 1, 'uss')
    assert d['status'] == 'success'
    assert d['data']['version'] == 2
    assert not d['data']['operators']
    # simple confirm get is empty
    g = self.mm.get(2, 1, 1)
    assert g['status'] == 'success'
    assert g['data']['version'] == 2
    assert not g['data']['operators']
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testSetCellWithOutdatedsync_token(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 3,1,1 get empty
    g = self.mm.get(3, 1, 1)
    assert g['status'] == 'success'
    assert g['data']['version'] == 0
    assert not g['data']['operators']
    # simple set with basic values
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss1', 'uss1-scope', 'GUTMA',
                    'https://g.co/flight', '2018-01-01T00:00:00+00:00',
                    '2018-01-01T01:00:00+00:00')
    assert s['status'] == 'success'
    assert s['data']['version'] == 1
    assert len(s['data']['operators']) == 1
    # now try to do a set with the original sync token
    s = self.mm.set(3, 1, 1, g['sync_token'], 'uss2', 'uss2-scope', 'GUTMA',
                    'https://h.com/f/3/1/1', '2018-01-01T11:00:00+00:00',
                    '2018-01-01T12:00:00+00:00')
    assert s['status'] == 'fail'
    # confirm version is still the first write
    g = self.mm.get(3, 1, 1)
    assert g['status'] == 'success'
    assert g['data']['version'] == 1
    assert len(g['data']['operators']) == 1
    # Make sure everything is clean
    self.mm.delete_testdata()

  def testSetCellsInParallelWithSamesync_token(self):
    # Make sure everything is clean
    self.mm.delete_testdata()
    # 4,1,1 get empty
    g = self.mm.get(4, 1, 1)
    assert g['status'] == 'success'
    assert g['data']['version'] == 0
    assert not g['data']['operators']
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
    assert g['status'] == 'success'
    assert g['data']['version'] == 1
    assert len(g['data']['operators']) == 1
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
      assert s['status'] == 'fail'
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
      assert s['status'] == 'success'
      o = s['data']['operators'][0]
      assert parser.parse(test[0]) == parser.parse(
          o['minimum_operation_timestamp'])
      assert parser.parse(test[1]) == parser.parse(
          o['maximum_operation_timestamp'])
    # Make sure everything is clean
    self.mm.delete_testdata()
    return


if __name__ == '__main__':
  unittest.main()
