"""Utilities for the InterUSS Platform Data Node storage API server tests.


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

import uuid
import uvrs


def make_uvr(uss_id, message_id=None, coords=None):
  """Generate a test UVR with the specified characteristics.

  Args:
    uss_id: ID of USS originator of this UVR
    message_id: UUID-formatted string naming this UVR
    coords: List of (longitude, latitude) pairs describing the polygon boundary
      of this UVR. Defaults to NASA example near Moffett Field. May also specify
      one of the named coordinate sets: {too_big}

  Returns:
    UVR instance
  """
  if not message_id:
    message_id = str(uuid.uuid4())
  if not coords:
    coords = [
      [-122.062176579, 37.40968041145],
      [-122.05187056889, 37.41786527236],
      [-122.03732647634, 37.41786440108],
      [-122.062176579, 37.40968041145]]
  if coords == 'too_big':
    coords = [[-130, 10], [-100, 10], [-115, 50], [-130, 10]]
  elif coords == 'corner_triangle':
    # Points in (500, 800), (499, 799), (499, 799), also spans (499, 800)
    coords = [[-92.1, 36.5], [-92.1, 36.6], [-92.2, 36.6], [-92.1, 36.5]]
  elif coords == '800box':
    # Spans (500, 800) and (501, 800)
    coords = [[-92.1, 36.5], [-91.9, 36.5], [-92.1, 36.55], [-92.1, 36.5]]
  return uvrs.Uvr({
    'message_id': message_id,
    'origin': 'USS',
    'originator_id': uss_id,
    'type': 'DYNAMIC_RESTRICTION',
    'cause': 'SAFETY',
    'geography': {
      'type': 'Polygon',
      'coordinates': [coords]},
    'effective_time_begin': '2016-11-29T01:16:41.727Z',
    'effective_time_end': '2016-11-30T01:16:41.727Z',
    'permitted_uas': ['NOT_SET'],
    'permitted_gufis': ['00000000-0000-4444-8888-FEEDDEADFFFF'],
    'min_altitude': {
      'altitude_value': 2300,
      'vertical_reference': 'WGS84',
      'units_of_measure': 'FT'},
    'max_altitude': {
      'altitude_value': 2700,
      'vertical_reference': 'WGS84',
      'units_of_measure': 'FT'},
    'reason': 'A UA leaving defined volume'})

def csv_coords_of_uvr(uvr):
  """Get a CSV string of lat,lng,lat,lng,... coords for UVR geography.

  Args:
    uvr: UVR instance with valid geography

  Returns:
    String containing CSV-format coordinates
  """
  return ','.join(['%g,%g' % (c[1], c[0])
                   for c in uvr['geography']['coordinates'][0]])
