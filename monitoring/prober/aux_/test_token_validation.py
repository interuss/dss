"""Test aux features"""

import pytest

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.auth import DummyOAuth
from ..rid.common import SCOPE_READ as RID_SCOPE_READ


@default_scope(RID_SCOPE_READ)
def test_validate(aux_session):
  resp = aux_session.get('/validate_oauth')
  assert resp.status_code == 200


@default_scope(RID_SCOPE_READ)
def test_validate_token_good_user(aux_session):
  if not isinstance(aux_session.auth_adapter, DummyOAuth):
    pytest.skip('User ID is not known for general auth providers')
  resp = aux_session.get('/validate_oauth?owner=fake_uss')
  assert resp.status_code == 200


@default_scope(RID_SCOPE_READ)
def test_validate_token_bad_user(aux_session):
  resp = aux_session.get('/validate_oauth?owner=bad_user')
  assert resp.status_code == 403
