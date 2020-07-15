"""Test version can be queried."""

def test_version(session):
  resp = session.get('/version')
  assert resp.status_code == 200
  assert resp.json()['version']['as_string']