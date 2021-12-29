
from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.scd import SCOPE_AA
from monitoring.prober.infrastructure import depends_on

@default_scope(SCOPE_AA)
def test_set_uss_availability(ids, scd_session2):
  resp = scd_session2.put(
    f'/uss_availability/uss1', scope=SCOPE_AA, json={'availability': 'normal'})
  assert resp.status_code == 200, resp.content
  data = resp.json()
  assert data['status']['uss'] == 'uss1'
  assert data['status']['availability'] == 'Normal'
  assert data['version']
  
  resp = scd_session2.put(
    f'/uss_availability/uss1', scope=SCOPE_AA, json={'availability': 'pUrPlE'})
  assert resp.status_code == 400, resp.content


@default_scope(SCOPE_AA)
@depends_on(test_set_uss_availability)
def test_get_uss_availability(ids, scd_session2):
  resp = scd_session2.get(f'/uss_availability/unknown_uss2', scope=SCOPE_AA)
  assert resp.status_code == 200, resp.content
  data = resp.json()
  assert data['status']['availability'] == 'Unknown'
  assert data['version'] == ''

  resp = scd_session2.get(f'/uss_availability/uss1', scope=SCOPE_AA)
  assert resp.status_code == 200, resp.content
  data = resp.json()
  assert data['status']['uss'] == 'uss1'
  assert data['status']['availability'] == 'Normal'
  assert data['version']
