import uuid
from typing import Optional

from . import auth, infrastructure

import pytest


def pytest_addoption(parser):
  parser.addoption('--dss-endpoint')

  parser.addoption('--rid-auth')
  parser.addoption('--scd-auth1')
  parser.addoption('--scd-auth2')


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


@pytest.fixture(scope='function')
def no_auth_session(pytestconfig):
  return make_session(pytestconfig, '/v1/dss')
