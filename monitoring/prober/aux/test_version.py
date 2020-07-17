"""Test version can be queried."""

def test_version(aux_session):
  resp = aux_session.get('/version')
  assert resp.status_code == 200
  assert resp.json()['version']['as_string']
