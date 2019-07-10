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

This module is the information interface to Cockroach.


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
import json
import logging

# Our data structure for the actual metadata stored
import uss_metadata
# Utilties for validating slippy
import slippy_util

import uvrs

# psycopg2 is a postgres database driver for python's DB API.
import psycopg2

# logging is our log infrastructure used for this application
logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_InformationInterface')

# CONSTANTS
# Lock stores in this format /uss/gridcells/{z}/{x}/{y}/manifest
USS_BASE_PREFIX = '/uss/gridcells/'
TEST_BASE_PREFIX = '/test/'
USS_METADATA_FILE = '/manifest'
DEFAULT_CONNECTION = 'host=localhost port=26257 dbname=defaultdb user=root password='
GRID_PATH = USS_BASE_PREFIX
MAX_SAFE_INTEGER = 9007199254740991
DELETE_ATTEMPTS = 3
OK_STATUS = {200, 204}

SQL_CREATE_TABLE = """
  CREATE TABLE IF NOT EXISTS
    grid (
      tile STRING PRIMARY KEY NOT NULL,
      uss_metadata STRING,
      updated_at TIMESTAMPTZ NOT NULL
    );
"""

SQL_SELECT_METADATA_AND_TRANSACTION_ID_FOR_TILE = """
  SELECT
    uss_metadata, updated_at
  FROM
    grid
  WHERE
    tile=%s;
"""

