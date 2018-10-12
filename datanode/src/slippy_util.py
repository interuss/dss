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

# Used to calculate degrees in a tile
EARTH_RADIUS_M = 6378137

# Configure the logger
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_SlippyUtil')


def ConverCSVtoCoordinates(csv):
  """Converts and validates string of CSV coords into array of lat/lon coords."""
  result = []
  try:
    coords = csv.split(',')
    if len(coords) % 2 != 0:
      raise ValueError
  except ValueError:
    return None
  log.debug('Split coordinates to %s and passed early validation...', coords)
  for a, b in _Pairwise(coords):
    try:
      lat = float(a)
      lon = float(b)
      if lat >= 90 or lat <= -90 or lon >= 180 or lon <= -180:
        raise ValueError
    except ValueError:
      return None
    result.append((lat, lon))
  return result


def ConvertPointToTile(zoom, latitude, longitude):
  """Calculation from lat/lon to tile at specific zoom."""
  log.debug('ConvertPointToTile for %.3f, %.3f...', latitude, longitude)
  latitude_rad = math.radians(latitude)
  n = 2.0 ** zoom
  xtile = int((longitude + 180.0) / 360.0 * n)
  ytile = int(
    (1.0 - math.log(math.tan(latitude_rad) +
                    (1 / math.cos(latitude_rad))) / math.pi) / 2.0 * n)
  return xtile, ytile


def ConvertPolygonToTile(zoom, points):
  """Conversion from a series of lat/lon to tile(s) at the specific zoom."""
  log.debug('ConvertPolygonToTile for %.3f, %.3f...', latitude, longitude)
  latitude_rad = math.radians(latitude)
  n = 2.0 ** zoom
  xtile = int((longitude + 180.0) / 360.0 * n)
  ytile = int(
    (1.0 - math.log(math.tan(latitude_rad) +
                    (1 / math.cos(latitude_rad))) / math.pi) / 2.0 * n)
  return xtile, ytile


def DegreesPerTile(zoom, latitude):
  """Calculates the number of degress stored in each tile"""
  return EARTH_RADIUS_M * math.cos(latitude) / 2 ** zoom


def _Pairwise(it):
  """Iterator for sets of lon,lat in an array."""
  it = iter(it)
  while True:
    yield next(it), next(it)
