"""The InterUSS Platform Data Node storage API server.

This flexible and distributed system is used to connect multiple USSs operating
in the same general area to share safety information while protecting the
privacy of USSs, businesses, operator and consumers. The system is focused on
facilitating communication amongst actively operating USSs with no details about
UAS operations stored or processed on the InterUSS Platform.

A data node contains all of the API, logic, and data consistency infrastructure
required to perform CRUD (Create, Read, Update, Delete) operations on specific
grid cells. Multiple data nodes can be executed to increase resilience and
availability. This is achieved by a stateless API to service USSs, an
information interface to translate grid cell USS information into the correct
data storage format, and an information consistency store to ensure data is up
to date.

This module is the slippy utilities for conversion to/from points/polygons/tiles.


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
# logging is our log infrastructure used for this application
import logging
import math
from shapely.geometry import LineString
from shapely.geometry import Polygon

# Configure the logger
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_SlippyUtil')

TILE_LIMIT = 50  # Number of tiles to allow for multi get/put/del


def validate_slippy(z, x, y, raise_error=False):
  """Validates slippy tile ranges.

  https://en.wikipedia.org/wiki/Tiled_web_map
  https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames

  Args:
    z: zoom level in slippy tile format
    x: x tile number in slippy tile format
    y: y tile number in slippy tile format
  Returns:
    true if valid, false if not
  """
  try:
    z = int(z)
    x = int(x)
    y = int(y)
    if not 0 <= z <= 20:
      raise ValueError('Zoom must be an integer from 0-20')
    if not 0 <= x < 2**z:
      raise ValueError('x must be an integer from 0-%d' % (2**z - 1))
    if not 0 <= y < 2**z:
      raise ValueError('y must be an integer from 0-%d' % (2**z - 1))
    return True
  except (ValueError, TypeError) as e:
    log.error('Invalid slippy format for tiles %sz, %s, %s!',
              str(z), str(x), str(y))
    log.error('Invalid slippy format error: %s', e)
    if raise_error:
      raise e
    else:
      return False


def convert_csv_to_coordinates(csv):
  """Converts and validates string of CSV coords into array of lat/lon coords."""
  if not csv:
    raise TypeError('CSV of coordinates must be a string in lat,lon[,lat,lon]...format')
  result = []
  coords = csv.split(',')
  if len(coords) % 2 != 0:
    raise ValueError('CSV of coordinates must in lat,lon pairs')
  log.debug('Split coordinates to %s and passed early validation...', coords)
  for a, b in _pairwise(coords):
    lat = float(a)
    lon = float(b)
    if lat >= 90 or lat <= -90 or lon >= 180 or lon <= -180:
      raise ValueError('CSV of coordinates must have valid lat,lon values')
    result.append((lat, lon))
  return result


def convert_tile_to_polygon(zoom, xtile, ytile):
  """Conversion from tile to lat,lon polygon at specific zoom.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      xtile: x tile in slippy terms
      ytile: y tile in slippy terms
  Raises:
      TypeError if parameters are not ints
      ValueError if parameters are not in respective bounds
  Returns:
      list of (lat,lon) points that make up the tile
  """
  validate_slippy(zoom, xtile, ytile, True)
  n = 2.0 ** zoom
  wlon = xtile / n * 360.0 - 180.0
  nlat = math.degrees(math.atan(math.sinh(math.pi * (1 - 2 * ytile / n))))
  elon = (xtile + 1) / n * 360.0 - 180.0
  slat = math.degrees(math.atan(math.sinh(math.pi * (1 - 2 * (ytile + 1) / n))))
  return [(nlat, wlon), (nlat, elon), (slat, elon), (slat, wlon), (nlat, wlon)]


def convert_point_to_tile(zoom, latitude, longitude):
  """Conversion from lat/lon to tile at specific zoom.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      latitude: decimal degrees latitude for the tile
      longitude: decimal degrees latitude for the tile
  Raises:
      TypeError if zoom is not integer or lat/on is not floats
      ValueError if parameters are not in respective bounds
  Returns:
      tuple representing the integer x-tile, y-tile for the
      specified zoom/lat/lon
  """
  zoom = int(zoom)
  latitude = float(latitude)
  longitude = float(longitude)
  if latitude >= 90 or latitude <= -90:
    raise ValueError('latitude must be within range -90 to 90')
  if longitude >= 180 or longitude <= -180:
    raise ValueError('longitude must be within range -180 to 180')
  if zoom < 0 or zoom > 20:
    raise ValueError('zoom level must be within range 0 to 20')
  log.debug('ConvertPointToTile for %.3f, %.3f...', latitude, longitude)
  latitude_rad = math.radians(latitude)
  n = 2.0 ** zoom
  xtile = int((longitude + 180.0) / 360.0 * n)
  ytile = int(
    (1.0 - math.log(math.tan(latitude_rad) +
                    (1 / math.cos(latitude_rad))) / math.pi) / 2.0 * n)
  return xtile, ytile


def convert_path_to_tiles(zoom, points):
  """Conversion from a series of lat/lon to tile(s) at the specific zoom.

  This makes sure to get tiles the path may cross without a point.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      points: list of (lat,lon) that make the path
  Raises:
      ValueError if no tiles can be found from the points, or if the
      points are not in the format of an list of (lat, lon)
      TypeError if the type of zoom or points is not correct
  """
  log.debug('ConvertPathToTiles for %dz and points %s...', zoom, str(points))
  return _convert_shape_to_tiles(zoom, points, is_poly=False)



def convert_polygon_to_tiles(zoom, points):
  """Conversion from a polygon to tile(s) at the specific zoom.

  This makes sure to get tiles the polygon encloses, but the boundary
  does not touch.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      points: list of (lat,lon) that make the polygon.
  Raises:
      ValueError if no tiles can be found from the points, or if the
      points are not in the format of an list of (lat, lon)
      TypeError if the type of zoom or points is not correct
  """
  log.debug('ConvertPolygonToTiles for %dz and points %s...', zoom, str(points))
  return _convert_shape_to_tiles(zoom, points, is_poly=True)


######################################################################
################       INTERNAL FUNCTIONS    #########################
######################################################################
def _convert_shape_to_tiles(zoom, points, is_poly=False):
  """Validates, creates the shape, and finds tiles """
  log.debug('_ConvertShapeToTiles for %dz and points %s for %s poly...',
            zoom, str(points), str(is_poly))
  if not isinstance(points, list):
    raise TypeError('Points must be a list of (lat,lon)')
  if is_poly and len(points) < 3:
    raise ValueError('Must have at least 3 points to make a polygon')
  if not is_poly and len(points) < 2:
    raise ValueError('Must have at least 2 points to make a path')
  if not isinstance(points[0], (tuple,list)):
    raise TypeError('Points must be a list of (lat,lon)')
  if len(points[0]) != 2:
    raise ValueError('Points must be a list of (lat,lon)')
  if is_poly and points[0] != points[len(points) - 1]:
    log.debug('Polygon did not close, assuming close to first point...')
    points.append(points[0])
  shape = Polygon(points) if is_poly else LineString(points)
  result = _calculate_tiles_from_bounding_box(zoom, shape.bounds)
  return _trim_unused_tiles(zoom, result, shape)


def _calculate_tiles_from_bounding_box(zoom, bbox):
  """Calculates tiles from the polygon representing a bounding box."""
  last = None
  result = []
  nlat = bbox[0]
  wlon = bbox[1]
  slat = bbox[2]
  elon = bbox[3]
  points = [(nlat, wlon), (nlat, elon),
            (slat, elon), (slat, wlon), (nlat, wlon)]
  for lat, lon in points:
    if last:
      latdiff = math.fabs(last[0] - lat)
      londiff = math.fabs(last[1] - lon)
      maxdiff = latdiff if latdiff > londiff else londiff
      splits = int(math.fabs(maxdiff) / _degrees_per_tile(zoom)) + 1
      dlat = (lat - last[0]) / float(splits)
      dlon = (lon - last[1]) / float(splits)
      for i in range(1, splits):
        x, y = convert_point_to_tile(zoom, last[0] + i * dlat, last[1] + i * dlon)
        if x is not None and (x, y) not in result:
          result.append((x, y))
    x, y = convert_point_to_tile(zoom, lat, lon)
    if (x, y) not in result:
      result.append((x, y))
      if len(result) > TILE_LIMIT:
        raise OverflowError('Limit of %d tiles impacted exceeded'
                            % (TILE_LIMIT))
    last = (lat, lon)
  if not result:
    raise ValueError('No tiles found for path coordinates')
  return _add_covering_tiles(result)


def _add_covering_tiles(tiles):
  """Fills in tiles enclosed by the tile list provided"""
  result = []
  lastx = -1
  # ensure they are sorted by the X tile and fill in the empty Y's
  for tile in sorted(tiles, key=lambda tup: tup[0]):
    if lastx == -1:
      # first time through
      result.append(tile)
      miny = maxy = tile[1]
    elif lastx == tile[0]:
      # continuation of the same set of x tiles
      result.append(tile)
      if tile[1] < miny:
        miny = tile[1]
      elif tile[1] > maxy:
        maxy = tile[1]
    else:
      # we have moved on to a different x, fill it in
      for y in range(miny, maxy):
        if (lastx, y) not in result:
          result.append((lastx, y))
      # and reset
      result.append(tile)
      miny = maxy = tile[1]
    lastx = tile[0]
  if lastx != -1:
    # take care of the straggler X
    for y in range(miny, maxy):
      if (lastx, y) not in result:
        result.append((lastx, y))
  return result


def _trim_unused_tiles(zoom, tiles, shape):
  """Removes tiles the shape does not touch"""
  result = []
  for tile in tiles:
    poly = Polygon(convert_tile_to_polygon(zoom, tile[0], tile[1]))
    if poly.intersects(shape) or poly.touches(shape):
      result.append(tile)
  return result


def _degrees_per_tile(zoom):
  """Calculates the number of degress stored in each tile

  This could be even more optimized, since longitude is 360 degrees and
  latitude is 180 degrees. However, we just use 180 as it is over
  protective and does not result in a performance impact.
  """
  return 180 / float(2 ** zoom)


def _pairwise(it):
  """Iterator for sets of lon,lat in an array."""
  it = iter(it)
  while True:
    yield next(it), next(it)
