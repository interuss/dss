"""Test version can be queried."""

from monitoring.monitorlib import rid

def test_version(aux_session):
  resp = aux_session.get('/version', scope=rid.SCOPE_READ)
  assert resp.status_code == 200
  version = resp.json()['version']['as_string']
  assert version
  assert 'undefined' not in version, version
