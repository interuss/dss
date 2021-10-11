from typing import Optional

from monitoring.monitorlib import infrastructure
from monitoring.monitorlib import auth, rid, scd
from monitoring.prober.infrastructure import VersionString, IDFactory

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

def make_session_async(pytestconfig, endpoint_suffix: str, auth_option: Optional[str] = None) -> Optional[infrastructure.AsyncUTMTestSession]:
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    pytest.skip('dss-endpoint option not set')

  auth_adapter = None
  if auth_option:
    auth_spec = pytestconfig.getoption(auth_option)
    if not auth_spec:
      pytest.skip('%s option not set' % auth_option)
    auth_adapter = auth.make_auth_adapter(auth_spec)

  s = infrastructure.AsyncUTMTestSession(dss_endpoint + endpoint_suffix, auth_adapter)
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
def scd_session_async(pytestconfig):
  return make_session_async(pytestconfig, '/dss/v1', 'scd_auth1')


@pytest.fixture(scope='session')
def scd_session2(pytestconfig):
  return make_session(pytestconfig, '/dss/v1', 'scd_auth2')


@pytest.fixture()
def ids(pytestconfig, session, scd_session, scd_session2):
  """Fixture that converts a ResourceType into an ID for that resource.

  This fixture is a function that accepts a ResourceType as the argument and
  returns a UUIDv4-format string containing an ID for that resource.

  See register_resource_type in infrastructure.py for how to create
  ResourceTypes to provide to this fixture, and also the "Resources" section of
  the README.
  """
  sub = None
  if session:
    session.get('/status', scope=rid.SCOPE_READ)
    rid_sub = session.auth_adapter.get_sub()
    if rid_sub:
      sub = rid_sub
  if sub is None and scd_session:
    scd_session.get('/status', scope=scd.SCOPE_SC)
    scd_sub = scd_session.auth_adapter.get_sub()
    if scd_sub:
      sub = scd_sub
  if sub is None and scd_session2:
    scd_session2.get('/status', scope=scd.SCOPE_SC)
    scd2_sub = scd_session2.auth_adapter.get_sub()
    if scd2_sub:
      sub = scd2_sub
  if sub is None:
    sub = 'unknown'
  factory = IDFactory(sub)
  return lambda id_code: factory.make_id(id_code)


@pytest.fixture(scope='function')
def no_auth_session(pytestconfig):
  return make_session(pytestconfig, '/v1/dss')


@pytest.fixture(scope='session')
def scd_api(pytestconfig):
  api = pytestconfig.getoption('scd_api_version')
  if api is None:
    api = scd.API_0_3_5
  return VersionString(api)
