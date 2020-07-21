"""Test version can be queried."""

def test_version(aux_session):
  resp = aux_session.get('/version')
  assert resp.status_code == 200
  version = resp.json()['version']['as_string']
  assert version
  assert 'undefined' not in version, version
