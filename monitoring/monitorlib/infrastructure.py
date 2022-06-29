import asyncio
import datetime
import functools
from typing import Dict, List, Optional
import urllib.parse
from aiohttp import ClientSession

import jwt
import requests

ALL_SCOPES = [
  'dss.write.identification_service_areas',
  'dss.read.identification_service_areas',
  'utm.strategic_coordination',
  'utm.constraint_management',
  'utm.constraint_consumption'
]

EPOCH = datetime.datetime.utcfromtimestamp(0)
TOKEN_REFRESH_MARGIN = datetime.timedelta(seconds=15)


class AuthAdapter(object):
  """Base class for an adapter that add JWTs to requests."""

  def __init__(self):
    self._tokens = {}

  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    """Subclasses must return a bearer token for the given audience."""

    raise NotImplementedError()

  def get_headers(self, url: str, scopes: List[str] = None) -> Dict[str, str]:
    if scopes is None:
      scopes = ALL_SCOPES
    intended_audience = urllib.parse.urlparse(url).hostname
    scope_string = ' '.join(scopes)
    if intended_audience not in self._tokens:
      self._tokens[intended_audience] = {}
    if scope_string not in self._tokens[intended_audience]:
      token = self.issue_token(intended_audience, scopes)
    else:
      token = self._tokens[intended_audience][scope_string]
    payload = jwt.decode(token, options={'verify_signature': False})
    expires = EPOCH + datetime.timedelta(seconds=payload['exp'])
    if datetime.datetime.utcnow() > expires - TOKEN_REFRESH_MARGIN:
      token = self.issue_token(intended_audience, scopes)
    self._tokens[intended_audience][scope_string] = token
    return {'Authorization': 'Bearer ' + token}

  def add_headers(self, request: requests.PreparedRequest, scopes: List[str]):
    for k, v in self.get_headers(request.url, scopes).items():
      request.headers[k] = v

  def get_sub(self) -> Optional[str]:
    """Retrieve `sub` claim from one of the existing tokens"""
    for _, tokens_by_scope in self._tokens.items():
      for token in tokens_by_scope.values():
        payload = jwt.decode(token, options={'verify_signature': False})
        if 'sub' in payload:
          return payload['sub']
    return None


class UTMClientSession(requests.Session):
  """Requests session that enables easy access to ASTM-specified UTM endpoints.

  Automatically applies authorization according to the `auth_adapter` specified
  at construction, when present.

  If the URL starts with '/', then automatically prefix the URL with the
  `prefix_url` specified on construction (this is usually the base URL of the
  DSS).
  """

  def __init__(self, prefix_url: str, auth_adapter: Optional[AuthAdapter] = None):
    super().__init__()

    self._prefix_url = prefix_url[0:-1] if prefix_url[-1] == '/' else prefix_url
    self.auth_adapter = auth_adapter
    self.default_scopes = None

  # Overrides method on requests.Session
  def prepare_request(self, request, **kwargs):
    # Automatically prefix any unprefixed URLs
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url

    return super().prepare_request(request, **kwargs)

  def adjust_request_kwargs(self, kwargs):
    if self.auth_adapter:
      scopes = None
      if 'scopes' in kwargs:
        scopes = kwargs['scopes']
        del kwargs['scopes']
      if 'scope' in kwargs:
        scopes = [kwargs['scope']]
        del kwargs['scope']
      if scopes is None:
        scopes = self.default_scopes
      def auth(prepared_request: requests.PreparedRequest) -> requests.PreparedRequest:
        if not scopes:
          raise ValueError('All tests must specify auth scope for all session requests.  Either specify as an argument for each individual HTTP call, or decorate the test with @default_scope.')
        self.auth_adapter.add_headers(prepared_request, scopes)
        return prepared_request
      kwargs['auth'] = auth
    return kwargs

  def request(self, method, url, **kwargs):
    if 'auth' not in kwargs:
      kwargs = self.adjust_request_kwargs(kwargs)

    return super().request(method, url, **kwargs)

