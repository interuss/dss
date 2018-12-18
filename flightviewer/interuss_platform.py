import datetime
import logging
import os

import requests


log = logging.getLogger('InterUSSPlatform')


class Client(object):
  def __init__(self, base_url, zoom, auth_url, username, password):
    self._base_url = base_url
    self._zoom = zoom
    self._auth_url = auth_url
    self._username = username
    self._password = password

    self._access_token = None
    self._token_expires = None

  def get_access_token(self):
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
    access_token = self.get_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.get(
      url=self._base_url + '/GridCellsMetaData/%d' % self._zoom,
      headers={'access_token': access_token},
      params={'coords': coords, 'coord_type': 'polygon'})
    response.raise_for_status()
    return response.json()['data']['operators']

  def get_public_portal(self, public_portal_endpoint, area):
    access_token = self.get_access_token()
    coords = ','.join('%.6f,%.6f' % (p.lat, p.lng) for p in area)
    response = requests.get(
      url=os.path.join(public_portal_endpoint, coords),
      headers={'access_token': access_token})
    response.raise_for_status()
    return response.json()

  def get_flight_info(self, flight_info_endpoint, uuid):
    access_token = self.get_access_token()
    response = requests.get(
      url=os.path.join(flight_info_endpoint, uuid),
      headers={'access_token': access_token})
    response.raise_for_status()
    return response.json()['data']