SQL_UPSERT_METADATA_FOR_TILE_AND_TRANSACTION_ID = """
WITH data (tile, uss_metadata, updated_at) AS (
  VALUES (%s, %s, transaction_timestamp())
)
UPSERT INTO
  grid (tile, uss_metadata, updated_at)
SELECT * FROM
  data
WHERE
  EXISTS(SELECT updated_at FROM grid WHERE tile=%s and updated_at=%s)
OR
  NOT EXISTS (SELECT tile FROM grid WHERE tile=%s)
RETURNING
  updated_at;
"""


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
        cockroach connection string - following postgres URL syntax
      testgroupid:
        ID to use if in test mode, none for normal mode
    """
    if testgroupid:
      self.set_testmode(testgroupid)
    if not connectionstring:
      connectionstring = DEFAULT_CONNECTION
    log.debug('Creating metadata manager object and connecting to cockroach...')
    try:
      self.conn = psycopg2.connect(connectionstring)
      with self.conn.cursor() as curs:
        curs.execute(SQL_CREATE_TABLE)
        self.conn.commit()
      if testgroupid:
        self.delete_testdata(testgroupid)
    except psycopg2.Error:
      log.error('Unable to connect to cockroach using %s connection string...',
                connectionstring)
      raise

  def __del__(self):
    log.debug('Destroying metadata manager object and disconnecting from cockroach...')
    self.conn.close()

  def set_verbose(self):
    log.setLevel(logging.DEBUG)

  def set_testmode(self, testgroupid='UNDEFINED_TESTER'):
    """Sets the mode to testing with the specific test ID, cannot be undone.

    Args:
      testgroupid: ID to use if in test mode, none for normal mode
    """
    global GRID_PATH
    # Adjust parameters specifically for the test
    GRID_PATH = TEST_BASE_PREFIX + testgroupid + USS_BASE_PREFIX
    log.debug('Setting test path to %s...', GRID_PATH)

  def delete_testdata(self, testgroupid=None):
    """Removes the test data from the servers.

    Be careful when using this in parallel as it removes everything under
    the testgroupid, or everything if no tetgroupid is provided.

    Args:
      testgroupid: ID to use if in test mode, none will remove all test data
    """
    if testgroupid:
      path = TEST_BASE_PREFIX + testgroupid + "%"
    else:
      path = TEST_BASE_PREFIX + "%"
    with self.conn.cursor() as curs:
      curs.execute('DELETE FROM grid WHERE tile LIKE %s', [path])
      self.conn.commit()

  def get(self, z, x, y):
    """Gets the metadata and snapshot token for a GridCell.

    Reads data from cockroach, including a snapshot token. The
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
    if slippy_util.validate_slippy(z, x, y):
      (content, txid) = self._get_raw(z, x, y)
      try:
        m = uss_metadata.USSMetadata(content)
        status = 200
        result = {
          'status': 'success',
          'sync_token': txid,
          'data': m.to_json()
        }
      except ValueError:
        status = 412
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
    if operations is None:
      operations = []

    if not slippy_util.validate_slippy(z, x, y):
      return self._format_status_code_to_jsend(400, 'Slippy validation failed')

    # First we have to get the cell content
    (content, txid) = self._get_raw(z, x, y)

    # Quick check of the token; another is done on the actual set to be sure but
    # this check fails early and fast
    if txid != sync_token:
      return self._format_status_code_to_jsend(409, 'Sync token does not match')

    try:
      m = uss_metadata.USSMetadata(content)
      log.debug('Setting metadata for %s...', uss_id)
      m.upsert_operator(uss_id, baseurl, announce,
                        earliest_operation, latest_operation,
                        z, x, y, operations)
    except ValueError as e:
      return self._format_status_code_to_jsend(400, e.message)
    try:
      status = self._set_raw(z, x, y, m, txid)
    except ValueError as e:
      log.error('Failed setting operator for %s with token %s because %s',
                uss_id, str(sync_token), str(e))
      return self._format_status_code_to_jsend(412)

    if status == 200:
      # Success, now get the metadata back to send back
      return self.get(z, x, y)
    else:
      return self._format_status_code_to_jsend(status)

  def set_operation(self, z, x, y, sync_token, uss_id, gufi,
      signature, begin, end):
    """Sets the operation metadata for a GridCell.

    Writes data, using the snapshot token for confirming data
    has not been updated since it was last read. Operator must be in
    grid cell before using this method.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      sync_token: token retrieved in the original GET GridCellMetadata,
      uss_id: plain text identifier for the USS,
      gufi: Unique flight identifier per NASA formatting standards
      signature: The JWS signature of the Operation,
      begin: start time of the operation.
      end: end time of the operation.
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    status = 500
    if not slippy_util.validate_slippy(z, x, y):
      return self._format_status_code_to_jsend(400, 'Slippy validation failed')

    # first we have to get the cell
    (content, txid) = self._get_raw(z, x, y)
    if not txid:
      return self._format_status_code_to_jsend(404)

    # Quick check of the token, another is done on the actual set to be sure
    #    but this check fails early and fast
    if str(txid) != str(sync_token):
      return self._format_status_code_to_jsend(409)

    try:
      m = uss_metadata.USSMetadata(content)
      log.debug('Setting metadata for %s - %s...', uss_id, gufi)
      found = m.upsert_operation(uss_id, gufi, signature, begin, end)
    except ValueError as e:
      return self._format_status_code_to_jsend(400, e.message)

    if not found:
      return self._format_status_code_to_jsend(
          404, 'Operator %s not found' % uss_id)

    try:
      status = self._set_raw(z, x, y, m, txid)
    except ValueError:
      log.error('Failed setting operation for %s with token %s...',
                gufi, str(sync_token))
      return self._format_status_code_to_jsend(412)

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
    if slippy_util.validate_slippy(z, x, y):
      # first we have to get the cell
      (content, txid) = self._get_raw(z, x, y)
      if txid:
        try:
          m = uss_metadata.USSMetadata(content)
          if m.remove_operator(uss_id):
            # TODO(pelletierb): Automatically retry on delete
            status = self._set_raw(z, x, y, m, txid)
          else:
            status = 404
        except ValueError:
          status = 412
      else:
        status = 404
    else:
      status = 400
    if status == 200:
      # Success, now get the metadata back to send back
      (content, txid) = self._get_raw(z, x, y)
      result = {
        'status': 'success',
        'sync_token': txid,
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
    if slippy_util.validate_slippy(z, x, y):
      # first we have to get the cell
      (content, txid) = self._get_raw(z, x, y)
      if txid:
        try:
          m = uss_metadata.USSMetadata(content)
          if m.remove_operation(uss_id, gufi):
            # TODO(pelletierb): Automatically retry on delete
            status = self._set_raw(z, x, y, m, txid)
          else:
            status = 404
        except ValueError:
          status = 412
      else:
        status = 404
    else:
      status = 400
    if status == 200:
      # Success, now get the transaction id back to send back
      (content, txid) = self._get_raw(z, x, y)
      result = {
        'status': 'success',
        'sync_token': txid,
        'data': m.to_json()
      }
    else:
      result = self._format_status_code_to_jsend(status)
    return result

  def get_multi(self, z, grids):
    """Gets the metadata and snapshot token for multiple GridCells.

    Reads data from cockroach, including a composite snapshot token. The
    snapshot token is used as a reference when writing to ensure
    the data has not been updated between read and write.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to retrieve
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    try:
      combined_meta, syncs = self._get_multi_raw(z, grids)
      log.debug('Found sync token %s for %d grids...',
                self._hash_transaction_ids(syncs), len(syncs))
      result = {
        'status': 'success',
        'sync_token': self._hash_transaction_ids(syncs),
        'data': combined_meta.to_json()
      }
    except ValueError as e:
      result = self._format_status_code_to_jsend(400, e.message)
    except IndexError as e:
      result = self._format_status_code_to_jsend(404, e.message)
    return result

  def set_multi(self, z, grids, sync_token, uss_id, baseurl, announce,
    earliest_operation, latest_operation, operations=None):
    """Sets multiple GridCells metadata at once.

    Writes data, using the hashed snapshot token for confirming data
    has not been updated since it was last read.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to update
      sync_token: composite token retrieved in the original get_multi,
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
    log.debug('Setting multiple grid metadata for %s...', uss_id)
    try:
      # first, get the affected grid's sync tokens
      m, syncs = self._get_multi_raw(z, grids)
      del m
      # Quick check of the token, another is done on the actual set to be sure
      #    but this check fails early and fast
      log.debug('Found sync token %d for %d grids...',
                self._hash_transaction_ids(syncs), len(syncs))
      if str(self._hash_transaction_ids(syncs)) == str(sync_token):
        log.debug('Composite sync_token matches, continuing...')
        self._set_multi_raw(z, grids, syncs, uss_id, baseurl, announce,
                            earliest_operation, latest_operation, operations)
        log.debug('Completed updating multiple grids...')
      else:
        raise KeyError('Composite sync_token has changed')
      combined_meta, new_syncs = self._get_multi_raw(z, grids)
      result = {
        'status': 'success',
        'sync_token': self._hash_transaction_ids(new_syncs),
        'data': combined_meta.to_json()
      }
    except (KeyError, psycopg2.DatabaseError) as e:
      result = self._format_status_code_to_jsend(409, e.message)
    except ValueError as e:
      result = self._format_status_code_to_jsend(400, e.message)
    except IndexError as e:
      result = self._format_status_code_to_jsend(404, e.message)
    return result

  def set_multi_operation(self, z, grids, sync_token, uss_id, gufi,
    signature, begin, end):
    """Sets the metadata for an operation for multiple GridCells.

    Writes data, using the snapshot token for confirming data
    has not been updated since it was last read. Operator must be in
    the grid cells before using this method.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to update
      sync_token: composite token retrieved in the original get_multi,
      uss_id: plain text identifier for the USS,
      gufi: Unique flight identifier per NASA formatting standards
      signature: The JWS signature of the Operation,
      begin: start time of the operation.
      end: end time of the operation.
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    log.debug('Setting multiple grid operation data for %s...', uss_id)
    try:
      # first, get the affected grid's sync tokens
      m, syncs = self._get_multi_raw(z, grids)
      del m
      # Quick check of the token, another is done on the actual set to be sure
      #    but this check fails early and fast
      log.debug('Found composite sync token %d for %d grids...',
                self._hash_transaction_ids(syncs), len(syncs))
      if str(self._hash_transaction_ids(syncs)) == str(sync_token):
        log.debug('Composite sync_token matches, continuing...')
        self._set_multi_operation_raw(z, grids, syncs, uss_id, gufi,
                                      signature, begin, end)
        log.debug('Completed updating multiple grids...')
      else:
        raise KeyError('Composite sync_token has changed')
      combined_meta, new_syncs = self._get_multi_raw(z, grids)
      result = {
        'status': 'success',
        'sync_token': self._hash_transaction_ids(new_syncs),
        'data': combined_meta.to_json()
      }
    except (KeyError, psycopg2.DatabaseError) as e:
      result = self._format_status_code_to_jsend(409, e.message)
    except ValueError as e:
      result = self._format_status_code_to_jsend(400, e.message)
    except IndexError as e:
      result = self._format_status_code_to_jsend(404, e.message)
    return result

  def insert_uvr(self, z, grids, uvr):
    """Emplaces a UVR in multiple GridCells' metadata at once.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to update
      uvr: validated UVR data structure, as defined in USSMetadata

    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    log.debug('Setting UVR %s in %d grid cells for %s...',
              uvr['message_id'], len(grids), uvr['uss_name'])
    try:
      # Get and update USSMetadata for all affected cells in memory
      contents = []
      for i in range(len(grids)):
        x = grids[i][0]
        y = grids[i][1]
        path = '%s/%s/%s/%s/%s' % (GRID_PATH, str(z), str(x), str(y),
                                   USS_METADATA_FILE)
        (content, txid) = self._get_raw(z, x, y)
        m = uss_metadata.USSMetadata(content)
        m.insert_uvr(uvr)
        contents.append((path, m, txid))

      # Commit changes to cockroach
      self._execute_transaction(contents)
      log.debug('Completed updating multiple grids with UVR...')

      return self.get_multi(z, grids)
    except psycopg2.DatabaseError as e:
      result = self._format_status_code_to_jsend(409, e.message)
    except (KeyError, ValueError) as e:
      result = self._format_status_code_to_jsend(400, e.message)
    except IndexError as e:
      result = self._format_status_code_to_jsend(404, e.message)
    return result

  def delete_multi(self, z, grids, uss_id):
    """Sets multiple GridCells metadata by removing the entry for the USS.

    Removes the operator from multiple cells. Does not return 404 on
    not finding the USS in a cell, since this should be a remove all
    type function, as some cells might have the ussid and some might not.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to delete
      uss_id: is the plain text identifier for the USS
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    log.debug('Deleting multiple grid metadata for %s...', uss_id)
    try:
      if not uss_id:
        raise ValueError('Invalid uss_id for deleting multi')
      for x, y in grids:
        if slippy_util.validate_slippy(z, x, y):
          (content, txid) = self._get_raw(z, x, y)
          if txid:
            m = uss_metadata.USSMetadata(content)
            m.remove_operator(uss_id)
            # TODO(pelletierb): Automatically retry on delete
            status = self._set_raw(z, x, y, m, txid)
        else:
          raise ValueError('Invalid slippy grids for lookup')
      result = self.get_multi(z, grids)
    except ValueError as e:
      result = self._format_status_code_to_jsend(400, e.message)
    return result

  def delete_multi_operation(self, z, grids, uss_id, gufi):
    """Sets multiple GridCells metadata by removing the operation for the USS.

    Removes the operator from multiple cells. Does not return 404 on
    empty, since this should be a remove all type function.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to delete
      uss_id: is the plain text identifier for the USS
      gufi: the uinique gufi to remove
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    log.debug('Deleting multiple grid operations for %s...', uss_id)
    try:
      if not uss_id:
        raise ValueError('Invalid uss_id for deleting multi')
      for x, y in grids:
        if slippy_util.validate_slippy(z, x, y):
          (content, txid) = self._get_raw(z, x, y)
          if txid:
            m = uss_metadata.USSMetadata(content)
            if m.remove_operation(uss_id, gufi):
              # TODO(pelletierb): Automatically retry on delete
              status = self._set_raw(z, x, y, m, txid)
        else:
          raise ValueError('Invalid slippy grids for lookup')
      result = self.get_multi(z, grids)
    except ValueError as e:
      result = self._format_status_code_to_jsend(400, e.message)
    return result

  def delete_uvr(self, z, grids, uvr):
    """Updates multiple GridCells' metadata by removing the entry for the UVR.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to delete
      uvr: validated nested-dict UVR data structure
    Returns:
      JSend formatted response (https://labs.omniti.com/labs/jsend)
    """
    log.debug('Deleting UVR %s in %d grid cells for %s...',
              uvr['message_id'], len(grids), uvr['uss_name'])
    try:
      status = 200
      for _ in range(DELETE_ATTEMPTS):
        # First, plan all deletions without modifying anything
        deletions = []
        for x, y in grids:
          if not slippy_util.validate_slippy(z, x, y):
            raise ValueError('Invalid slippy grids for lookup')
          (content, txid) = self._get_raw(z, x, y)
          if txid:
            m = uss_metadata.USSMetadata(content)
            removed_uvr = m.remove_uvr(uvr['message_id'])
            if removed_uvr:
              diff = uvrs.diff(removed_uvr, uvr)
              if diff:
                return self._format_status_code_to_jsend(
                    400, 'When deleting a UVR, the UVR definition must match '
                    'the existing UVR exactly. Found %s in existing UVR in '
                    'grid cell (%d, %d) and %s in UVR deletion request' %
                    (diff[0], x, y, diff[1]))
              deletions.append((z, x, y, m, txid))

        # Now, attempt all mutations at once
        status = 200
        for d in deletions:
          status = self._set_raw(*d)
          if status not in OK_STATUS:
            break

        if status in OK_STATUS:
          break

      if status in OK_STATUS:
        return self.get_multi(z, grids)
      else:
        return self._format_status_code_to_jsend(
          status, 'UVR deletion failed for grid (%d, %d)' % (d[1], d[2]))
    except ValueError as e:
      return self._format_status_code_to_jsend(400, e.message)

  ######################################################################
  ################       INTERNAL FUNCTIONS    #########################
  ######################################################################
  def _get_raw(self, z, x, y):
    """Gets the raw content and the timestamp of the last update for a GridCell from cockroach.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
    Returns:
      content: USS metadata
      transaction_id: id of the last transaction that updated the content (might be None).
    """
    path = '%s/%s/%s/%s/%s' % (GRID_PATH, str(z), str(x), str(y),
                               USS_METADATA_FILE)
    log.debug('Getting metadata from cockroach@%s...', path)
    with self.conn.cursor() as curs:
      c = None
      ts = None
      curs.execute(
        SQL_SELECT_METADATA_AND_TRANSACTION_ID_FOR_TILE,
        [path]
      )
      if curs.rowcount > 0:
        c, ts = curs.fetchone()
      if c:
        log.debug('Received raw content and metadata from cockroach: %s', c)
      if ts:
        log.debug('Received timestamp from cockroach: %s', ts)

      return c, ts

  def _set_raw(self, z, x, y, m, version):
    """Grabs the lock and updates the raw content for a GridCell in cockroach.

    Args:
      z: zoom level in slippy tile format
      x: x tile number in slippy tile format
      y: y tile number in slippy tile format
      m: metadata object to write
      version: the metadata version verified from the sync_token match
    Returns:
      200 for success, 409 for conflict, 408 for unable to get the lock
    """
    path = '%s/%s/%s/%s/%s' % (GRID_PATH, str(z), str(x), str(y),
                               USS_METADATA_FILE)
    try:
      log.debug('Setting metadata to %s...', str(m.to_json(True)))
      with self.conn.cursor() as curs:
        curs.execute(
          SQL_UPSERT_METADATA_FOR_TILE_AND_TRANSACTION_ID,
          [path, json.dumps(m.to_json(True)), path, version, path]
        )
        # TODO(vosst): Returning the new timestamp would be helpful to USSs
        # as they don't have to resync after an update.
        ts = curs.fetchone()
        self.conn.commit()
        status = 200
    except (psycopg2.ProgrammingError, psycopg2.IntegrityError) as e:
      log.error('Sync token updated before write for %s: %s', path, e)
      status = 409
    return status

  def _get_multi_raw(self, z, grids):
    """Gets the raw content and metadata for multiple GridCells from cockroach.

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to retrieve
    Returns:
      content: Combined USS metadata
      syncs: list of sync tokens in the same order as the grids
    Raises:
      IndexError: if it cannot find anything in cockroach
      ValueError: if the grid data is not in the right format
    """
    log.debug('Getting multiple grid metadata for %s...', str(grids))
    combined_meta = None
    syncs = []
    for x, y in grids:
      if slippy_util.validate_slippy(z, x, y):
        (content, txid) = self._get_raw(z, x, y)
        combined_meta += uss_metadata.USSMetadata(content)
        syncs.append(txid)
      else:
        raise ValueError('Invalid slippy grids for lookup')
    if len(syncs) == 0:
      raise IndexError('Unable to find metadata in platform')
    return combined_meta, syncs

  def _set_multi_raw(self, z, grids, versions, uss_id, baseurl, announce,
    earliest_operation, latest_operation, operations=None):
    """Grabs the lock and updates the raw content for multiple GridCells

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to retrieve
      sync_tokens: list of the sync tokens received during get operation
      uss_id: plain text identifier for the USS,
      ws_scope: scope to use to obtain OAuth token,
      operation_format: output format for operation ws (i.e. NASA, GUTMA),
      operation_ws: submitting USS endpoint where all flights in
        this cell can be retrieved from,
      earliest_operation: lower bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      latest_operation: upper bound of active or planned flight timestamp,
        used for quick filtering conflicts.
      operations: array of individual operations for this uss in these cells
    Returns:
      updatedVersions: new versions of all updated tiles
    Raises:
      IndexError: if it cannot find anything in cockroach
      ValueError: if the grid data is not in the right format
    """
    log.debug('Setting multiple grid metadata for %s...', str(grids))
    contents = []
    for i in range(len(grids)):
      # First, get and update them all in memory, validate the sync_token
      x = grids[i][0]
      y = grids[i][1]
      version = versions[i]
      path = '%s/%s/%s/%s/%s' % (GRID_PATH, str(z), str(x), str(y),
                                  USS_METADATA_FILE)
      (content, txid) = self._get_raw(z, x, y)
      if str(version) == str(txid):
        log.debug('version matches for %d, %d...', x, y)
        m = uss_metadata.USSMetadata(content)
        m.upsert_operator(uss_id, baseurl, announce,
                          earliest_operation, latest_operation,
                          z, x, y, operations)
        contents.append((path, m, version))
      else:
        log.error(
          'Version from USS (%s) does not match version in cockroach (%s)...',
          str(version), str(txid))
        raise KeyError('Composite version has changed')
    return self._execute_transaction(contents)

  def _set_multi_operation_raw(self, z, grids, sync_tokens, uss_id, gufi,
    signature, begin, end):
    """Grabs the lock and updates the raw content for multiple GridCells

    Args:
      z: zoom level in slippy tile format
      grids: list of (x,y) tiles to retrieve
      sync_tokens: list of the sync tokens received during get operation
      uss_id: plain text identifier for the USS,
      gufi: Unique flight identifier per NASA formatting standards
      signature: The JWS signature of the Operation,
      begin: start time of the operation.
      end: end time of the operation.
    Raises:
      IndexError: if it cannot find anything in cockroach
      ValueError: if the grid data is not in the right format
      psycopg2.DatabaseError: if update transaction does not complete
    """
    log.debug('Setting multiple grid operation metadata for %s...', str(grids))
    try:
      contents = []
      for i in range(len(grids)):
        # First, get and update them all in memory, validate the sync_token
        x = grids[i][0]
        y = grids[i][1]
        sync_token = sync_tokens[i]
        path = '%s/%s/%s/%s/%s' % (GRID_PATH, str(z), str(x), str(y),
                                   USS_METADATA_FILE)
        (content, txid) = self._get_raw(z, x, y)
        if str(txid) == str(sync_token):
          log.debug('Sync_token matches for %d, %d...', x, y)
          m = uss_metadata.USSMetadata(content)
          # TODO(hikevin): refactor with multi_operator, as only one line diffs
          if not m.upsert_operation(uss_id, gufi, signature, begin, end):
            log.error('Failed setting operation for %s with token %s...',
                      gufi, str(sync_token))
            raise ValueError('Failed setting operation in grid %d/%d/%d' %
                             (z, x, y))
          contents.append((path, m, txid))
        else:
          log.error(
            'Sync token from USS (%s) does not match token from zk (%s)...',
            str(sync_token), str(txid))
          raise KeyError('Composite sync_token has changed')

      self._execute_transaction(contents)
    except (KeyError, ValueError, IndexError, psycopg2.DatabaseError) as e:
      log.error('Error caught in set_multi_raw %s.', e.message)
      raise e

  def _execute_transaction(self, contents):
    """Write to multiple grid cells in a single transaction with auto-rollback.

    The version data will catch any outside changes and roll back any attempted
    updates to the grids.

    Args:
      contents: iterable of tuples of (path, USSMetadata, version)
        path: key of the grid cell data
        USSMetadata: USSMetadata instance containing updated info
        version: previous version of grid cell data upon which the updated info
          is based

    Raises:
      psycopg2.DatabaseError: if transaction could not be completed due to intermediate
        updates.
    """
    log.debug('Starting transaction to write all grids at once...')
    updatedVersions = []
    with self.conn.cursor() as curs:
      for path, m, version in contents:
        curs.execute(
          SQL_UPSERT_METADATA_FOR_TILE_AND_TRANSACTION_ID,
          [path, json.dumps(m.to_json(True)), path, version, path]
        )
        updatedVersions.append(curs.fetchone())
      self.conn.commit()
      return updatedVersions

  def _format_status_code_to_jsend(self, status, message=None):
    """Formats a response based on HTTP status code.

    Args:
      status: HTTP status code
      message: optional message to override preset message for codes
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
          'message': 'Unable to find metadata in uss discovery platform.'
      }
    elif status == 408:
      result = {
          'status': 'fail',
          'code': status,
          'message': 'Timeout trying to update metadata information.'
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
    elif status == 412:
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
    if message:
      result['message'] = message
    return result

  @staticmethod
  def _hash_transaction_ids(txids):
    """Hashes a list of transaction ids into a single, positive 64-bit int.

    For various languages, the limit to integers may be different, therefore
    we truncate to ensure the hash is the same on all implementations.
    """
    return abs(hash(tuple(sorted(filter(lambda x: x != None, txids))))) % MAX_SAFE_INTEGER
