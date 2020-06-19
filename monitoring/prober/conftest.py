import copy
import requests
import urllib.parse
import uuid
import traceback
from typing import Dict, List

from google.auth.transport import requests as google_requests
from google.oauth2 import service_account
import pytest

SCOPES = [
    'dss.write.identification_service_areas',
    'dss.read.identification_service_areas',
    'utm.strategic_coordination',
    'utm.constraint_management',
    'utm.constraint_consumption'
]


class AuthAdapter(object):
  """Base class for requests adapters that add JWTs to requests."""

  def __init__(self):
    self._tokens = {}

  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    """Subclasses must return a bearer token for the given audience."""

    raise NotImplementedError()

  def get_headers(self, url: str) -> Dict[str, str]:
    intended_audience = urllib.parse.urlparse(url).hostname
    if intended_audience not in self._tokens:
      self._tokens[intended_audience] = self.issue_token(intended_audience, SCOPES)
    token = self._tokens[intended_audience]
    return {'Authorization': 'Bearer ' + token}

  def add_headers(self, request: requests.Request, **kwargs):
    for k, v in self.get_headers(request.url).items():
      request.headers[k] = v


class DummyOAuthServerAdapter(AuthAdapter):
  """Requests adapter that gets JWTs that uses the Dummy OAuth Server"""

  def __init__(self, token_endpoint: str, sub: str):
    super().__init__()

    oauth_session = requests.Session()

    self._oauth_token_endpoint = token_endpoint
    self._sub = sub
    self._oauth_session = oauth_session

  def issue_token(self, intended_audience: str, scopes: List[str]) -> str:
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}&issuer=dummy&sub={}'.format(
        self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
        urllib.parse.quote(intended_audience), self._sub)
    response = self._oauth_session.post(url).json()
    return response['access_token']

class ServiceAccountAuthAdapter(AuthAdapter):
  """Requests adapter that gets JWTs using a service account."""

  def __init__(self, token_endpoint, service_account_json):
    super().__init__()

    credentials = service_account.Credentials.from_service_account_file(
        service_account_json).with_scopes(['email'])
    oauth_session = google_requests.AuthorizedSession(credentials)

    self._oauth_token_endpoint = token_endpoint
    self._oauth_session = oauth_session

  def issue_token(self, intended_audience, scopes):
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}'.format(
        self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
        urllib.parse.quote(intended_audience))
    response = self._oauth_session.post(url).json()
    return response['access_token']


class UsernamePasswordAuthAdapter(AuthAdapter):
  """Requests adapter that gets JWTs using a username and password."""

  def __init__(self, token_endpoint, username, password, client_id):
    super().__init__()

    self._oauth_token_endpoint = token_endpoint
    self._username = username
    self._password = password
    self._client_id = client_id

  def issue_token(self, intended_audience, scopes):
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

  # Overrides methods on requests.Session
  def prepare_request(self, request, **kwargs):
    # Automatically prefix any unprefixed URLs
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url

    # Automatically add auth header if auth adapter exists
    if self._auth_adapter:
      for k, v in self._auth_adapter.get_headers(request.url).items():
        request.headers[k] = v

    return super().prepare_request(request, **kwargs)

  def issue_token(self, scopes):
    intended_audience = urllib.parse.urlparse(self._prefix_url).hostname
    return self._auth_adapter.issue_token(intended_audience, scopes)


def pytest_addoption(parser):
  parser.addoption('--api-version-role')
  parser.addoption('--dss-endpoint')
  parser.addoption('--scd-dss-endpoint')
  parser.addoption('--oauth-token-endpoint')

  parser.addoption('--oauth-service-account-json')

  parser.addoption('--oauth-username')
  parser.addoption('--oauth-password')
  parser.addoption('--oauth-client-id')

  parser.addoption('--oauth2-service-account-json')

  parser.addoption('--oauth2-username')
  parser.addoption('--oauth2-password')
  parser.addoption('--oauth2-client-id')

  parser.addoption('--use-dummy-oauth')


def make_auth_adapter(pytestconfig, prefix: str, dummy_oauth_sub: str):
  oauth_token_endpoint = pytestconfig.getoption('oauth_token_endpoint')

  # Create an auth adapter to get JWTs using the given credentials.  We can use
  # either a service account, a username/password/client_id or a dummy oauth server.
  if pytestconfig.getoption(prefix + '_service_account_json') is not None:
    auth_adapter = ServiceAccountAuthAdapter(oauth_token_endpoint,
                                             pytestconfig.getoption(prefix + '_service_account_json'))
  elif pytestconfig.getoption(prefix + '_username') is not None:
    auth_adapter = UsernamePasswordAuthAdapter(oauth_token_endpoint,
                                               pytestconfig.getoption(prefix + '_username'),
                                               pytestconfig.getoption(prefix + '_password'),
                                               pytestconfig.getoption(prefix + '_client_id'))
  elif pytestconfig.getoption('use_dummy_oauth') is not None:
    auth_adapter = DummyOAuthServerAdapter(oauth_token_endpoint, dummy_oauth_sub)
  else:
    raise ValueError(
      'You must provide either an OAuth service account, or a username, '
      'password and client ID')

  return auth_adapter


@pytest.fixture(scope='session')
def session(pytestconfig):
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')
  api_version_role = pytestconfig.getoption('api_version_role', '')

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss')
  s = DSSTestSession(dss_endpoint + api_version_role, auth_adapter)
  return s


@pytest.fixture(scope='session')
def scd_session(pytestconfig):
  scd_dss_endpoint = pytestconfig.getoption('scd_dss_endpoint')
  if scd_dss_endpoint is None:
    return None

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss')
  s = DSSTestSession(scd_dss_endpoint, auth_adapter)
  return s


@pytest.fixture(scope='session')
def scd_session2(pytestconfig):
  scd_dss_endpoint = pytestconfig.getoption('scd_dss_endpoint')
  if scd_dss_endpoint is None:
    return None

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss2')
  s = DSSTestSession(scd_dss_endpoint, auth_adapter)
  return s


@pytest.fixture(scope='function')
def rogue_session(pytestconfig):
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  api_version_role = pytestconfig.getoption('api_version_role', '')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')

  s = DSSTestSession(dss_endpoint + api_version_role, None)
  return s


@pytest.fixture(scope='module')
def isa1_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='function')
def isa2_uuid():
  # short lived as this uuid used to test failure cases
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def sub1_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def sub2_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def sub3_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def op1_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def op2_uuid():
  return str(uuid.uuid4())
