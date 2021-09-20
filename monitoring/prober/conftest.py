from typing import Optional

from monitoring.monitorlib import infrastructure
from monitoring.monitorlib import auth, scd
from monitoring.prober.infrastructure import VersionString

import pytest


def pytest_addoption(parser):
  parser.addoption('--dss-endpoint')

  parser.addoption('--rid-auth')
  parser.addoption('--scd-auth1')
  parser.addoption('--scd-auth2')
  parser.addoption('--test-owner')
  parser.addoption('--scd-api-version')


def make_session(pytestconfig, endpoint_suffix: str, auth_option: Optional[str] = None) -> Optional[infrastructure.DSSTestSession]:
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    pytest.skip('dss-endpoint option not set')

  auth_adapter = None
  if auth_option:
    auth_spec = pytestconfig.getoption(auth_option)
    if not auth_spec:
      pytest.skip('%s option not set' % auth_option)
    auth_adapter = auth.make_auth_adapter(auth_spec)

  s = infrastructure.DSSTestSession(dss_endpoint + endpoint_suffix, auth_adapter)
  return s


@pytest.fixture(scope='session')
def session(pytestconfig):
  return make_session(pytestconfig, '/v1/dss', 'rid_auth')


@pytest.fixture(scope='session')
def aux_session(pytestconfig):
  return make_session(pytestconfig, '/aux/v1', 'rid_auth')


@pytest.fixture(scope='session')
def scd_session(pytestconfig):
  return make_session(pytestconfig, '/dss/v1', 'scd_auth1')


@pytest.fixture(scope='session')
def scd_session2(pytestconfig):
  return make_session(pytestconfig, '/dss/v1', 'scd_auth2')


@pytest.fixture()
def test_owner(pytestconfig):
  if pytestconfig.getoption('test_owner'):
    return pytestconfig.getoption('test_owner')
  else:
    pytest.exit(
      ValueError("""
      --test-owner required.
      Please follow the instructions in monitoring/prober/README.md to set the value"""))


@pytest.fixture(scope='function')
def no_auth_session(pytestconfig):
  return make_session(pytestconfig, '/v1/dss')

@pytest.fixture(scope='session')
def scd_api(pytestconfig):
  api = pytestconfig.getoption('scd_api_version')
  if api is None:
    api = scd.API_0_3_5
  return VersionString(api)
