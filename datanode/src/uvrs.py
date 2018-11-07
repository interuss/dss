"""Encapsulation of UVR data type.


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

import collections
import copy
import re

import format_utils
import slippy_util


# UVR validation constants
UVR_FIELDS = {
  'message_id', 'origin', 'originator_id', 'type', 'cause', 'geography',
  'effective_time_begin', 'effective_time_end', 'permitted_uas',
  'permitted_gufis', 'actual_time_end', 'min_altitude', 'max_altitude',
  'reason', 'timestamp'}
UVR_ENUMS = {'origin': {'FIMS', 'USS'},
             'type': {'DYNAMIC_RESTRICTION', 'STATIC_ADVISORY'},
             'cause': {'WEATHER', 'ATC', 'SECURITY', 'SAFETY', 'MUNICIPALITY'}}
UVR_PERMITTED_UAS = {
  'NOT_SET', 'PUBLIC_SAFETY', 'SECURITY', 'NEWS_GATHERING', 'VLOS',
  'PART_107', 'PART_101E', 'PART_107X', 'RADIO_LINE_OF_SIGHT'}
UVR_VERTICAL_REFERENCES = {'W84'}
UVR_UNITS_OF_MEASURE = {'FT'}
UUID_MATCHER = re.compile('^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[8-b][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$')
UVR_MIN_POLYGON_COORDS = 4


class Uvr(object):
  def __init__(self, json_dict):
    self._timestamp = json_dict.get('timestamp', None)
    self._core = _validate_uvr(json_dict)

  def __getitem__(self, item):
    if item not in UVR_FIELDS:
      raise KeyError('"%s" is not a valid UVR field')
    elif item == 'timestamp':
      return self._timestamp
    else:
      return self._core[item]

  def __setitem__(self, key, value):
    if key not in UVR_FIELDS:
      raise KeyError('"%s" is not a valid UVR field')
    elif key == 'timestamp':
      self._timestamp = value
    else:
      raise KeyError('"%s" UVR field cannot be set')

  def __eq__(self, other):
    if not isinstance(other, Uvr):
      return False
    return self._core == other._core

  def __ne__(self, other):
    return not self.__eq__(other)

  def to_json(self):
    """Convert to a nested-dict that can be serialized with json.dumps.

    Returns:
      Nested-dict structure.
    """
    result = copy.deepcopy(self._core)
    if self._timestamp:
      result['timestamp'] = self._timestamp
    return result

  def get_tiles(self, zoom):
    """Converts Polygon in UVR into slippy tile format at the specified zoom.

    Args:
      zoom: slippy zoom level
    """
    coord_list = [(c[1], c[0]) for c in self['geography']['coordinates'][0]]
    return slippy_util.convert_polygon_to_tiles(zoom, coord_list)


def diff(uvr1, uvr2):
  """Find the first difference between two UVRs.

  UVRs differing only in timestamp are considered identical.

  Args:
    uvr1: Uvr instance
    uvr2: Uvr instance

  Returns:
    str1: uvr1's value that differs from uvr2
    str2: uvr2's value that differs from uvr1
  """
  return _diff(uvr1._core, uvr2._core)


def _diff(d1, d2):
  """Find the first difference between two provided dicts.

  Args:
    d1: dict data structure
    d2: dict data structure

  Returns:
    str1: d1's value that differs from d2
    str2: d2's value that differs from d1
  """
  for key in sorted(d1):
    if key not in d2:
      return key, '<missing>'

  for key in sorted(d2):
    if key not in d1:
      return '<missing>', key

  for key in sorted(d1):
    v1 = d1[key]
    v2 = d2[key]
    if isinstance(v1, dict):
      if isinstance(v2, dict):
        diff = _diff(v1, v2)
        if diff:
          return key + '.' + diff[0], key + '.' + diff[1]
      else:
        return key + '=<dict>', key + '=' + str(v2)
    else:
      if isinstance(v2, dict):
        return key + '=' + str(v1), key + '=<dict>'
      else:
        if v1 != v2:
          return key + '=' + str(v1), key + '=' + str(v2)

  return None


def _validate_uvr(unvalidated_uvr):
  """Returns a validated UVR dict or raises an error.

  Args:
    unvalidated_uvr: nested dict containing unvalidated UVR data structure

  Returns:
    Validated UVR dict data structure.

  Raises:
    KeyError: when a required part of the UVR data structure is missing
    ValueError: when an invalid value is found
    TypeError: when timestamps are provided improperly
    OverflowError: with invalid timestamps
  """
  for key in unvalidated_uvr:
    if key not in UVR_FIELDS:
      raise ValueError('Unsupported field "%s" found in UVR' % key)

  uvr = copy.deepcopy(unvalidated_uvr)
  if 'timestamp' in uvr:
    del uvr['timestamp']

  if not UUID_MATCHER.match(uvr['message_id']):
    raise ValueError('message_id "%s" is not a valid UUID')

  for key, values in UVR_ENUMS.iteritems():
    if uvr[key] not in values:
      raise ValueError('Unexpected UVR %s value "%s"; must be one of %s' %
                       (key, uvr[key], '|'.join(values)))

  for permitted_uas in uvr['permitted_uas']:
    if permitted_uas not in UVR_PERMITTED_UAS:
      raise ValueError('Unexpected UVR permitted_uas entry "%s"; must be one '
                       'of %s' % (permitted_uas,
                                  '|'.join(UVR_PERMITTED_UAS)))

  for gufi in uvr['permitted_gufis']:
    if not UUID_MATCHER.match(gufi):
      raise ValueError('permitted_gufi "%s" is not a valid UUID' % gufi)

  if uvr['geography']['type'] != 'Polygon':
    raise ValueError('Unexpected geography type "%s"; expected "Polygon"' %
                     uvr['geography']['type'])
  coordinates = uvr['geography']['coordinates']
  if not isinstance(coordinates, collections.Sized):
    raise ValueError('Expected list of Polygons in UVR geography coordinates')
  if len(coordinates) != 1:
    raise ValueError('Found %d UVR geography coordinates Polygons; expected '
                     '1' % len(coordinates))
  polygon = coordinates[0]
  if not isinstance(polygon, collections.Sized):
    raise ValueError('Expected list of coordinate pairs in UVR geography '
                     'coordinates Polygon')
  if len(polygon) < UVR_MIN_POLYGON_COORDS:
    raise ValueError('Found %d coordinates in UVR geography coordinates '
                     'Polygon; expected at least %d' %
                     (len(polygon), UVR_MIN_POLYGON_COORDS))
  if any([not isinstance(coords, collections.Sized) for coords in polygon]):
    raise ValueError('Expected all elements of UVR geography coordinates '
                     'Polygon to be coordinate pairs')
  if any([len(coords) != 2 for coords in polygon]):
    raise ValueError('All coordinates in UVR geography coordinates must '
                     'contain exactly 2 elements (longitude, latitude)')
  if not all(v[0] == v[1] for v in zip(polygon[0], polygon[-1])):
    raise ValueError('Last coordinate pair in UVR geography coordinates must '
                     'match first coordinate pair')

  time_begin = format_utils.parse_timestamp(uvr['effective_time_begin'])
  time_end = format_utils.parse_timestamp(uvr['effective_time_end'])
  if time_begin >= time_end:
    raise ValueError('effective_time_end must be after effective_time_begin')
  uvr['effective_time_begin'] = format_utils.format_ts(time_begin)
  uvr['effective_time_end'] = format_utils.format_ts(time_end)

  if 'actual_time_end' in uvr and uvr['actual_time_end']:
    raise ValueError('Already-ended UVRs may not be emplaced in the grid')

  for key in ['min_altitude', 'max_altitude']:
    value = uvr[key]['vertical_reference']
    if value not in UVR_VERTICAL_REFERENCES:
      raise ValueError('Unexpected UVR %s vertical_reference value "%s"; '
                       'must be one of %s' %
                       (key, value, UVR_VERTICAL_REFERENCES))
    value = uvr[key]['units_of_measure']
    if value not in UVR_UNITS_OF_MEASURE:
      raise ValueError('Unexpected UVR %s units_of_measure value "%s"; '
                       'must be one of %s' %
                       (key, value, UVR_UNITS_OF_MEASURE))
    if 'altitude_value' not in uvr[key]:
      raise KeyError('Missing altitude_value in UVR ' + key)

  return uvr
