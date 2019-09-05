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
  """Requests adapter that gets JWTs from an OAuth token provider."""

  def __init__(self, oauth_token_endpoint, oauth_session):
    super().__init__()

    self._oauth_token_endpoint = oauth_token_endpoint
    self._oauth_session = oauth_session
    self._tokens = {}

  def _get_token(self, intended_audience):
    if intended_audience not in self._tokens:
      url = '{}?grant_type=client_credentials&scope={}&intended_audience={}'.format(
          self._oauth_token_endpoint, urllib.parse.quote(' '.join(SCOPES)),
          urllib.parse.quote(intended_audience))
      response = self._oauth_session.post(url).json()
      self._tokens[intended_audience] = response['access_token']

    return self._tokens[intended_audience]

  def add_headers(self, request, **kwargs):
    intended_audience = urllib.parse.urlparse(request.url).hostname
    token = self._get_token(intended_audience)
    request.headers['Authorization'] = 'Bearer ' + token


class PrefixURLSession(requests.Session):
  """Requests session that adds a prefix to URLs that start with a '/'."""

  def __init__(self, prefix_url):
    super().__init__()

    self._prefix_url = prefix_url

  def prepare_request(self, request, **kwargs):
    if request.url.startswith('/'):
      request.url = self._prefix_url + request.url
    return super().prepare_request(request, **kwargs)


def pytest_addoption(parser):
  parser.addoption('--service-account-json')
  parser.addoption('--oauth-token-endpoint')
  parser.addoption('--dss-endpoint')


@pytest.fixture(scope='session')
def session(pytestconfig):
  service_account_json = pytestconfig.getoption('service_account_json')
  oauth_token_endpoint = pytestconfig.getoption('oauth_token_endpoint')

  credentials = service_account.Credentials.from_service_account_file(
      service_account_json).with_scopes(['email'])
  oauth_session = google_requests.AuthorizedSession(credentials)

  s = PrefixURLSession(pytestconfig.getoption('dss_endpoint'))
  s.mount('http://', AuthAdapter(oauth_token_endpoint, oauth_session))
  s.mount('https://', AuthAdapter(oauth_token_endpoint, oauth_session))
  return s


@pytest.fixture(scope='module')
def isa1_uuid():
  return str(uuid.uuid4())


@pytest.fixture(scope='module')
def sub1_uuid():
  return str(uuid.uuid4())
