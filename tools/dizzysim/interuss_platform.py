"""Tools for interacting with the InterUSS Platform.

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

import datetime
import logging

import requests

import formatting

log = logging.getLogger('InterUSSPlatform')


class Client(object):
  """Client wrapper around interactions with an InterUSS Platform server."""

  def __init__(self, base_url, zoom, auth_url, username, password,
               public_portal_endpoint, flight_info_endpoint):
    """Instantiate a client to communicate with an InterUSS Platform server.
    
    Args:
      base_url: URL of InterUSS Platform server, without any verbs.
      zoom: Integer zoom level in which operations are being coordinated.
      auth_url: URL to POST to with credentials to obtain access token.
      username: Basic authorization username to obtain access token.
      password: Basic authorization password to obtain access token.
      public_portal_endpoint: The public_portal endpoint of this USS, which
        will be advertised to the InterUSS Platform grid.
      flight_info_endpoint: The flight_info endpoint of this USS, which will be
        advertised to the InterUSS Platform grid.
    """
    self._base_url = base_url
    self._zoom = zoom
    self._auth_url = auth_url
    self._username = username
    self._password = password
    self._public_portal_endpoint = public_portal_endpoint
    self._flight_info_endpoint = flight_info_endpoint

    self._access_token = None
    self._token_expires = None

    self._op_area = None

  def _refresh_access_token(self):
    """Ensure that self._access_token contains a valid access token."""
    if self._access_token and datetime.datetime.utcnow() < self._token_expires:
      return
    log.info('Retrieving new token')
    t0 = datetime.datetime.utcnow()
    response = requests.post(
      url=self._auth_url,
      auth=(self._username, self._password))
    response.raise_for_status()
    r = response.json()
    self._access_token = r['access_token']
    self._token_expires = t0 + datetime.timedelta(seconds=r['expires_in'])

  def set_operations(self, area, minimum_operation_timestamp,
                     maximum_operation_timestamp):
    """Inform the InterUSS Platform of intended operations from this USS.
    
    Args:
      area: Sequence of geo points (with .lat and .lng degrees) describing the
        polygon area containing all operations to be conducted. Note that
        remove_operations should eventually be called after this method to avoid
        leaving orphaned entries in the grid.
      minimum_operation_timestamp: Python datetime indicating the earliest time
        an operation may begin.
      maximum_operation_timestamp: Python datetime indicating the latest time
        an operation may still be in effect.
    """
    if self._op_area is not None:
      self.remove_operations()

    self._refresh_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.put(
      url=self._base_url + '/GridCellsMetaData/%d' % self._zoom,
      headers={'access_token': self._access_token},
      json={
        'coords': coords,
        'coord_type': 'polygon',
        'scope': 'interussplatform.com_operators.read',
        'public_portal_endpoint': self._public_portal_endpoint,
        'flight_info_endpoint': self._flight_info_endpoint,
        'minimum_operation_timestamp': formatting.timestamp(
          minimum_operation_timestamp),
        'maximum_operation_timestamp': formatting.timestamp(
          maximum_operation_timestamp),
      })
    response.raise_for_status()
    self._op_area = area

  def remove_operations(self):
    """Inform the InterUSS Platform that operations have ceased."""
    if self._op_area is None:
      raise ValueError('Cannot remove operations when no operations are active')
    self._refresh_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in self._op_area)
    response = requests.delete(
      url=self._base_url + '/GridCellsMetaData/%d' % self._zoom,
      headers={'access_token': self._access_token},
      json={
        'coords': coords,
        'coord_type': 'polygon'
      })
    response.raise_for_status()
    self._op_area = None
