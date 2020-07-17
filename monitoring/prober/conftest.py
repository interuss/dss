import copy
import uuid

from . import infrastructure

import pytest


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
    auth_adapter = infrastructure.ServiceAccountAuthAdapter(
        oauth_token_endpoint,
        pytestconfig.getoption(prefix + '_service_account_json'))
  elif pytestconfig.getoption(prefix + '_username') is not None:
    auth_adapter = infrastructure.UsernamePasswordAuthAdapter(
        oauth_token_endpoint,
        pytestconfig.getoption(prefix + '_username'),
        pytestconfig.getoption(prefix + '_password'),
        pytestconfig.getoption(prefix + '_client_id'))
  elif pytestconfig.getoption('use_dummy_oauth') is not None:
    auth_adapter = infrastructure.DummyOAuthServerAdapter(oauth_token_endpoint, dummy_oauth_sub)
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
  s = infrastructure.DSSTestSession(dss_endpoint + api_version_role, auth_adapter)
  return s


@pytest.fixture(scope='session')
def aux_session(pytestconfig):
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss')
  s = infrastructure.DSSTestSession(dss_endpoint + '/aux/v1', auth_adapter)
  return s


@pytest.fixture(scope='session')
def scd_session(pytestconfig):
  scd_dss_endpoint = pytestconfig.getoption('scd_dss_endpoint')
  if scd_dss_endpoint is None:
    return None

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss')
  s = infrastructure.DSSTestSession(scd_dss_endpoint, auth_adapter)
  return s


@pytest.fixture(scope='session')
def scd_session2(pytestconfig):
  scd_dss_endpoint = pytestconfig.getoption('scd_dss_endpoint')
  if scd_dss_endpoint is None:
    return None

  auth_adapter = make_auth_adapter(pytestconfig, 'oauth', 'fake_uss2')
  s = infrastructure.DSSTestSession(scd_dss_endpoint, auth_adapter)
  return s


@pytest.fixture(scope='function')
def no_auth_session(pytestconfig):
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  api_version_role = pytestconfig.getoption('api_version_role', '')
  if dss_endpoint is None:
    raise ValueError('Missing required --dss-endpoint')

  s = infrastructure.DSSTestSession(dss_endpoint + api_version_role, None)
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

@pytest.fixture(scope='module')
def c1_uuid():
  return str(uuid.uuid4())
