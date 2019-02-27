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
import os

import requests


log = logging.getLogger('InterUSSPlatform')


class Client(object):
  def __init__(self, base_url, zoom, auth_url, username, password):
    """Instantiate a client to communicate with an InterUSS Platform server.

    Args:
      base_url: URL of InterUSS Platform server, without any verbs.
      zoom: Integer zoom level in which operations are being coordinated.
      auth_url: URL to POST to with credentials to obtain access token.
      username: Basic authorization username to obtain access token.
      password: Basic authorization password to obtain access token.
    """
    self._base_url = base_url
    self._zoom = zoom
    self._auth_url = auth_url
    self._username = username
    self._password = password

    self._access_token = None
    self._token_expires = None

  def get_access_token(self):
    """Return a valid access token, retrieving a new one if necessary."""
    if self._access_token and datetime.datetime.utcnow() < self._token_expires:
      return self._access_token
    log.info('Retrieving new token')
    t0 = datetime.datetime.utcnow()
    response = requests.post(
      url=self._auth_url,
      auth=(self._username, self._password))
    response.raise_for_status()
    r = response.json()
    self._access_token = r['access_token']
    self._token_expires = t0 + datetime.timedelta(seconds=r['expires_in'])
    return self._access_token

  def get_operators(self, area):
    """Retrieve operators in specified area from an InterUSS Platform server.

    Args:
      area: Sequence of geo points (with .lat and .lng degrees) describing the
        polygon area of interest.

    Returns:
      List of operators in InterUSS Platform API format.
    """
    access_token = self.get_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.get(
      url=self._base_url + '/GridCellsMetaData/%d' % self._zoom,
      headers={'access_token': access_token},
      params={'coords': coords, 'coord_type': 'polygon'})
    response.raise_for_status()
    return response.json()['data']['operators']

  def get_public_portal(self, public_portal_endpoint, area):
    """Query USS (not InterUSS Platform) for its public operations in an area.

    Args:
      public_portal_endpoint: USS's public_portal endpoint conforming with the
        InterUSS Platform public portal spec.
      area: Sequence of geo points (with .lat and .lng degrees) describing the
        polygon area of interest.

    Returns:
      Bare public portal response in the InterUS Platform public portal format.
    """
    access_token = self.get_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.get(
      url=os.path.join(public_portal_endpoint, coords),
      headers={'access_token': access_token})
    response.raise_for_status()
    return response.json()

  def get_flight_info(self, flight_info_endpoint, uuid):
    """Query USS (not InterUSS Platform) for details about a specific flight.

    Args:
      flight_info_endpoint: USS's flight_info endpoint conforming with the
        InterUSS Platform public portal spec.
      uuid: Unique ID of flight as indicated in the USS's public portal
        response.

    Returns:
      Data of flight info response in the InterUS Platform public portal format.
    """
    access_token = self.get_access_token()
    response = requests.get(
      url=os.path.join(flight_info_endpoint, uuid),
      headers={'access_token': access_token})
    response.raise_for_status()
    return response.json()['data']