class AsyncUTMTestSession:
  """
  Requests Asyncio client session that provides additional functionality for running DSS concurrency tests:
    * Adds a prefix to URLs that start with a '/'.
    * Automatically applies authorization according to adapter, when present
  """

  def __init__(self, prefix_url: str, auth_adapter: Optional[AuthAdapter] = None):
    self._client = None
    loop = asyncio.get_event_loop()
    loop.run_until_complete(self.build_session())

    self._prefix_url = prefix_url[0:-1] if prefix_url[-1] == '/' else prefix_url
    self.auth_adapter = auth_adapter
    self.default_scopes = None

  async def build_session(self):
    self._client = ClientSession()

  def close(self):
    loop = asyncio.get_event_loop()
    loop.run_until_complete(self._client.close())
  
  def adjust_request_kwargs(self, url, method, kwargs):
    if self.auth_adapter:
      scopes = None
      if 'scopes' in kwargs:
        scopes = kwargs['scopes']
        del kwargs['scopes']
      if 'scope' in kwargs:
        scopes = [kwargs['scope']]
        del kwargs['scope']
      if scopes is None:
        scopes = self.default_scopes
      if not scopes:
        raise ValueError('All tests must specify auth scope for all session requests.  Either specify as an argument for each individual HTTP call, or decorate the test with @default_scope.')
      headers = {}
      for k, v in self.auth_adapter.get_headers(url, scopes).items():
        headers[k] = v
      kwargs['headers'] = headers
      if method == 'PUT' and kwargs.get('data'):
        kwargs['json'] = kwargs['data']
        del kwargs['data']
    return kwargs

  async def put(self, url, **kwargs):
    url = self._prefix_url + url
    if 'auth' not in kwargs:
      kwargs = self.adjust_request_kwargs(url, 'PUT', kwargs)
    async with self._client.put(url, **kwargs) as response:
      return response.status, await response.json()
  
  async def get(self, url, **kwargs):
    url = self._prefix_url + url
    if 'auth' not in kwargs:
      kwargs = self.adjust_request_kwargs(url, 'GET', kwargs)
    async with self._client.get(url, **kwargs) as response:
      return response.status, await response.json()
  
  async def post(self, url, **kwargs):
    url = self._prefix_url + url
    if 'auth' not in kwargs:
      kwargs = self.adjust_request_kwargs(url, 'POST', kwargs)
    async with self._client.post(url, **kwargs) as response:
      return response.status, await response.json()
  
  async def delete(self, url, **kwargs):
    url = self._prefix_url + url
    if 'auth' not in kwargs:
      kwargs = self.adjust_request_kwargs(url, 'DELETE', kwargs)
    async with self._client.delete(url, **kwargs) as response:
      return response.status, await response.json()


def default_scopes(scopes: List[str]):
  """Decorator for tests that modifies UTMClientSession args to use scopes.

  A test function decorated with this decorator will modify all arguments which
  are UTMClientSessions to set their default_scopes to the scopes specified in
  this decorator (and restore the original default_scopes afterward).

  :param scopes: List of scopes to retrieve (by default) for tokens used to
    authorize requests sent using any of the UTMClientSession arguments to the
    decorated test.
  """
  def decorator_default_scope(func):
    @functools.wraps(func)
    def wrapper_default_scope(*args, **kwargs):
      # Change <UTMClientSession>.default_scopes to scopes for all UTMClientSession arguments
      old_scopes = []
      for arg in args:
        if isinstance(arg, UTMClientSession) or isinstance(arg, AsyncUTMTestSession):
          old_scopes.append(arg.default_scopes)
          arg.default_scopes = scopes
      for k, v in kwargs.items():
        if isinstance(v, UTMClientSession) or isinstance(v, AsyncUTMTestSession):
          old_scopes.append(v.default_scopes)
          v.default_scopes = scopes

      result = func(*args, **kwargs)

      # Restore original values of <UTMClientSession>.default_scopes for all UTMClientSession arguments
      for arg in args:
        if isinstance(arg, UTMClientSession) or isinstance(arg, AsyncUTMTestSession):
          arg.default_scopes = old_scopes[0]
          old_scopes = old_scopes[1:]
      for k, v in kwargs.items():
        if isinstance(v, UTMClientSession) or isinstance(v, AsyncUTMTestSession):
          v.default_scopes = old_scopes[0]
          old_scopes = old_scopes[1:]

      return result
    return wrapper_default_scope
  return decorator_default_scope


def default_scope(scope: str):
  """Decorator for tests that modifies UTMClientSession args to use a scope.

    A test function decorated with this decorator will modify all arguments which
    are UTMClientSessions to set their default_scopes to the scope specified in
    this decorator (and restore the original default_scopes afterward).

    :param scopes: Single scope to retrieve (by default) for tokens used to
      authorize requests sent using any of the UTMClientSession arguments to the
      decorated test.
    """
  return default_scopes([scope])


def get_token_claims(headers: Dict) -> Dict:
  auth_key = [key for key in headers if key.lower() == 'authorization']
  if len(auth_key) == 0:
    return {'error': 'Missing Authorization header'}
  if len(auth_key) > 1:
    return {'error': 'Multiple Authorization headers: ' + ', '.join(auth_key)}
  token: str = headers[auth_key[0]]
  if token.lower().startswith('bearer '):
    token = token[len('bearer '):]
  try:
    return jwt.decode(token, options={'verify_signature': False})
  except ValueError as e:
    return {'error': 'ValueError: ' + str(e)}
  except jwt.exceptions.DecodeError as e:
    return {'error': 'DecodeError: ' + str(e)}


class KMLGenerationSession(requests.Session):
  """
  Requests session that provides additional functionality for generating KMLs:
    * Adds a prefix to URLs that start with a '/'.
  """

  def __init__(self, prefix_url: str, kml_folder: str):
    super().__init__()

    self._prefix_url = prefix_url[0:-1] if prefix_url[-1] == '/' else prefix_url
    self.kml_folder = kml_folder

  # Overrides method on requests.Session
  def prepare_request(self, request, **kwargs):
    # Automatically prefix any unprefixed URLs
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url

    return super().prepare_request(request, **kwargs)
