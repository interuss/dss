"""Test version can be queried."""

def test_version(session):
  resp = session.get('/version')
  print(resp)
  assert resp.status_code == 200