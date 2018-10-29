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

import format_utils
import uvrs

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')


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
        origin: <FIMS|USS>
        originator_id: <uss_name or originator domain name>
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
                         PART_107|PART_101E|PART_107X|RADIO_LINE_OF_SIGHT>, ...],
        permitted_gufis: [<uuid>, ...]
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
    # Parse the metadata or create a new one if none
    if content:
      m = json.loads(content)
      self.version = m['version']
      self.timestamp = m['timestamp']
      self.operators = m['operators']
      self.uvrs = [uvrs.Uvr(j) for j in m['uvrs']]
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

  def to_json(self):
    return {
      'version': self.version,
      'timestamp': self.timestamp,
      'operators': self.operators,
      'uvrs': [uvr.to_json() for uvr in self.uvrs]
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
    Returns:
      true if valid, false if not
    """
    if operations is None:
      operations = []
    # Remove the existing operator, if any
    self.remove_operator(uss_id)
    try:
      earliest_operation = format_utils.parse_timestamp(earliest)
      latest_operation = format_utils.parse_timestamp(latest)
      if earliest_operation >= latest_operation:
        raise ValueError
    except (TypeError, ValueError, OverflowError):
      log.error('Invalid date format/values for operators %s, %s',
                earliest, latest)
      return False
    # validate the operations (if any)
    for oper in operations:
      oper['timestamp'] = format_utils.format_ts()
      oper['version'] = self.version
    # Now add the new record
    operator = {
      'uss': uss_id,
      'uss_baseurl': baseurl,
      'version': self.version,
      'timestamp': format_utils.format_ts(),
      'minimum_operation_timestamp': format_utils.format_ts(earliest_operation),
      'maximum_operation_timestamp': format_utils.format_ts(latest_operation),
      'announcement_level': str(announce),
      'operations': operations
    }
    self.operators.append(operator)
    self._update_grid_location(zoom, x, y)
    self.timestamp = format_utils.format_ts()
    return True

  def remove_operator(self, uss_id):
    num_operators = len(self.operators)
    self.version += 1
    # Remove the existing operator, if any
    self.operators[:] = [
      d for d in self.operators if d.get('uss').upper() != uss_id.upper()
    ]
    self.timestamp = format_utils.format_ts()
    return len(self.operators) == num_operators - 1

  def upsert_operation(self, uss_id, gufi, signature, begin, end):
    """Inserts or updates an operation, with gufi as the key.

    Args:
      uss_id: plain text identifier for the USS,
      gufi: Unique flight identifier per NASA formatting standards
      signature: The JWS signature of the Operation,
      begin: start time of the operation.
      end: end time of the operation.
    Returns:
      true if valid, false if it cannot find the USS
    """
    # clean up the datetimestamps, setting to nothing if invalid rather
    #   than failing, as they are optional
    found = False
    # Remove the existing operation, if any
    self.remove_operation(uss_id, gufi)
    try:
      effective_time_begin = format_utils.parse_timestamp(begin)
      effective_time_end = format_utils.parse_timestamp(end)
      if effective_time_begin >= effective_time_end:
        raise ValueError
    except (TypeError, ValueError, OverflowError):
      log.error('Invalid date format/values for operators %s, %s',
                begin, end)
      return False
    # Now add the new record
    operation = {
      'version': self.version,
      'gufi': gufi,
      'operation_signature': signature,
      'effective_time_begin': format_utils.format_ts(effective_time_begin),
      'effective_time_end': format_utils.format_ts(effective_time_end),
      'timestamp': format_utils.format_ts()
    }
    # find the operator entry and add the operation
    for oper in self.operators:
      if oper.get('uss').upper() == uss_id.upper():
        found = True
        oper['version'] = self.version
        oper['timestamp'] = format_utils.format_ts()
        oper['operations'].append(operation)
        break
      self.timestamp = format_utils.format_ts()
    return found


  def remove_operation(self, uss_id, gufi):
    found = False
    self.version += 1
    # find the operator entry
    for oper in self.operators:
      if oper.get('uss').upper() == uss_id.upper():
        # Remove the existing operation, if any
        num_operations = len(oper['operations'])
        oper['operations'][:] = [
          d for d in oper['operations'] if d.get('gufi').upper() != gufi.upper()
        ]
        found = (len(oper['operations']) == num_operations - 1)
        oper['version'] = self.version
        oper['timestamp'] = format_utils.format_ts()
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
