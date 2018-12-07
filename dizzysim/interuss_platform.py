import datetime
import logging

import requests

import formatting

log = logging.getLogger('InterUSSPlatform')


class Client(object):
  def __init__(self, base_url, zoom, auth_url, username, password,
               public_portal_endpoint, flight_info_endpoint):
    self._base_url = base_url
    self._zoom = zoom
    self._auth_url = auth_url
    self._username = username
    self._password = password
    self._public_portal_endpoint = public_portal_endpoint
    self._flight_info_endpoint = flight_info_endpoint

    self._access_token = None
    self._token_expires = None

  def _refresh_access_token(self):
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

  def remove_operations(self, area):
    self._refresh_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.delete(
      url=self._base_url + '/GridCellsMetaData/%d' % self._zoom,
      headers={'access_token': self._access_token},
      json={
        'coords': coords,
        'coord_type': 'polygon'
      })
    response.raise_for_status()
