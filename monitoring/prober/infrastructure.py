import functools
from typing import Dict, List, Optional
import urllib.parse

import requests

ALL_SCOPES = [
  'dss.write.identification_service_areas',
  'dss.read.identification_service_areas',
  'utm.strategic_coordination',
  'utm.constraint_management',
  'utm.constraint_consumption'
]


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
      self._tokens[intended_audience][scope_string] = self.issue_token(intended_audience, scopes)
    token = self._tokens[intended_audience][scope_string]
    return {'Authorization': 'Bearer ' + token}

  def add_headers(self, request: requests.PreparedRequest, scopes: List[str]):
    for k, v in self.get_headers(request.url, scopes).items():
      request.headers[k] = v


class DSSTestSession(requests.Session):
  """
  Requests session that provides additional functionality for DSS tests:
    * Adds a prefix to URLs that start with a '/'.
    * Automatically applies authorization according to adapter, when present
  """

  def __init__(self, prefix_url: str, auth_adapter: Optional[AuthAdapter] = None):
    super().__init__()

    self._prefix_url = prefix_url
    self.auth_adapter = auth_adapter
    self.default_scopes = ALL_SCOPES

  # Overrides method on requests.Session
  def prepare_request(self, request, **kwargs):
    # Automatically prefix any unprefixed URLs
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url

    return super().prepare_request(request, **kwargs)

  def request(self, method, url, **kwargs):
    if 'auth' not in kwargs and self.auth_adapter:
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
        self.auth_adapter.add_headers(prepared_request, scopes)
        return prepared_request
      kwargs['auth'] = auth

    return super().request(method, url, **kwargs)


def default_scopes(scopes: List[str]):
  """Decorator for tests that modifies DSSTestSession args to use scopes.

  A test function decorated with this decorator will modify all arguments which
  are DSSTestSessions to set their default_scopes to the scopes specified in
  this decorator (and restore the original default_scopes afterward).

  :param scopes: List of scopes to retrieve (by default) for tokens used to
    authorize requests sent using any of the DSSTestSession arguments to the
    decorated test.
  """
  def decorator_default_scope(func):
    @functools.wraps(func)
    def wrapper_default_scope(*args, **kwargs):
      # Change <DSSTestSession>.default_scopes to scopes for all DSSTestSession arguments
      old_scopes = []
      for arg in args:
        if isinstance(arg, DSSTestSession):
          old_scopes.append(arg.default_scopes)
          arg.default_scopes = scopes
      for k, v in kwargs.items():
        if isinstance(v, DSSTestSession):
          old_scopes.append(v.default_scopes)
          v.default_scopes = scopes

      result = func(*args, **kwargs)

      # Restore original values of <DSSTestSession>.default_scopes for all DSSTestSession arguments
      for arg in args:
        if isinstance(arg, DSSTestSession):
          arg.default_scopes = old_scopes[0]
          old_scopes = old_scopes[1:]
      for k, v in kwargs.items():
        if isinstance(v, DSSTestSession):
          v.default_scopes = old_scopes[0]
          old_scopes = old_scopes[1:]

      return result
    return wrapper_default_scope
  return decorator_default_scope


def default_scope(scope: str):
  """Decorator for tests that modifies DSSTestSession args to use a scope.

    A test function decorated with this decorator will modify all arguments which
    are DSSTestSessions to set their default_scopes to the scope specified in
    this decorator (and restore the original default_scopes afterward).

    :param scopes: Single scope to retrieve (by default) for tokens used to
      authorize requests sent using any of the DSSTestSession arguments to the
      decorated test.
    """
  return default_scopes([scope])
