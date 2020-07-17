"""Test Authentication validation
"""

from ..infrastructure import default_scope


@default_scope('dss.read.identification_service_areas')
def test_validate(aux_session):
  resp = aux_session.get('/validate_oauth')
  assert resp.status_code == 200


@default_scope('dss.read.identification_service_areas')
def test_validate_token_good_user(aux_session):
  # TODO: These tests should work with any AuthAdapter, not just the dummy OAuth.  Hardcoding fake_uss here won't work in general.
  resp = aux_session.get('/validate_oauth?owner=fake_uss')
  assert resp.status_code == 200


@default_scope('dss.read.identification_service_areas')
def test_validate_token_bad_user(aux_session):
  resp = aux_session.get('/validate_oauth?owner=bad_user')
  assert resp.status_code == 403
