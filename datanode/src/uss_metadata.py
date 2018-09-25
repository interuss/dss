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
    [{uss: <ussid>, uss_baseurl: <base_url_for_NASA_API>,
    version: <last_version_for_this_uss>, timestamp: <last_updated>,
    announcement_level: <flag_for_requesting_announcements_from _other_uss>,
    minimum_operation_timestamp: <lowest_start_time_of_operations_in_this_cell>,
    maximum_operation_timestamp: <highest_end_time_of_operations_in_this_cell>,
    operations: [{version: <version>, gufi: <unique_identifier>,
      operation_signature: <jws_signature_for_operation>,
      effective_time_begin: <operation_start_time>,
      effective_time_end: <operation_end_time>,
      timestamp: <last_updated> }
      ...other operations as appropriate... ]
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

  def upsert_operator(self, uss_id, baseurl, announce,
      earliest, latest, operations=None):
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
    # validate the operations (if any)
    for oper in operations:
      oper['timestamp'] = self.format_ts()
      oper['version'] = self.version
    # Now add the new record
    operator = {
      'uss': uss_id,
      'uss_baseurl': baseurl,
      'version': self.version,
      'timestamp': self.format_ts(),
      'minimum_operation_timestamp': self.format_ts(earliest_operation),
      'maximum_operation_timestamp': self.format_ts(latest_operation),
      'announcement_level': str(announce),
      'operations': operations
    }
    self.operators.append(operator)
    self.timestamp = self.format_ts()
    return True

  def remove_operator(self, uss_id):
    num_operators = len(self.operators)
    self.version += 1
    # Remove the existing operator, if any
    self.operators[:] = [
      d for d in self.operators if d.get('uss').upper() != uss_id.upper()
    ]
    self.timestamp = self.format_ts()
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
      effective_time_begin = parser.parse(begin)
      if effective_time_begin.tzinfo is None:
        effective_time_begin = effective_time_begin.replace(tzinfo=pytz.utc)
      effective_time_end = parser.parse(end)
      if effective_time_end.tzinfo is None:
        effective_time_end = effective_time_end.replace(tzinfo=pytz.utc)
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
      'effective_time_begin': self.format_ts(effective_time_begin),
      'effective_time_end': self.format_ts(effective_time_end),
      'timestamp': self.format_ts()
    }
    # find the operator entry and add the operation
    for oper in self.operators:
      if oper.get('uss').upper() == uss_id.upper():
        found = True
        oper['version'] = self.version
        oper['timestamp'] = self.format_ts()
        oper['operations'].append(operation)
        break
      self.timestamp = self.format_ts()
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
        oper['timestamp'] = self.format_ts()
    return found

  def format_ts(self, timestamp=None):
    r = datetime.datetime.now(pytz.utc) if timestamp is None else timestamp
    r = r.astimezone(pytz.utc)
    return '{0}Z'.format(r.strftime('%Y-%m-%dT%H:%M:%S.%f')[:23])