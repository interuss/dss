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
import datetime
import json
import logging
import pytz
from dateutil import parser

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')


class USSMetadata(object):
  """Data structure for the metadata stored for USS entries in a GridCell.

  Format: {version: <version>, timestamp: <last_updated>, operators:
    [{uss: <ussid>, scope: <used_for_obtaining_oauth_tokens>,
    version: <last_version_for_this_uss>, timestamp: <last_updated>,
    operation_endpoint: <endpoint_to_retrieve_operations_in_this_grid>,
    operation_format: <output_format_of_uas_operations>,
    minimum_operation_timestamp: <lowest_start_time_of_operations_in_this_cell>,
    maximum_operation_timestamp: <highest_end_time_of_operations_in_this_cell>
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

  def to_json(self):
    return {
      'version': self.version,
      'timestamp': self.timestamp,
      'operators': self.operators
    }

  def upsert_operator(self, uss_id, ws_scope, operation_format, operation_ws,
    earliest, latest):
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
      'maximum_operation_timestamp': self.format_ts(latest_operation)
    }
    self.operators.append(operator)
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
