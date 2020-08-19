"""Test version can be queried."""

from ..rid.common import SCOPE_READ

def test_version(aux_session):
  resp = aux_session.get('/version', scope=SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['version']['as_string']
  assert version
  assert '0.0.0-undefined' not in version, version
