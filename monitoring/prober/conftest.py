import argparse
from typing import Callable, Optional

from monitoring.monitorlib.infrastructure import UTMClientSession, AsyncUTMTestSession
from monitoring.monitorlib import auth, rid, scd
from monitoring.prober.infrastructure import add_test_result, IDFactory, ResourceType, VersionString

import pytest


OPT_RID_AUTH = 'rid_auth'
OPT_RID_V2_AUTH = 'rid_v2_auth'
OPT_SCD_AUTH1 = 'scd_auth1'
OPT_SCD_AUTH2 = 'scd_auth2'

BASE_URL_RID = ''
BASE_URL_RID_V2 = '/rid/v2'
BASE_URL_SCD = '/dss/v1'
BASE_URL_AUX = '/aux/v1'


def str2bool(v) -> bool:
  if isinstance(v, bool):
    return v
  if v.lower() in ('yes', 'true', 't', 'y', '1'):
    return True
  elif v.lower() in ('no', 'false', 'f', 'n', '0'):
    return False
  else:
    raise argparse.ArgumentTypeError('Boolean value expected.')


def pytest_addoption(parser):
  parser.addoption(
    '--dss-endpoint',
    help='Base URL of DSS to test',
    metavar='URL',
    dest='dss_endpoint')

  parser.addoption(
    '--rid-auth',
    help='Auth spec (see Authorization section of README.md) for performing remote ID v1 actions in the DSS',
    metavar='SPEC',
    dest='rid_auth')

  parser.addoption(
    '--rid-v2-auth',
    help='Auth spec (see Authorization section of README.md) for performing remote ID v2 actions in the DSS',
    metavar='SPEC',
    dest='rid_v2_auth')

  parser.addoption(
    '--scd-auth1',
    help='Auth spec (see Authorization section of README.md) for performing primary strategic deconfliction actions in the DSS',
    metavar='SPEC',
    dest='scd_auth1')
  parser.addoption(
    '--scd-auth1-cp',
    help='True if the USS specified in scd-auth1 has utm.constraint_processing privileges',
    type=str2bool,
    nargs='?',
    default=True,
    dest='scd_auth1_cp')
  parser.addoption(
    '--scd-auth1-cm',
    help='True if the USS specified in scd-auth1 has utm.constraint_management privileges',
    type=str2bool,
    nargs='?',
    default=True,
    dest='scd_auth1_cm')

  parser.addoption(
    '--scd-auth2',
    help='Auth spec (see Authorization section of README.md) for performing secondary strategic deconfliction actions (like observing primary actions, causing notification generation, etc) in the DSS',
    metavar='SPEC',
    dest='scd_auth2')

  parser.addoption(
    '--scd-api-version',
    help='SCD API version to target',
    choices=[scd.API_1_0_0],
    default=scd.API_1_0_0,
    dest='scd_api_version')


@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item, call):
  outcome = yield
  result = outcome.get_result()

  if result.when == 'call':
    add_test_result(item, result)


def make_session(pytestconfig, endpoint_suffix: str, auth_option: Optional[str] = None) -> Optional[UTMClientSession]:
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    pytest.skip('dss-endpoint option not set')

  auth_adapter = None
  if auth_option:
    auth_spec = pytestconfig.getoption(auth_option)
    if not auth_spec:
      pytest.skip('%s option not set' % auth_option)
    auth_adapter = auth.make_auth_adapter(auth_spec)

  s = UTMClientSession(dss_endpoint + endpoint_suffix, auth_adapter)
  return s

def make_session_async(pytestconfig, endpoint_suffix: str, auth_option: Optional[str] = None) -> Optional[AsyncUTMTestSession]:
  dss_endpoint = pytestconfig.getoption('dss_endpoint')
  if dss_endpoint is None:
    pytest.skip('dss-endpoint option not set')

  auth_adapter = None
  if auth_option:
    auth_spec = pytestconfig.getoption(auth_option)
    if not auth_spec:
      pytest.skip('%s option not set' % auth_option)
    auth_adapter = auth.make_auth_adapter(auth_spec)

  s = AsyncUTMTestSession(dss_endpoint + endpoint_suffix, auth_adapter)
  return s


@pytest.fixture(scope='session')
def session_ridv1(pytestconfig) -> UTMClientSession:
  return make_session(pytestconfig, BASE_URL_RID, OPT_RID_AUTH)


@pytest.fixture(scope='session')
def session_ridv2(pytestconfig) -> UTMClientSession:
    return make_session(pytestconfig, BASE_URL_RID_V2, OPT_RID_V2_AUTH)


@pytest.fixture(scope='session')
def session_ridv1_async(pytestconfig):
  session = make_session_async(pytestconfig, BASE_URL_RID, OPT_RID_AUTH)
  yield session
  session.close()


@pytest.fixture(scope='session')
def aux_session(pytestconfig) -> UTMClientSession:
  return make_session(pytestconfig, BASE_URL_AUX, OPT_RID_AUTH)


@pytest.fixture(scope='session')
def scd_session(pytestconfig) -> UTMClientSession:
  return make_session(pytestconfig, BASE_URL_SCD, OPT_SCD_AUTH1)

@pytest.fixture(scope='session')
def scd_session_async(pytestconfig):
  session = make_session_async(pytestconfig, '/dss/v1', 'scd_auth1')
  yield session
  session.close()


@pytest.fixture(scope='session')
def scd_session_cp(pytestconfig) -> bool:
  """True iff SCD auth1 user is authorized for constraint processing"""
  return pytestconfig.getoption('scd_auth1_cp')


@pytest.fixture(scope='session')
def scd_session_cm(pytestconfig) -> bool:
  """True iff SCD auth1 user is authorized for constraint management"""
  return pytestconfig.getoption('scd_auth1_cm')


@pytest.fixture(scope='session')
def scd_session2(pytestconfig) -> UTMClientSession:
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
def no_auth_session_ridv1(pytestconfig) -> UTMClientSession:
  return make_session(pytestconfig, BASE_URL_RID)


@pytest.fixture(scope='function')
def no_auth_session_ridv2(pytestconfig) -> UTMClientSession:
    return make_session(pytestconfig, BASE_URL_RID_V2)


@pytest.fixture(scope='session')
def scd_api(pytestconfig) -> str:
  api = pytestconfig.getoption('scd_api_version')
  return VersionString(api)
