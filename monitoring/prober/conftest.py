import copy
import requests
import urllib.parse
import uuid

from google.auth.transport import requests as google_requests
from google.oauth2 import service_account
import pytest

SCOPES = [
    'dss.write.identification_service_areas',
    'dss.read.identification_service_areas',
]


class AuthAdapter(requests.adapters.HTTPAdapter):
  """Base class for requests adapters that add JWTs to requests."""

  def issue_token(self, intended_audience, scopes):
    """Subclasses must return a bearer token for the given audience."""

    raise NotImplementedError()

  def add_headers(self, request, **kwargs):
    intended_audience = urllib.parse.urlparse(request.url).hostname
    if intended_audience not in self._tokens:
      self._tokens[intended_audience] = self.issue_token(intended_audience, SCOPES)
    token = self._tokens[intended_audience]
    request.headers['Authorization'] = 'Bearer ' + token


class DummyOAuthServerAdapter(AuthAdapter):
  """Requests adapter that gets JWTs that uses the Dummy OAuth Server"""

  def __init__(self, token_endpoint):
    super().__init__()

    oauth_session = requests.Session()

    self._oauth_token_endpoint = token_endpoint
    self._oauth_session = oauth_session
    self._tokens = {}

  def issue_token(self, intended_audience, scopes):
    url = '{}?grant_type=client_credentials&scope={}&intended_audience={}'.format(
        self._oauth_token_endpoint, urllib.parse.quote(' '.join(scopes)),
        urllib.parse.quote(intended_audience))
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
    self._tokens = {}

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
    self._tokens = {}

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


class PrefixURLSession(requests.Session):
  """Requests session that adds a prefix to URLs that start with a '/'."""

  def __init__(self, prefix_url, version_role_prefix):
    super().__init__()

    self._prefix_url = prefix_url
    self._version_role_prefix = version_role_prefix

  def prepare_request(self, request, **kwargs):
    if request.url.startswith('/'):
      request.url = self._prefix_url + self._version_role_prefix + request.url
    return super().prepare_request(request, **kwargs)

  def issue_token(self, scopes):
    adapter = self.get_adapter(self._prefix_url)
    intended_audience = urllib.parse.urlparse(self._prefix_url).hostname
    return adapter.issue_token(intended_audience, scopes)

def pytest_addoption(parser):
  parser.addoption('--api-version-role')
  parser.addoption('--dss-endpoint')
  parser.addoption('--oauth-token-endpoint')

  parser.addoption('--oauth-service-account-json')

  parser.addoption('--oauth-username')
  parser.addoption('--oauth-password')
  parser.addoption('--oauth-client-id')

  parser.addoption('--use-dummy-oauth')


@pytest.fixture(scope='session')
def session(pytestconfig):
  oauth_token_endpoint = pytestconfig.getoption('oauth_token_endpoint')

  # Create an auth adapter to get JWTs using the given credentials.  We can use
  # either a service account, a username/password/client_id or a dummy oauth server.
  if pytestconfig.getoption('oauth_service_account_json') is not None:
    auth_adapter = ServiceAccountAuthAdapter(oauth_token_endpoint,
        pytestconfig.getoption('oauth_service_account_json'))
  elif pytestconfig.getoption('oauth_username') is not None:
    auth_adapter = UsernamePasswordAuthAdapter(oauth_token_endpoint,
        pytestconfig.getoption('oauth_username'),
        pytestconfig.getoption('oauth_password'),
        pytestconfig.getoption('oauth_client_id'))
  elif pytestconfig.getoption('use_dummy_oauth') is not None:
    auth_adapter = DummyOAuthServerAdapter(oauth_token_endpoint)
  else:
    raise ValueError(
        'You must provide either an OAuth service account, or a username, '
        'password and client ID')

  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')
  api_version_role = pytestconfig.getoption('api_version_role', '')

  s = PrefixURLSession(dss_endpoint, api_version_role)
  s.mount('http://', auth_adapter)
  s.mount('https://', auth_adapter)
  return s

@pytest.fixture(scope='function')
def rogue_session(pytestconfig):
  auth_adapter = requests.Session()
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  api_version_role = pytestconfig.getoption('api_version_role', '')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')

  s = PrefixURLSession(dss_endpoint, api_version_role)
  s.mount('http://', auth_adapter)
  s.mount('https://', auth_adapter)
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


@pytest.fixture(scope='function')
def sub2_uuid():
  # short lived as this uuid used to test failure cases
  return str(uuid.uuid4())