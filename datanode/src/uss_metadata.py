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
import datetime
import json
import logging
import pytz
from dateutil import parser

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')


def _format_endpoint(e, z, x, y):
  """Format a raw endpoint by replacing placeholders with real values.

  Args:
    e: Prototype endpoint to format with real values.
    z: Value to substitute for {zoom} or {z}.
    x: Value to substitute for {x}.
    y: Value to substitute for {y}.

  Returns:
    Formatted endpoint.
  """
  e = e.replace('{zoom}', str(z)).replace('{z}', str(z))
  e = e.replace('{x}', str(x))
  e = e.replace('{y}', str(y))
  return e


class USSMetadata(object):
  """Data structure for the USS metadata stored for one or more GridCells.

  Format: {version: <version>, timestamp: <last_updated>, operators:
    [{uss: <ussid>, scope: <used_for_obtaining_oauth_tokens>,
    version: <last_version_for_this_uss>, timestamp: <last_updated>,
    operation_endpoint: <endpoint_to_retrieve_operations_in_this_grid>,
    operation_format: <output_format_of_uas_operations>,
    minimum_operation_timestamp: <lowest_start_time_of_operations_in_this_cell>,
    maximum_operation_timestamp: <highest_end_time_of_operations_in_this_cell>,
    zoom: <slippy_zoom_level_for_the_gridcell>,
    x: <slippy_x_level_for_the_gridcell>,
    y: <slippy_y_level_for_the_gridcell>,
    },
      ...other USSs as appropriate... ]
  }

  """

  def __init__(self, content=None):
    # Parse the metadata or create a new one if none
    if content:
      m = json.loads(content)
      self.version = m['version']
      self.timestamp = m['timestamp']
      self.operators = m['operators']
    else:
      self.version = 0
      self.timestamp = self.format_ts()
      self.operators = []

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
      'operators': self.operators
    }

  def upsert_operator(self, uss_id, ws_scope, operation_format, operation_ws,
    earliest, latest, public_portal_endoint, flight_info_endpoint, zoom, x, y):
    """Inserts or updates an operation, with uss_id as the key.

    Args:
      uss_id: plain text identifier for the USS,
      ws_scope: scope to use to obtain OAuth token,
      operation_format: output format for operation ws (i.e. NASA, GUTMA),
      operation_ws: submitting USS endpoint where all flights in
        this cell can be retrieved from,
      earliest: lower bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      latest: upper bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      public_portal_endpoint: Submitting USS web service endpoint where all
        public flight remote identification telemetry in this cell can be
        retrieved.
      flight_info_endpoint: Submitting USS web service endpoint where a public
        flight's remote identification details can be retrieved.
      zoom, x, y: grid reference for this cell
    Returns:
      true if valid, false if not
    """
    # Remove the existing operator, if any
    self.remove_operator(uss_id)
    try:
      earliest_operation = parser.parse(earliest)
      if earliest_operation.tzinfo is None:
        earliest_operation = earliest_operation.replace(tzinfo=pytz.utc)
      latest_operation = parser.parse(latest)
      if latest_operation.tzinfo is None:
        latest_operation = latest_operation.replace(tzinfo=pytz.utc)
      if earliest_operation >= latest_operation:
        raise ValueError
    except (TypeError, ValueError, OverflowError):
      log.error('Invalid date format/values for operators %s, %s',
                earliest, latest)
      return False
    # Now add the new record
    operator = {
      'uss': uss_id,
      'scope': ws_scope,
      'version': self.version,
      'timestamp': self.format_ts(),
      'operation_endpoint': operation_ws,
      'operation_format': operation_format,
      'minimum_operation_timestamp': self.format_ts(earliest_operation),
      'maximum_operation_timestamp': self.format_ts(latest_operation),
      'public_portal_endpoint': public_portal_endoint,
      'flight_info_endpoint': flight_info_endpoint,
    }
    self.operators.append(operator)
    self._update_grid_location(zoom, x, y)
    self.timestamp = self.format_ts()
    return True

  def remove_operator(self, uss_id):
    self.version += 1
    # Remove the existing operator, if any
    self.operators[:] = [
      d for d in self.operators if d.get('uss').upper() != uss_id.upper()
    ]
    self.timestamp = self.format_ts()

  def format_ts(self, timestamp=None):
    r = datetime.datetime.now(pytz.utc) if timestamp is None else timestamp
    r = r.astimezone(pytz.utc)
    return r.isoformat()

  def _update_grid_location(self, z, x, y):
    """Updates the z, x, y fields and any variables in the endpoints"""
    if z is None or x is None or y is None:
      raise ValueError('Slippy values not set for grid location')
    else:
      for operator in self.operators:
        operator['zoom'] = z
        operator['x'] = x
        operator['y'] = y
        operator['operation_endpoint'] = _format_endpoint(
          operator['operation_endpoint'], z, x, y)
        operator['public_portal_endpoint'] = _format_endpoint(
          operator['public_portal_endpoint'], z, x, y)
        operator['flight_info_endpoint'] = _format_endpoint(
          operator['flight_info_endpoint'], z, x, y)
