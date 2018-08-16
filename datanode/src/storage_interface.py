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

This module is the information interface to Zookeeper.


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
from dateutil import parser

# Kazoo is the zookeeper wrapper for python
from kazoo.client import KazooClient
from kazoo.exceptions import LockTimeout
from kazoo.handlers.threading import KazooTimeoutError
from kazoo.protocol.states import KazooState

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')

# CONSTANTS
# Lock stores in this format /uss/gridcells/{z}/{x}/{y}/manifest
USS_BASE_PREFIX = '/uss/gridcells/'
TEST_BASE_PREFIX = '/test/'
USS_METADATA_FILE = '/manifest'
BAD_CHARACTER_CHECK = '\';(){}[]!@#$%^&*|"<>'
CONNECTION_TIMEOUT = 5.0  # seconds
LOCK_TIMEOUT = 5.0  # seconds
DEFAULT_CONNECTION = 'localhost:2181'
GRID_PATH = USS_BASE_PREFIX


class USSMetadataManager(object):
  """Interfaces with the locking system to get, put, and delete USS metadata.

  Metadata gets/stores/deletes the USS information for a partiular grid,
  including current version number, a list of USSs with active operations,
  and the endpoints to get that information. Locking is assured through a
  snapshot token received when getting, and used when putting.
  """

  def __init__(self, connectionstring=DEFAULT_CONNECTION, testgroupid=None):
    """Initializes the class.

    Args:
      connectionstring:
        Zookeeper connection string - server:port,server:port,...
      testgroupid:
        ID to use if in test mode, none for normal mode
    """
    if testgroupid:
      self.set_testmode(testgroupid)
    if not connectionstring:
      connectionstring = DEFAULT_CONNECTION
    log.debug('Creating metadata manager object and connecting to zookeeper...')
    try:
      if set(BAD_CHARACTER_CHECK) & set(connectionstring):
        raise ValueError
      self.zk = KazooClient(hosts=connectionstring, timeout=CONNECTION_TIMEOUT)
      self.zk.add_listener(self.zookeeper_connection_listener)
      self.zk.start()
      if testgroupid:
        self.delete_testdata(testgroupid)
    except KazooTimeoutError:
      log.error('Unable to connect to zookeeper using %s connection string...',
                connectionstring)
      raise
    except ValueError:
      log.error('Connection string %s seems invalid...', connectionstring)
      raise

  def __del__(self):
    log.debug('Destroying metadata manager object and disconnecting from zk...')
    self.zk.stop()

  def set_verbose(self):
    log.setLevel(logging.DEBUG)

  def set_testmode(self, testgroupid='UNDEFINED_TESTER'):
    """Sets the mode to testing with the specific test ID, cannot be undone.

    Args:
      testgroupid: ID to use if in test mode, none for normal mode
    """
    global GRID_PATH
    global CONNECTION_TIMEOUT
    # Adjust parameters specifically for the test
    GRID_PATH = TEST_BASE_PREFIX + testgroupid + USS_BASE_PREFIX
    log.debug('Setting test path to %s...', GRID_PATH)
    CONNECTION_TIMEOUT = 1.0

  def zookeeper_connection_listener(self, state):
    if state == KazooState.LOST:
      # Register somewhere that the session was lost
      log.error('Lost connection with the zookeeper servers...')
    elif state == KazooState.SUSPENDED:
      # Handle being disconnected from Zookeeper
      log.error('Suspended connection with the zookeeper servers...')
    elif state == KazooState.CONNECTED:
      # Handle being connected/reconnected to Zookeeper
      log.info('Connection restored with the zookeeper servers...')

  def delete_testdata(self, testgroupid=None):
    """Removes the test data from the servers.

    Be careful when using this in parallel as it removes everything under
    the testgroupid, or everything if no tetgroupid is provided.

    Args:
      testgroupid: ID to use if in test mode, none will remove all test data
    """
    if testgroupid:
      path = TEST_BASE_PREFIX + testgroupid
    else:
      path = TEST_BASE_PREFIX
    self.zk.delete(path, recursive=True)

  def get(self, z, x, y):
    """Gets the metadata and snapshot token for a GridCell.

    Reads data from zookeeper, including a snapshot token. The
    snapshot token is used as a reference when writing to ensure
    the data has not been updated between read and write.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    # TODO(hikevin): Change to use our own error codes and let the server
    #                   convert them to http error codes. For now, this is
    #                   at least in a standard JSend format.
    status = 500
    result = None
    if self._validate_slippy(z, x, y):
      (content, metadata) = self._get_raw(z, x, y)
      if metadata:
        try:
          m = USSMetadata(content)
          status = 200
          result = {
              'status': 'success',
              'sync_token': metadata.last_modified_transaction_id,
              'data': m.to_json()
          }
        except ValueError:
          status = 424
      else:
        status = 404
    else:
      status = 400
    if status != 200:
      result = self._format_status_code_to_jsend(status)
    return result

  def set(self, z, x, y, sync_token, uss_id, baseurl, announce,
          earliest_operation, latest_operation, operations=None):
    """Sets the metadata for a GridCell.

    Writes data, using the snapshot token for confirming data
    has not been updated since it was last read.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      sync_token: token retrieved in the original GET GridCellMetadata,
      uss_id: plain text identifier for the USS,
      baseurl: Base URL for the USSs web service endpoints hosting the
        required NASA API (https://app.swaggerhub.com/apis/utm/uss/).
      announce: The level of announcements the USS would like to receive related
        to operations in this grid cell. Current just a binary, but expect this
        enumeration to grow as use cases are developed. For example, USSs may
        want just security related announcements, or would only like
        announcements that involve changed geographies.
      earliest_operation: lower bound of active or planned flight timestamp,
      latest_operation: upper bound of active or planned flight timestamp,
        dates are used for quick filtering conflicts.
      operations: complete list of operations for this operator
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    status = 500
    if operations is None:
      operations = []
    if self._validate_slippy(z, x, y):
      # first we have to get the cell
      status = 0
      (content, metadata) = self._get_raw(z, x, y)
      if metadata:
        # Quick check of the token, another is done on the actual set to be sure
        #    but this check fails early and fast
        if str(metadata.last_modified_transaction_id) == str(sync_token):
          try:
            m = USSMetadata(content)
            log.debug('Setting metadata for %s...', uss_id)
            if not m.upsert_operator(uss_id, baseurl, announce,
                                     earliest_operation, latest_operation,
                                     operations):
              log.error('Failed setting operator for %s with token %s...',
                        uss_id, str(sync_token))
              raise ValueError
            status = self._set_raw(z, x, y, m, uss_id, sync_token)
          except ValueError:
            status = 424
        else:
          status = 409
      else:
        status = 404
    else:
      status = 400
    if status == 200:
      # Success, now get the metadata back to send back
      result = self.get(z, x, y)
    else:
      result = self._format_status_code_to_jsend(status)
    return result

  def delete(self, z, x, y, uss_id):
    """Sets the metadata for a GridCell by removing the entry for the USS.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      uss_id: is the plain text identifier for the USS
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    status = 500
    m = None
    if self._validate_slippy(z, x, y):
      # first we have to get the cell
      (content, metadata) = self._get_raw(z, x, y)
      if metadata:
        try:
          m = USSMetadata(content)
          if m.remove_operator(uss_id):
            # TODO(pelletierb): Automatically retry on delete
            status = self._set_raw(z, x, y, m, uss_id,
                                   metadata.last_modified_transaction_id)
          else:
            status = 404
        except ValueError:
          status = 424
      else:
        status = 404
    else:
      status = 400
    if status == 200:
      # Success, now get the metadata back to send back
      (content, metadata) = self._get_raw(z, x, y)
      result = {
          'status': 'success',
          'sync_token': metadata.last_modified_transaction_id,
          'data': m.to_json()
      }
    else:
      result = self._format_status_code_to_jsend(status)
    return result

  def delete_operation(self, z, x, y, uss_id, gufi):
    """Removes an operation from an operator.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      uss_id: is the plain text identifier for the USS
      gufi: Unique flight identifier per NASA formatting standards
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    status = 500
    m = None
    if self._validate_slippy(z, x, y):
      # first we have to get the cell
      (content, metadata) = self._get_raw(z, x, y)
      if metadata:
        try:
          m = USSMetadata(content)
          if m.remove_operation(uss_id, gufi):
            # TODO(pelletierb): Automatically retry on delete
            status = self._set_raw(z, x, y, m, uss_id,
                                   metadata.last_modified_transaction_id)
          else:
            status = 404
        except ValueError:
          status = 424
      else:
        status = 404
    else:
      status = 400
    if status == 200:
      # Success, now get the metadata back to send back
      (content, metadata) = self._get_raw(z, x, y)
      result = {
        'status': 'success',
        'sync_token': metadata.last_modified_transaction_id,
        'data': m.to_json()
      }
    else:
      result = self._format_status_code_to_jsend(status)
    return result

  ######################################################################
  ################       INTERNAL FUNCTIONS    #########################
  ######################################################################
  def _get_raw(self, z, x, y):
    """Gets the raw content and metadata for a GridCell from zookeeper.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
    Returns:
      content: USS metadata
      metadata: straight from zookeeper
    """
    path = GRID_PATH + '/'.join((str(z), str(x), str(y))) + USS_METADATA_FILE
    log.debug('Getting metadata from zookeeper@%s...', path)
    self.zk.ensure_path(path)
    c, m = self.zk.get(path)
    if c:
      log.debug('Received raw content and metadata from zookeeper: %s', c)
    if m:
      log.debug('Received raw metadata from zookeeper: %s', m)
    return c, m

  def _set_raw(self, z, x, y, m, uss_id, sync_token):
    """Grabs the lock and updates the raw content for a GridCell in zookeeper.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      m: metadata object to write
      uss_id: the plain text identifier for the USS
      sync_token: the sync token received during get operation
    Returns:
      200 for success, 409 for conflict, 408 for unable to get the lock
    """
    status = 500
    path = GRID_PATH + '/'.join((str(z), str(x), str(y))) + USS_METADATA_FILE
    # TODO(hikevin): Remove Lock and use built in set with version
    lock = self.zk.WriteLock(path, uss_id)
    try:
      log.debug('Getting metadata lock from zookeeper@%s...', path)
      lock.acquire(timeout=LOCK_TIMEOUT)
      (content, metadata) = self._get_raw(z, x, y)
      del content
      if str(metadata.last_modified_transaction_id) == str(sync_token):
        log.debug('Setting metadata to %s...', str(m))
        self.zk.set(path, json.dumps(m.to_json()))
        status = 200
      else:
        log.error(
            'Sync token from USS (%s) does not match token from zk (%s)...',
            str(sync_token), str(metadata.last_modified_transaction_id))
        status = 409
      log.debug('Releasing the lock...')
      lock.release()
    except LockTimeout:
      log.error('Unable to acquire the lock for %s...', path)
      status = 408
    return status

  def _format_status_code_to_jsend(self, status):
    """Formats a response based on HTTP status code.

    Args:
      status: HTTP status code
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """

    if status == 200 or status == 204:
      result = {'status': 'success', 'code': 204, 'message': 'Empty data set.'}
    elif status == 400:
      result = {
          'status': 'fail',
          'code': status,
          'message': 'Parameters are not following the correct format.'
      }
    elif status == 404:
      result = {
          'status': 'fail',
          'code': status,
          'message': 'Unable to pull metadata from lock system.'
      }
    elif status == 408:
      result = {
          'status': 'fail',
          'code': status,
          'message': 'Timeout trying to get lock.'
      }
    elif status == 409:
      result = {
          'status':
              'fail',
          'code':
              status,
          'message':
              'Content in metadata has been updated since provided sync token.'
      }
    elif status == 424:
      result = {
          'status':
              'fail',
          'code':
              status,
          'message':
              'Content in metadata is not following JSON format guidelines.'
      }
    else:
      result = {
          'status': 'fail',
          'code': status,
          'message': 'Unknown error code occurred.'
      }
    return result

  def _validate_slippy(self, z, x, y):
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
        raise ValueError
      if not 0 <= x < 2**z:
        raise ValueError
      if not 0 <= y < 2**z:
        raise ValueError
      return True
    except (ValueError, TypeError):
      log.error('Invalid slippy format for tiles %sz, %s,%s!', z, x, y)
      return False


class USSMetadata(object):
  """Data structure for the metadata stored for USS entries in a GridCell.

  Format: {version: <version>, timestamp: <last_updated>, operators:
    [{uss: <ussid>, baseurl: <base_url_for_NASA_API>,
    version: <last_version_for_this_uss>, timestamp: <last_updated>,
    announce: <flag_for_requesting_announcements_from _other_uss>,
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
      self.timestamp = datetime.datetime.now().isoformat()
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
                      earliest_operation, latest_operation, operations=None):
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
      earliest_operation: lower bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      latest_operation: upper bound of active or planned flight timestamp,
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
      earliest_operation = parser.parse(earliest_operation)
      latest_operation = parser.parse(latest_operation)
      if earliest_operation >= latest_operation:
        raise ValueError
    except (TypeError, ValueError, OverflowError):
      log.error('Invalid date format/values for operators %s, %s',
                earliest_operation, latest_operation)
      return False
    # validate the operations (if any)
    for oper in operations:
      oper['timestamp'] = datetime.datetime.now().isoformat()
      oper['version'] = self.version
    # Now add the new record
    operator = {
        'uss': uss_id,
        'uss_baseurl': baseurl,
        'version': self.version,
        'timestamp': datetime.datetime.now().isoformat(),
        'minimum_operation_timestamp': earliest_operation.isoformat(),
        'maximum_operation_timestamp': latest_operation.isoformat(),
        'announcement_level': announce,
        'operations': operations
    }
    self.operators.append(operator)
    self.timestamp = datetime.datetime.now().isoformat()
    return True

  def remove_operator(self, uss_id):
    num_operators = len(self.operators)
    self.version += 1
    # Remove the existing operator, if any
    self.operators[:] = [
        d for d in self.operators if d.get('uss').upper() != uss_id.upper()
    ]
    self.timestamp = datetime.datetime.now().isoformat()
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
    try:
      effective_time_begin = parser.parse(begin)
      effective_time_end = parser.parse(end)
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
      'effective_time_begin': begin.isoformat(),
      'effective_time_end': end.isoformat(),
      'timestamp': datetime.datetime.now().isoformat()
    }
    # find the operator entry and add the operation
    for oper in self.operators:
      if oper.get('uss').upper() == uss_id.upper():
        found = True
        # Remove the existing operation, if any
        self.remove_operation(uss_id, gufi)
        oper.operations.append(operation)
        break
      self.timestamp = datetime.datetime.now().isoformat()
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
        oper['timestamp'] = datetime.datetime.now().isoformat()
    return found