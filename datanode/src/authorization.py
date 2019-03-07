"""The InterUSS Platform Data Node authorization tools.

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

import abc
import json
import logging
import numbers
import os
import sys
import flask
import jwt
import requests
from rest_framework import status
from shapely.geometry import Polygon

import slippy_util

logging.basicConfig(format='%(levelname)s:%(message)s', level=logging.INFO)
log = logging.getLogger('InterUSS_DataNode_Authorization')


def JoinZoom(zoom, tiles):
  """Combine single zoom and multiple tiles into tile triplets."""
  return ((zoom, t[0], t[1]) for t in tiles)


class Authorizer(object):
  """Manages authorization on a per-area basis."""

  def __init__(self, public_key, auth_config_string):
    self.test_id = None
    self.authorities = []
    self._cache = {}

    if public_key:
      log.info('Using global auth provider from single public key')
      self.authorities.append(
        _AuthorizationAuthority(public_key, 'Global authority'))

    if auth_config_string:
      self.authorities.extend(_ParseAuthorities(auth_config_string))

  def SetTestId(self, testid):
    self.test_id = testid
    log.info('Authorization set to test mode with test ID=%s' % self.test_id)

  def ValidateAccessToken(self, headers, tiles=None):
    """Checks the access token, aborting if it does not pass.

    Uses one or more OAuth public keys to validate an access token.

    Args:
      headers: dict of headers from flask.request
      tiles: collection of (zoom, x,y) slippy tiles user is attempting to access

    Returns:
      USS identification from OAuth client_id or sub field

    Raises:
      HTTPException: when the access token is invalid or inappropriate
    """
    uss_id = None
    if self.test_id:
      if self.test_id in headers.get('access_token', ''):
        return headers['access_token']
      elif self.test_id in headers.get('Authorization', ''):
        return headers['Authorization']
      elif 'access_token' not in headers and 'Authorization' not in headers:
        return self.test_id

    # TODO(hikevin): Replace with OAuth Discovery and JKWS
    token = None
    if 'Authorization' in headers:
      token = headers['Authorization'].replace('Bearer ', '')
    elif 'access_token' in headers:
      token = headers['access_token']
    if not token:
      log.error('Attempt to access resource without access_token in header.')
      flask.abort(status.HTTP_403_FORBIDDEN,
                  'Valid OAuth access_token must be provided in header.')

    # Verify validity of claims without checking signature
    try:
      r = jwt.decode(token, algorithms='RS256', verify=False)
      #TODO(hikevin): Check scope is valid for InterUSS Platform
      uss_id = r['client_id'] if 'client_id' in r else r.get('sub', None)
    except jwt.ImmatureSignatureError:
      log.error('Access token is immature.')
      flask.abort(status.HTTP_401_UNAUTHORIZED,
                  'OAuth access_token is invalid: token is immature.')
    except jwt.ExpiredSignatureError:
      log.error('Access token has expired.')
      flask.abort(status.HTTP_401_UNAUTHORIZED,
                  'OAuth access_token is invalid: token has expired.')
    except jwt.DecodeError:
      log.error('Access token is invalid and cannot be decoded.')
      flask.abort(status.HTTP_400_BAD_REQUEST,
                  'OAuth access_token is invalid: token cannot be decoded.')
    except jwt.InvalidTokenError as e:
      log.error('Unexpected InvalidTokenError: %s', str(e))
      flask.abort(status.HTTP_500_INTERNAL_SERVER_ERROR,
                  'Unexpected token error: ' + str(e))
    issuer = r.get('iss', None)

    if tiles:
      # Check only authorities that manage all specified tiles
      authorities = set.intersection(
          *[self._GetAuthorities(issuer, tile) for tile in tiles])
    else:
      authorities = self._GetAuthorities(issuer, None)
    if not authorities:
      flask.abort(status.HTTP_401_UNAUTHORIZED,
                  'No authorization authorities could be found')

    # Check signature against all possible public keys
    valid = False
    for authority in authorities:
      try:
        jwt.decode(token, authority.public_key, algorithms='RS256')
        valid = True
        break
      except jwt.InvalidSignatureError:
        # Access token signature not valid for this public key, but might be
        # valid for a different public key.
        pass
    if not valid:
      # Check against all authorities
      for authority in self.authorities:
        try:
          jwt.decode(token, authority.public_key, algorithms='RS256')
          invalid_tile = None
          if tiles:
            for tile in tiles:
              if not authority.is_applicable(issuer, tile):
                invalid_tile = tile
                break
          if invalid_tile:
            flask.abort(
                status.HTTP_401_UNAUTHORIZED,
                'Access token has valid signature but %s is not applicable to '
                'tile %s' % (authority.name, str(invalid_tile)))
          else:
            flask.abort(status.HTTP_401_UNAUTHORIZED,
                        'Access token has valid signature but does not match '
                        'server\'s authority configuration')
        except jwt.InvalidSignatureError:
          # Token signature isn't valid for this authority
          pass
      flask.abort(status.HTTP_401_UNAUTHORIZED,
                  'Access token signature is invalid')

    return uss_id

  def _GetAuthorities(self, issuer, tile):
    """Retrieve set of applicable AuthorizationAuthorities."""
    cache_key = (issuer, tile)
    if cache_key not in self._cache:
      self._cache[cache_key] = set(a for a in self.authorities
                                   if a.is_applicable(issuer, tile))
    return self._cache[cache_key]


class _AuthorizationAuthority(object):
  """Authority that grants access tokens to access part of this data node."""

  def __init__(self, public_key, name):
    # ENV variables sometimes don't pass newlines, spec says white space
    # doesn't matter, but pyjwt cares about it, so fix it
    public_key = public_key.replace(' PUBLIC ', '_PLACEHOLDER_')
    public_key = public_key.replace(' ', '\n')
    public_key = public_key.replace('_PLACEHOLDER_', ' PUBLIC ')
    self.public_key = public_key
    self.constraints = []
    self.name = name

  def is_applicable(self, issuer, tile):
    """Determine whether an AuthorizationAuthority is applicable for tile.

    Args:
      issuer: Content of access token's `iss` JWT field.
      tile: Slippy (zoom, x, y) for tile of interest.

    Returns:
      True if this AuthorizationAuthority is applicable for the specified tile.
    """
    if self.constraints:
      return all(c.is_applicable(issuer, tile) for c in self.constraints)
    else:
      return True


class _AuthorizationConstraint(object):
  """Base class for constraints on when an AuthorizationAuthority applies."""
  __metaclass__ = abc.ABCMeta

  @abc.abstractmethod
  def is_applicable(self, issuer, tile):
    """Determine whether an AuthorizationAuthority is applicable for tile.

    Args:
      issuer: Content of access token's `iss` JWT field.
      tile: Slippy (zoom, x, y) for tile of interest.

    Returns:
      True if the associated AuthorizationAuthority is applicable for the
      specified tile.
    """
    raise NotImplementedError('Abstract method is_applicable not implemented')


class _IssuerConstraint(_AuthorizationConstraint):
  """Access token `iss` field must match issuer."""

  def __init__(self, issuer):
    self._issuer = issuer

  def is_applicable(self, issuer, tile):
    # Overrides method in parent class.
    return issuer == self._issuer


class _AreaConstraint(_AuthorizationConstraint):
  """Tiles must lie in or out of an arbitrary geo polygon."""

  def __init__(self, points, inside):
    self._polygon = Polygon(points)
    self._inside = inside

  def is_applicable(self, issuer, tile):
    # Overrides method in parent class.
    if not tile:
      return True
    tile_polygon = Polygon(slippy_util.convert_tile_to_polygon(*tile))
    if self._inside:
      return self._polygon.intersects(tile_polygon)
    else:
      return not tile_polygon.within(self._polygon)


class _RangeConstraint(_AuthorizationConstraint):
  """Tiles must lie inside one or more explicit Slippy tile ranges."""

  def __init__(self, tile_ranges, inclusive):
    self._tile_ranges = []
    for tile_range in tile_ranges:
      z_range, x_range, y_range = tile_range
      z_min, z_max = _GetRangeBounds(z_range)
      x_min, x_max = _GetRangeBounds(x_range)
      y_min, y_max = _GetRangeBounds(y_range)
      self._tile_ranges.append((z_min, z_max, x_min, x_max, y_min, y_max))
    self._inclusive = inclusive

  def is_applicable(self, issuer, tile):
    # Overrides method in parent class.
    if not tile:
      return True
    zoom, x, y = tile
    in_range = False
    for tile_range in self._tile_ranges:
      z_min, z_max, x_min, x_max, y_min, y_max = tile_range
      if ((z_min <= zoom <= z_max) and
          (x_min <= x <= x_max) and
          (y_min <= y <= y_max)):
        in_range = True
        break
    return in_range == self._inclusive


def _GetRangeBounds(range_spec):
  """Convert a JSON range spec into min and max bounds of that range.

  Args:
    range_spec: Decoded JSON structure describing range.
      Number: Value must match specified value exactly.
      "*": Any value accepted.
      "XXX:YYY": Any value accepted between XXX and YYY, inclusive.

  Returns:
    min_bound: Minimum inclusive bound of range.
    max_bound: Maximum inclusive bound of range.
  """
  if isinstance(range_spec, numbers.Number):
    return range_spec, range_spec
  elif isinstance(range_spec, basestring):
    if range_spec == '*':
      return float('-inf'), float('inf')
    elif ':' in range_spec:
      bounds = range_spec.split(':')
      if len(bounds) != 2:
        raise ValueError('Too many bounds in range_spec')
      return int(bounds[0]), int(bounds[1])
  raise ValueError('Invalid range_spec')


def _ParseAuthorities(config_string):
  """Create a list of AuthorizationAuthorities based on JSON configuration.

  Example JSON:
  [{"public_key": "-----BEGIN PUBLIC KEY----- ..."},
   {"name": "Specific area authority",
    "public_key": "-----BEGIN PUBLIC KEY----- ...",
    "constraints": [{
      "type": "issuer",
      "issuer": "gov.area.authority"}, {
      "type": "area",
      "outline": [[30.064,-99.147],[30.054,-99.147],[30.054,-99.134],[30.064,-99.134]],

  Also see authorization_test.py.

  Args:
    config_string: Configuration description of authorization authorities.
      If a resource URL (file|http|https://), first load content from URL.
      Interpreted as JSON per examples.

  Returns:
    List of AuthorizationAuthorities described in provided config.
  """
  if config_string.startswith('file://'):
    with open(config_string[len('file://'):], 'r') as f:
      config_string = f.read()
  if (config_string.startswith('http://') or
      config_string.startswith('https://')):
    req = requests.get(config_string)
    config_string = req.content

  authority_specs = json.loads(config_string)
  authorities = []
  for i, authority_spec in enumerate(authority_specs):
    authority = _AuthorizationAuthority(
      authority_spec['public_key'],
      authority_spec.get('name', 'Authority %d' % i))
    if 'constraints' in authority_spec:
      for constraint_spec in authority_spec['constraints']:
        if constraint_spec['type'] == 'issuer':
          constraint = _IssuerConstraint(constraint_spec['issuer'])
        elif constraint_spec['type'] == 'area':
          constraint = _AreaConstraint(
            constraint_spec['outline'], constraint_spec.get('inside', True))
        elif constraint_spec['type'] == 'range':
          constraint = _RangeConstraint(
            constraint_spec['tiles'], constraint_spec.get('inclusive', True))
        else:
          raise ValueError('Invalid constraint type: ' +
                           constraint_spec.get('type', '<not specified>'))
        authority.constraints.append(constraint)
    authorities.append(authority)
  return authorities
