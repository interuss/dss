"""Test of the InterUSS Platform Data Node slippy utilities.

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
import unittest
import slippy_util


class InterUSSSlippyUtilitiesTestCase(unittest.TestCase):

  def testValidCSVConversions(self):
    self.assertEqual([(0.0, 0.0)], slippy_util.ConverCSVtoCoordinates('0,0'))
    self.assertEqual([(40.0, 0.0)], slippy_util.ConverCSVtoCoordinates('40,0'))
    self.assertEqual([(40.4, 0.0)], slippy_util.ConverCSVtoCoordinates('40.4,0'))
    self.assertEqual([(40.4, 110.0)], slippy_util.ConverCSVtoCoordinates('40.4,110'))
    self.assertEqual([(40.4, 110.1)], slippy_util.ConverCSVtoCoordinates('40.4,110.1'))

  def testInvalidCSVConversions(self):
    with self.assertRaises(TypeError):
      slippy_util.ConverCSVtoCoordinates(None)
    with self.assertRaises(TypeError):
      slippy_util.ConverCSVtoCoordinates(0)
    with self.assertRaises(TypeError):
      slippy_util.ConverCSVtoCoordinates('')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('1')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('10 100')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('COORDS')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('10,C')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('91,10')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('10,191')
    with self.assertRaises(ValueError):
      slippy_util.ConverCSVtoCoordinates('10,11,12')

  def testValidPointConversions(self):
    self.assertEqual((0, 0), slippy_util.ConvertPointToTile(0, 0, 0))
    self.assertEqual((1, 1), slippy_util.ConvertPointToTile(1, 0, 0))
    self.assertEqual((2, 2), slippy_util.ConvertPointToTile(2, 0, 0))
    self.assertEqual((3, 1), slippy_util.ConvertPointToTile(2, 34, 110))
    self.assertEqual((412, 204), slippy_util.ConvertPointToTile(9, 34, 110))
    self.assertEqual((412, 307), slippy_util.ConvertPointToTile(9, -34, 110))
    self.assertEqual((99, 307), slippy_util.ConvertPointToTile(9, -34, -110))
    self.assertEqual((99, 204), slippy_util.ConvertPointToTile(9, 34, -110))

  def testInvalidPointConversions(self):
    with self.assertRaises(ValueError):
      self.assertIsNone(slippy_util.ConvertPointToTile(-1, 0, 0))
    with self.assertRaises(ValueError):
      self.assertIsNone(slippy_util.ConvertPointToTile(21, 0, 0))
    with self.assertRaises(ValueError):
      self.assertIsNone(slippy_util.ConvertPointToTile(1, 91, 10))
    with self.assertRaises(ValueError):
      self.assertIsNone(slippy_util.ConvertPointToTile(1, 10, 191))
    with self.assertRaises(ValueError):
      self.assertIsNone(slippy_util.ConvertPointToTile(1, 10, None))

  def testValidPathConversions(self):
    self.assertEqual(1, len(slippy_util.ConvertPathToTile(0, [(0, 0)])))
    self.assertEqual(1, len(slippy_util.ConvertPathToTile(0, [(0, 0), (1, 1.5)])))
    self.assertEqual(2, len(slippy_util.ConvertPathToTile(5, [(0, 0), (1, 1.5)])))
    self.assertEqual(184, len(slippy_util.ConvertPathToTile(15, [(0, 0), (1, 1.5)])))
    # One segment should be the same as two segments that overlapp
    self.assertEqual(len(slippy_util.ConvertPathToTile(10, [(0, 0), (1, 1.5)])),
                     len(slippy_util.ConvertPathToTile(10, [(0, 0), (1, 1.5), (0, 0)])))
    # 4 points are in 4 separate grids, and there are 2 grids underlapping the path
    self.assertEqual(6, len(slippy_util.ConvertPathToTile(9, [(47.5, -103), (47.5, -102.5),
                                                                 (48, -102.5), (48, -103), (47.5, -103)])))

  def testInvalidPathConversions(self):
    with self.assertRaises(TypeError):
      slippy_util.ConvertPathToTile(0, None)
    with self.assertRaises(TypeError):
      slippy_util.ConvertPathToTile(0, 0)
    with self.assertRaises(ValueError):
      slippy_util.ConvertPathToTile(0, '0,0,1,1.5')
    with self.assertRaises(ValueError):
      slippy_util.ConvertPathToTile(0, [])
    with self.assertRaises(TypeError):
      slippy_util.ConvertPathToTile(0, [(0),(1)])

  def testValidPolygonConversions(self):
    self.assertEqual(1, len(slippy_util.ConvertPolygonToTile(0, [(0, 0), (1, 1.5), (2, 0), (0, 0)])))
    self.assertEqual(2, len(slippy_util.ConvertPolygonToTile(5, [(0, 0), (1, 1.5), (2, 0), (0, 0)])))
    # 4 points are in 4 separate grids, and there are 4 grids underlapping the path, and 1 grid surrounded
    self.assertEqual(9, len(slippy_util.ConvertPolygonToTile(9, [(47.5, -103), (47.5, -101.8),
                                                              (48, -101.8), (48, -103), (47.5, -103)])))
    # test the duration of a lot of tiles calculation
    self.assertEqual(7590, len(slippy_util.ConvertPolygonToTile(15, [(47.5, -103), (47.5, -101.8),
                                                                 (48, -101.8), (48, -103), (47.5, -103)])))
