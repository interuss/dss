from typing import Callable, Optional

from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib import auth, rid, scd
from monitoring.prober.infrastructure import add_test_result, IDFactory, ResourceType, VersionString

import pytest


OPT_RID_AUTH = 'rid_auth'
OPT_SCD_AUTH1 = 'scd_auth1'
OPT_SCD_AUTH2 = 'scd_auth2'

BASE_URL_RID = '/v1/dss'
BASE_URL_SCD = '/dss/v1'
BASE_URL_AUX = '/aux/v1'


def pytest_addoption(parser):
  parser.addoption('--dss-endpoint')

  parser.addoption('--rid-auth')

  parser.addoption('--scd-auth1')
  parser.addoption('--scd-auth1-cp')
  parser.addoption('--scd-auth1-cm')

  parser.addoption('--scd-auth2')

  parser.addoption('--scd-api-version')


@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
  outcome = yield
  result = outcome.get_result()

  if result.when == 'call':
    add_test_result(item, result)


def _bool_value_of(pytestconfig, flag: str, default_value: bool) -> bool:
  value = pytestconfig.getoption(flag)
  if value is None:
    return default_value
  if isinstance(value, bool):
    return value
  if default_value == True and value.lower() == 'false':
    return False
  if default_value == False and value.lower() == 'true':
    return True
  return default_value


def make_session(pytestconfig, endpoint_suffix: str, auth_option: Optional[str] = None) -> Optional[DSSTestSession]:
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    pytest.skip('dss-endpoint option not set')

  auth_adapter = None
  if auth_option:
    auth_spec = pytestconfig.getoption(auth_option)
    if not auth_spec:
      pytest.skip('%s option not set' % auth_option)
    auth_adapter = auth.make_auth_adapter(auth_spec)

  s = DSSTestSession(dss_endpoint + endpoint_suffix, auth_adapter)
  return s


@pytest.fixture(scope='session')
def session(pytestconfig) -> DSSTestSession:
  return make_session(pytestconfig, BASE_URL_RID, OPT_RID_AUTH)


@pytest.fixture(scope='session')
def aux_session(pytestconfig) -> DSSTestSession:
  return make_session(pytestconfig, BASE_URL_AUX, OPT_RID_AUTH)


@pytest.fixture(scope='session')
def scd_session(pytestconfig) -> DSSTestSession:
  return make_session(pytestconfig, BASE_URL_SCD, OPT_SCD_AUTH1)


@pytest.fixture(scope='session')
def scd_session_cp(pytestconfig) -> bool:
  """True iff SCD auth1 user is authorized for constraint processing"""
  return _bool_value_of(pytestconfig, 'scd_auth1_cp', True)


@pytest.fixture(scope='session')
def scd_session_cm(pytestconfig) -> bool:
  """True iff SCD auth1 user is authorized for constraint management"""
  return _bool_value_of(pytestconfig, 'scd_auth1_cm', True)


@pytest.fixture(scope='session')
def scd_session2(pytestconfig) -> DSSTestSession:
  return make_session(pytestconfig, BASE_URL_SCD, OPT_SCD_AUTH2)


@pytest.fixture()
def subscriber(pytestconfig) -> Optional[str]:
  """Subscriber of USS making UTM API calls"""
  if pytestconfig.getoption(OPT_RID_AUTH):
    session = make_session(pytestconfig, BASE_URL_RID, OPT_RID_AUTH)
    session.get('/status', scope=rid.SCOPE_READ)
    rid_sub = session.auth_adapter.get_sub()
    if rid_sub:
      return rid_sub
  if pytestconfig.getoption(OPT_SCD_AUTH1):
    scd_session = make_session(pytestconfig, BASE_URL_SCD, OPT_SCD_AUTH1)
    scd_session.get('/status', scope=scd.SCOPE_SC)
    scd_sub = scd_session.auth_adapter.get_sub()
    if scd_sub:
      return scd_sub
  if pytestconfig.getoption(OPT_SCD_AUTH2):
    scd_session2 = make_session(pytestconfig, BASE_URL_SCD, OPT_SCD_AUTH2)
    scd_session2.get('/status', scope=scd.SCOPE_SC)
    scd2_sub = scd_session2.auth_adapter.get_sub()
    if scd2_sub:
      return scd2_sub
  return None


@pytest.fixture()
def ids(pytestconfig, subscriber) -> Callable[[ResourceType], str]:
  """Fixture that converts a ResourceType into an ID for that resource.

  This fixture is a function that accepts a ResourceType as the argument and
  returns a UUIDv4-format string containing an ID for that resource.

  See register_resource_type in infrastructure.py for how to create
  ResourceTypes to provide to this fixture, and also the "Resources" section of
  the README.
  """
  sub = subscriber
  if sub is None:
    sub = 'unknown'
  factory = IDFactory(sub)
  return lambda id_code: factory.make_id(id_code)


@pytest.fixture(scope='function')
def no_auth_session(pytestconfig) -> DSSTestSession:
  return make_session(pytestconfig, BASE_URL_RID)


@pytest.fixture(scope='session')
def scd_api(pytestconfig) -> str:
  api = pytestconfig.getoption('scd_api_version')
  if api is None:
    api = scd.API_0_3_5
  return VersionString(api)
