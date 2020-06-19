import functools
import requests
from typing import Dict, List
import urllib.parse

from google.auth.transport import requests as google_requests
from google.oauth2 import service_account

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


class DummyOAuthServerAdapter(AuthAdapter):
  """Auth adapter that gets JWTs that uses the Dummy OAuth Server"""

  def __init__(self, token_endpoint: str, sub: str):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._sub = sub
    self._oauth_session = requests.Session()

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}&issuer=dummy&sub={}'.format(
      self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
      urllib.parse.quote(intended_audience), self._sub)
    response = self._oauth_session.post(url).json()
    return response['access_token']

class ServiceAccountAuthAdapter(AuthAdapter):
  """Auth adapter that gets JWTs using a service account."""

  def __init__(self, token_endpoint, service_account_json):
    super().__init__()

    credentials = service_account.Credentials.from_service_account_file(
      service_account_json).with_scopes(['email'])
    oauth_session = google_requests.AuthorizedSession(credentials)

    self._oauth_token_endpoint = token_endpoint
    self._oauth_session = oauth_session

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}'.format(
      self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
      urllib.parse.quote(intended_audience))
    response = self._oauth_session.post(url).json()
    return response['access_token']


class UsernamePasswordAuthAdapter(AuthAdapter):
  """Auth adapter that gets JWTs using a username and password."""

  def __init__(self, token_endpoint, username, password, client_id):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._username = username
    self._password = password
    self._client_id = client_id

  # Overrides method in AuthAdapter
  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    scopes.append('aud:{}'.format(intended_audience))
    response = requests.post(self._oauth_token_endpoint, data={
      'grant_type': "password",
      'username': self._username,
      'password': self._password,
      'client_id': self._client_id,
      'scope': ' '.join(scopes),
    }).json()
    return response['access_token']


class DSSTestSession(requests.Session):
  """
  Requests session that provides additional functionality for DSS tests:
    * Adds a prefix to URLs that start with a '/'.
    * Automatically applies authorization according to adapter, when present
  """

  def __init__(self, prefix_url: str, auth_adapter: AuthAdapter = None):
    super().__init__()

    self._prefix_url = prefix_url
    self._auth_adapter = auth_adapter
    self.default_scopes = ALL_SCOPES

  # Overrides method on requests.Session
  def prepare_request(self, request, **kwargs):
    # Automatically prefix any unprefixed URLs
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url

    return super().prepare_request(request, **kwargs)

  def request(self, method, url, **kwargs):
    if 'auth' not in kwargs and self._auth_adapter:
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
        self._auth_adapter.add_headers(prepared_request, scopes)
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
