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

# Configure the logger
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_SlippyUtil')


def ConverCSVtoCoordinates(csv):
  """Converts and validates string of CSV coords into array of lat/lon coords."""
  if not csv:
    raise TypeError('CSV of coordinates must be a string in lat,lon[,lat,lon]...format')
  result = []
  coords = csv.split(',')
  if len(coords) % 2 != 0:
    raise ValueError('CSV of coordinates must in lat,lon pairs')
  log.debug('Split coordinates to %s and passed early validation...', coords)
  for a, b in _Pairwise(coords):
    lat = float(a)
    lon = float(b)
    if lat >= 90 or lat <= -90 or lon >= 180 or lon <= -180:
      raise ValueError('CSV of coordinates must have valid lat,lon values')
    result.append((lat, lon))
  return result


def ConvertPointToTile(zoom, latitude, longitude):
  """Calculation from lat/lon to tile at specific zoom."""
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


def ConvertPathToTile(zoom, points):
  """Conversion from a series of lat/lon to tile(s) at the specific zoom.

  This is a bit difficult, since you have to make sure there are enough points
  in the path to ensure each point touches a tile.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      points: array of (lat,lon) that make the path
  Raises:
      ValueError if no tiles can be found from the points
  """
  log.debug('ConvertPathToTile for %dz and points %s...', zoom, str(points))
  last = None
  result = []
  for lat, lon in points:
    if last:
      latdiff = math.fabs(last[0] - lat)
      londiff = math.fabs(last[1] - lon)
      maxdiff = latdiff if latdiff > londiff else londiff
      splits = int(math.fabs(maxdiff) / DegreesPerTile(zoom)) + 1
      dlat = (lat - last[0]) / float(splits)
      dlon = (lon - last[1]) / float(splits)
      for i in range(1, splits):
        x, y = ConvertPointToTile(zoom, last[0] + i * dlat, last[1] + i * dlon)
        if x is not None and (x, y) not in result:
          result.append((x, y))
    x, y = ConvertPointToTile(zoom, lat, lon)
    if (x, y) not in result:
      result.append((x, y))
    last = (lat, lon)
  if not result:
    raise ValueError('No tiles found for path coordinates')
  return sorted(result, key=lambda tup: tup[0])


def ConvertPolygonToTile(zoom, points):
  """Conversion from a series of lat/lon defining a polygon to tile(s) at the specific zoom.

  This is a bit difficult, since you have to make sure there are enough points
  in the polygon to ensure each point touches a tile, and you have to make sure
  the volume within touches all the tiles. Quick way is to make a bounding box for the
  polygon. Expensive way (but minimizes tiles) is to get the tiles for the path and
  fill in the blanks.

  Args:
      zoom: level of zoom in slippy terms (0-20)
      points: array of (lat,lon) that make the polygon. No validity checking is done on
      the polygon for closure or crossing.
  Raises:
      ValueError if no tiles can be found from the points
  """
  log.debug('ConvertPolygonToTile for %dz and points %s...', zoom, str(points))
  pathtiles = ConvertPathToTile(zoom, points)
  result = []
  lastx = -1
  for tile in pathtiles:
    if lastx == -1:
      # first time through
      result.append(tile)
      miny = maxy = tile[1]
    elif lastx == tile[0]:
      # continuation of the same set of tiles
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
    # take care of the straggler
    for y in range(miny, maxy):
      if (lastx, y) not in result:
        result.append((lastx, y))
  return result


def DegreesPerTile(zoom):
  """Calculates the number of degress stored in each tile"""
  return 180 / float(2 ** zoom)


def _Pairwise(it):
  """Iterator for sets of lon,lat in an array."""
  it = iter(it)
  while True:
    yield next(it), next(it)
