"""Test aux features"""

import pytest

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import rid
from monitoring.monitorlib.auth import DummyOAuth


@default_scope(rid.SCOPE_READ)
def test_validate(aux_session):
  resp = aux_session.get('/validate_oauth')
  assert resp.status_code == 200


@default_scope(rid.SCOPE_READ)
def test_validate_token_good_user(aux_session, subscriber):
  resp = aux_session.get('/validate_oauth?owner={}'.format(subscriber))
  assert resp.status_code == 200


@default_scope(rid.SCOPE_READ)
def test_validate_token_bad_user(aux_session):
  resp = aux_session.get('/validate_oauth?owner=bad_user')
  assert resp.status_code == 403
