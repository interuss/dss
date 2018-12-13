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

This module is the information wrapper for the actual JSON data structure.


Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the 'License');
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an 'AS IS' BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""
import copy
import json
import logging
from dateutil import parser
import re

import format_utils
import uvrs

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')

GUFI_VALIDATOR = re.compile('^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[8-b][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$')


class USSMetadata(object):
  """Data structure for the metadata stored for USS entries in a GridCell.

  Format:
  {
    version: <version>,
    timestamp: <last_updated>,
    operators: [
      {
        uss: <ussid>,
        uss_baseurl: <base_url_for_NASA_API>,
        version: <last_version_for_this_uss>,
        timestamp: <last_updated>,
        announcement_level: <flag_for_requesting_announcements_from _other_uss>,
        minimum_operation_timestamp: <lowest_start_time_of_operations_in_this_cell>,
        maximum_operation_timestamp: <highest_end_time_of_operations_in_this_cell>,
        zoom: <slippy_zoom_level_for_the_gridcell>,
        x: <slippy_x_level_for_the_gridcell>,
        y: <slippy_y_level_for_the_gridcell>,
        operations: [
          {
            version: <version>,
            gufi: <unique_identifier>,
            operation_signature: <jws_signature_for_operation>,
            effective_time_begin: <operation_start_time>,
            effective_time_end: <operation_end_time>,
            timestamp: <last_updated> }
          }, ...other operations as appropriate... ]
      }, ...other operator USSs as appropriate...],
    uvrs: [
      {
        message_id: <uuid>
        uss_name: <USS ID matching access token>
        type: <DYNAMIC_RESTRICTION|STATIC_ADVISORY>
        cause: <WEATHER|ATC|SECURITY|SAFETY|MUNICIPALITY>
        geography:
          {
            type: Polygon
            coordinates: [[[<longitude>, <latitude>],
              ...other polygon vertices as appropriate ...]]
          }
        effective_time_begin: <YYYY-MM-DDTHH:mm:ss.fffZ>
        effective_time_end: <YYYY-MM-DDTHH:mm:ss.fffZ>
        permitted_uas: [<NOT_SET|PUBLIC_SAFETY|SECURITY||NEWS_GATHERING|VLOS|
                         SUPPORT_LEVEL|PART_107|PART_101E|PART_107X|
                         RADIO_LINE_OF_SIGHT>, ...],
        permitted_gufis: [<uuid>, ...]
        required_support: [<V2V|DAA|ADSB_OUT|ADSB_IN|CONSPICUITY|
                            ENHANCED_NAVIGATION|ENHANCED_SAFE_LANDING>, ...]
        actual_time_end: <YYYY-MM-DDTHH:mm:ss.fffZ>
        min_altitude: {altitude_value: <altitude>, vertical_reference: "W84",
                       units_of_measure: "FT"}
        max_altitude: {altitude_value: <altitude>, vertical_reference: "W84",
                       units_of_measure: "FT"}
        reason: <human-readable string>
      }, ...other UVRs as appropriate... ]
  }

  """

  def __init__(self, content=None):
    """Parse metadata in storage format or create a new one if no content.

    Args:
      content: String containing JSON dict with metadata information.
    """
    #
    if content:
      m = json.loads(content)
      self.version = m['version']
      self.timestamp = m['timestamp']
      self.operators = m['operators']
      if 'uvrs' in m:
        self.uvrs = [uvrs.Uvr(j, True) for j in m['uvrs']]
      else:
        self.uvrs = []
    else:
      self.version = 0
      self.timestamp = format_utils.format_ts()
      self.operators = []
      self.uvrs = []

  def __str__(self):
    return str(self.to_json())

  def __add__(self, other):
    """Adds two metadata objects together"""
    combined = copy.deepcopy(self)
    if other is not None:
      if (parser.parse(other.timestamp) > parser.parse(combined.timestamp) or
        combined.version == 0):
        combined.version = other.version
        combined.timestamp = other.timestamp
      for operator in other.operators:
        combined.operators.append(operator)
      my_uvrs = {c['message_id']: c for c in self.uvrs}
      for uvr in other.uvrs:
        if uvr['message_id'] in my_uvrs:
          if uvr != my_uvrs[uvr['message_id']]:
            raise ValueError(('Conflicting UVRs with message_id %s' %
                              uvr['message_id']))
        else:
          combined.uvrs.append(uvr)
    keys = []
    for o in combined.operators:
      key = (o['uss'], o['zoom'], o['x'], o['y'])
      if key in keys:
        raise ValueError('Duplicate USS ID, zoom, x, y found during add')
      else:
        keys.append(key)
    return combined

  def __radd__(self, other):
    if other == 0:
      return self
    else:
      return self.__add__(other)

  def to_json(self, storage=False):
    """Convert this USSMetadata object into a plain JSON dict.

    Args:
      storage: If True, return a JSON dict suitable for internal grid storage.
        Otherwise, return an API-compatible JSON dict.

    Returns:
      JSON dict with content from this USSMetadata object.
    """
    return {
      'version': self.version,
      'timestamp': self.timestamp,
      'operators': self.operators,
      'uvrs': [uvr.to_json(storage) for uvr in self.uvrs]
    }

  def upsert_operator(self, uss_id, baseurl, announce,
      earliest, latest, zoom, x, y, operations=None):
    """Inserts or updates an operator, with uss_id as the key.

    Args:
      uss_id: plain text identifier for the USS,
      baseurl: Base URL for the USSs web service endpoints hosting the
        required NASA API (https://app.swaggerhub.com/apis/utm/uss/).
      announce: The level of announcements the USS would like to receive related
        to operations in this grid cell. Current just a binary, but expect this
        enumeration to grow as use cases are developed. For example, USSs may
        want just security related announcements, or would only like
        announcements that involve changed geographies.
      earliest: lower bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      latest: upper bound of active or planned flight timestamp,
      zoom, x, y: grid reference for this cell,
      operations: complete list of operations for this operator

        used for quick filtering conflicts.
    Raises:
      ValueError: when input parameters are invalid.
    """
    if operations is None:
      operations = []

    # Remove the existing operator, if any, and increment version
    self.remove_operator(uss_id)

    # Validate earliest and latest timestamps
    try:
      earliest_operation = format_utils.parse_timestamp(earliest)
      latest_operation = format_utils.parse_timestamp(latest)
      if earliest_operation >= latest_operation:
        raise ValueError()
    except (TypeError, ValueError, OverflowError):
      msg = 'Invalid date format/values for operator %s, %s' % (earliest,
                                                                latest)
      log.error(msg)
      raise ValueError(msg)

    # Add the new record without operations
    operator = {
      'uss': uss_id,
      'uss_baseurl': baseurl,
      'version': self.version,
      'timestamp': format_utils.format_ts(),
      'minimum_operation_timestamp': format_utils.format_ts(earliest_operation),
      'maximum_operation_timestamp': format_utils.format_ts(latest_operation),
      'announcement_level': str(announce),
      'operations': []
    }
    self.operators.append(operator)

    # Insert the operations
    for operation in operations:
      for required_field in ('effective_time_begin', 'effective_time_end',
                             'gufi', 'operation_signature'):
        if required_field not in operation:
          raise ValueError('Operation missing ' + required_field)
      self.upsert_operation(uss_id, operation['gufi'],
                            operation['operation_signature'],
                            operation['effective_time_begin'],
                            operation['effective_time_end'],
                            same_version=True, expand_operator_window=False)

    # Check for duplicate GUFIs
    for i in range(len(operations) - 1):
      gufi1 = operations[i]['gufi']
      for j in range(i + 1, len(operations)):
        if _gufis_match(gufi1, operations[j]['gufi']):
          raise ValueError('Duplicate GUFI found: ' + gufi1)

    self._update_grid_location(zoom, x, y)
    self.timestamp = format_utils.format_ts()

  def remove_operator(self, uss_id):
    num_operators = len(self.operators)
    self.version += 1
    # Remove the existing operator, if any
    self.operators[:] = [
      d for d in self.operators if d.get('uss').upper() != uss_id.upper()
    ]
    self.timestamp = format_utils.format_ts()
    return len(self.operators) == num_operators - 1

  def upsert_operation(self, uss_id, gufi, signature, begin, end,
                       same_version=False, expand_operator_window=True):
    """Inserts or updates an operation, with gufi as the key.

    Args:
      uss_id: plain text identifier for the USS,
      gufi: Unique flight identifier per NASA formatting standards
      signature: The JWS signature of the Operation,
      begin: start time of the operation.
      end: end time of the operation.
      same_version: True to avoid incrementing the metadata version.
      expand_operator_window: True to expand min and max operation timestamps in
        operator according to this operation, False to raise an error when
        operation is outside operator's time window.
    Returns:
      True if successful, False if it cannot find the Operator.
    Raises:
      ValueError: when inputs are invalid.
    """
    if not same_version:
      self.version += 1

    # Validate GUFI
    if not GUFI_VALIDATOR.match(gufi):
      raise ValueError('Invalid GUFI format: ' + gufi)

    # Find the operator entry
    found = False
    for operator in self.operators:
      if operator.get('uss').upper() == uss_id.upper():
        found = True
        break
      self.timestamp = format_utils.format_ts()
    if not found:
      return False

    operator['version'] = self.version
    operator['timestamp'] = format_utils.format_ts()
    operator['operations'] = [op for op in operator['operations']
                              if not _gufis_match(gufi, op['gufi'])]

    try:
      effective_time_begin = format_utils.parse_timestamp(begin)
      effective_time_end = format_utils.parse_timestamp(end)
    except (TypeError, ValueError, OverflowError):
      msg = 'Invalid date format/values for operation %s, %s' % (begin, end)
      log.error(msg)
      raise ValueError(msg)

    if effective_time_begin >= effective_time_end:
      raise ValueError('Operation ends before it starts')

    earliest_operation = format_utils.parse_timestamp(
      operator['minimum_operation_timestamp'])
    latest_operation = format_utils.parse_timestamp(
      operator['maximum_operation_timestamp'])
    if effective_time_begin < earliest_operation:
      if expand_operator_window:
        earliest_operation = effective_time_begin
        operator['minimum_operation_timestamp'] = format_utils.format_ts(
          earliest_operation)
      else:
        raise ValueError('Operation begins before minimum operation timestamp')
    if effective_time_end > latest_operation:
      if expand_operator_window:
        latest_operation = effective_time_end
        operator['maximum_operation_timestamp'] = format_utils.format_ts(
          latest_operation)
      else:
        raise ValueError('Operation ends after maximum operation timestamp')

    # Now add the new operation
    operation = {
      'version': self.version,
      'gufi': gufi,
      'operation_signature': signature,
      'effective_time_begin': format_utils.format_ts(effective_time_begin),
      'effective_time_end': format_utils.format_ts(effective_time_end),
      'timestamp': format_utils.format_ts(),
    }
    operator['operations'].append(operation)
    return True


  def remove_operation(self, uss_id, gufi):
    found = False
    self.version += 1
    # find the operator entry
    for operator in self.operators:
      if operator.get('uss').upper() == uss_id.upper():
        # Remove the existing operation, if any
        num_operations = len(operator['operations'])
        operator['operations'][:] = [d for d in operator['operations']
                                     if not _gufis_match(d.get('gufi'), gufi)]
        found = len(operator['operations']) < num_operations
        operator['version'] = self.version
        operator['timestamp'] = format_utils.format_ts()
    return found

  def insert_uvr(self, uvr):
    """Inserts a UVR, failing if another UVR with the same ID already exists.

    Args:
      uvr: UVR instance

    Raises:
      ValueError: if a UVR already exists with the new UVR's message_id
    """

    if uvr['message_id'] in [existing['message_id'] for existing in self.uvrs]:
      raise ValueError('UVR with message_id %s already exists; it must be '
                       'removed before inserting a new UVR with that ID' %
                       (uvr['message_id']))

    # Now add the new record
    uvr['timestamp'] = format_utils.format_ts()
    self.uvrs.append(uvr)
    self.version += 1

  def remove_uvr(self, message_id):
    """Removes a UVR, if it exists.  Increments version regardless.

    Args:
      message_id: ID of UVR to remove

    Returns:
      Removed UVR, or None if no matching UVR was found.
    """
    self.version += 1
    # find the UVR entry
    for existing_uvr in self.uvrs:
      if existing_uvr['message_id'] == message_id:
        self.uvrs.remove(existing_uvr)
        return existing_uvr
    return None

  def _update_grid_location(self, z, x, y):
    """Updates the z, x, y fields and any variables in the endpoints"""
    if z is None or x is None or y is None:
      raise ValueError('Slippy values not set for grid location')
    else:
      for operator in self.operators:
        operator['zoom'] = z
        operator['x'] = x
        operator['y'] = y
        e = operator['uss_baseurl']
        e = e.replace('{zoom}', str(z)).replace('{z}', str(z))
        e = e.replace('{x}', str(x))
        e = e.replace('{y}', str(y))
        operator['uss_baseurl'] = e


def _gufis_match(gufi1, gufi2):
  """Indicates whether two GUFIs match.

  Args:
    gufi1: String containing GUFI.
    gufi2: String containing GUFI.

  Returns:
    True if GUFIs match, False otherwise.
  """
  return gufi1.upper() == gufi2.upper()
