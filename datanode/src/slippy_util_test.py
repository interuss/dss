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

  def testValidateSlippy(self):
    pass

  def testValidateSlippy(self):
    pass

  def testValidCSVConversions(self):
    self.assertEqual([(0.0, 0.0)], slippy_util.convert_csv_to_coordinates('0,0'))
    self.assertEqual([(40.0, 0.0)], slippy_util.convert_csv_to_coordinates('40,0'))
    self.assertEqual([(40.4, 0.0)],
                     slippy_util.convert_csv_to_coordinates('40.4,0'))
    self.assertEqual([(40.4, 110.0)],
                     slippy_util.convert_csv_to_coordinates('40.4,110'))
    self.assertEqual([(40.4, 110.1)],
                     slippy_util.convert_csv_to_coordinates('40.4,110.1'))

  def testInvalidCSVConversions(self):
    with self.assertRaises(TypeError):
      slippy_util.convert_csv_to_coordinates(None)
    with self.assertRaises(TypeError):
      slippy_util.convert_csv_to_coordinates(0)
    with self.assertRaises(TypeError):
      slippy_util.convert_csv_to_coordinates('')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('1')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('10 100')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('COORDS')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('10,C')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('91,10')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('10,191')
    with self.assertRaises(ValueError):
      slippy_util.convert_csv_to_coordinates('10,11,12')

  def testConversionOfTilesToPolygons(self):
    pass


  def testValidPointConversions(self):
    self.assertEqual((0, 0), slippy_util.convert_point_to_tile(0, 0, 0))
    self.assertEqual((1, 1), slippy_util.convert_point_to_tile(1, 0, 0))
    self.assertEqual((2, 2), slippy_util.convert_point_to_tile(2, 0, 0))
    self.assertEqual((3, 1), slippy_util.convert_point_to_tile(2, 34, 110))
    self.assertEqual((412, 204), slippy_util.convert_point_to_tile(9, 34, 110))
    self.assertEqual((412, 307), slippy_util.convert_point_to_tile(9, -34, 110))
    self.assertEqual((99, 307), slippy_util.convert_point_to_tile(9, -34, -110))
    self.assertEqual((99, 204), slippy_util.convert_point_to_tile(9, 34, -110))

  def testInvalidPointConversions(self):
    with self.assertRaises(ValueError):
      slippy_util.convert_point_to_tile(-1, 0, 0)
    with self.assertRaises(ValueError):
      slippy_util.convert_point_to_tile(21, 0, 0)
    with self.assertRaises(ValueError):
      slippy_util.convert_point_to_tile(1, 91, 10)
    with self.assertRaises(ValueError):
      slippy_util.convert_point_to_tile(1, 10, 191)
    with self.assertRaises(TypeError):
      slippy_util.convert_point_to_tile(1, 10, None)
    with self.assertRaises(ValueError):
      slippy_util.convert_path_to_tiles(0, [(0, 0)])
    with self.assertRaises(OverflowError):
      slippy_util.convert_path_to_tiles(15, [(0, 0), (1, 1.5)])

  def testValidPathConversions(self):
    self.assertEqual(1,
                     len(slippy_util.convert_path_to_tiles(0, [(0, 0), (1, 1.5)])))
    self.assertEqual(2,
                     len(slippy_util.convert_path_to_tiles(5, [(0, 0), (1, 1.5)])))
    # One segment should be the same as two segments that overlapp
    self.assertEqual(len(slippy_util.convert_path_to_tiles(10, [(0, 0), (1, 1.5)])),
                     len(slippy_util.convert_path_to_tiles(10, [(0, 0), (1, 1.5),
                                                                (0, 0)])))
    # 4 points are in 4 separate grids,
    # and there are 2 grids underlapping the path
    self.assertEqual(6, len(
      slippy_util.convert_path_to_tiles(9, [(47.5, -103), (47.5, -102.5),
                                            (48, -102.5), (48, -103),
                                            (47.5, -103)])))
    # Corner cutter case that two points are in two grids, but they cut
    # a corner and that grid should be included
    self.assertEqual(3, len(
      slippy_util.convert_path_to_tiles(
        9, [(37.936541030367316, -122.377713074509),
            (37.69672993401783, -122.10422390269278)])))

  def testInvalidPathConversions(self):
    with self.assertRaises(TypeError):
      slippy_util.convert_path_to_tiles(0, None)
    with self.assertRaises(TypeError):
      slippy_util.convert_path_to_tiles(0, 0)
    with self.assertRaises(TypeError):
      slippy_util.convert_path_to_tiles(0, '0,0,1,1.5')
    with self.assertRaises(ValueError):
      slippy_util.convert_path_to_tiles(0, [])
    with self.assertRaises(TypeError):
      slippy_util.convert_path_to_tiles(0, [(0), (1)])
    # test a lot of tiles calculation
    with self.assertRaises(OverflowError):
      slippy_util.convert_polygon_to_tiles(15, [(47.5, -103), (47.5, -101.8),
                                                (48, -101.8), (48, -103),
                                                (47.5, -103)])

  def testValidPolygonConversions(self):
    self.assertEqual(1, len(
      slippy_util.convert_polygon_to_tiles(0, [(0, 0), (1, 1.5), (2, 0), (0, 0)])))
    self.assertEqual(2, len(
      slippy_util.convert_polygon_to_tiles(5, [(0, 0), (1, 1.5), (2, 0), (0, 0)])))
    # check auto closing
    self.assertEqual(
      slippy_util.convert_polygon_to_tiles(9, [(0, 0), (1, 1.5), (2, 0)]),
      slippy_util.convert_polygon_to_tiles(9, [(0, 0), (1, 1.5), (2, 0), (0, 0)]))
    # 4 points are in 4 separate grids,
    # and there are 4 grids underlapping the path, and 1 grid surrounded
    self.assertEqual(9, len(
      slippy_util.convert_polygon_to_tiles(9, [(47.5, -103), (47.5, -101.8),
                                               (48, -101.8), (48, -103),
                                               (47.5, -103)])))

  def testInvalidPolygonConversions(self):
    with self.assertRaises(TypeError):
      slippy_util.convert_polygon_to_tiles(0, None)
    with self.assertRaises(TypeError):
      slippy_util.convert_polygon_to_tiles(0, 0)
    with self.assertRaises(TypeError):
      slippy_util.convert_polygon_to_tiles(0, '0,0,1,1.5')
    with self.assertRaises(ValueError):
      slippy_util.convert_polygon_to_tiles(0, [])
    with self.assertRaises(ValueError):
      slippy_util.convert_polygon_to_tiles(0, [(0), (1)])

  def testSlippyConversionsForSpecialCases(self):
    # 4x4 grid used for these tests at zoom 4
    # 8,8   9,8   10,8    11,8
    # 8,9   9,9   10,9    11,9
    # 8,10  9,10  10,10   11,10
    # 8,11  9,11  10,11   11,11
    # points of interest
    point_8x8 = (-19.808, 20.039)
    point_8x11 = (-65.730, 19.160)
    point_11x11 = (-58.263, 71.367)
    point_11x8 = (-6.839, 82.441)
    # all 16 for all four by polygon
    self.assertEqual(16, len(
      slippy_util.convert_polygon_to_tiles(
        4, [point_8x8, point_8x11, point_11x11, point_11x8])))
    # only 10 by path (no closing the path)
    self.assertEqual(10, len(
      slippy_util.convert_path_to_tiles(
        4, [point_8x8, point_8x11, point_11x11, point_11x8])))
    # corner to corner should be 7
    self.assertEqual(7, len(
      slippy_util.convert_path_to_tiles(
        4, [point_8x8, point_11x11])))
    self.assertEqual(7, len(
      slippy_util.convert_path_to_tiles(
        4, [point_8x11, point_11x8])))

    # triangle to the bottom is 11
    self.assertEqual(11, len(
      slippy_util.convert_polygon_to_tiles(
        4, [point_11x8, point_8x11, point_11x11, point_11x8])))
