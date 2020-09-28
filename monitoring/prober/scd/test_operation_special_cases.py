"""Basic Operation tests:

  - make sure the Operation doesn't exist with get or query
  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import datetime
import json
import uuid

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib.scd import SCOPE_SC

OP1_ID = '00000020-b6ee-4082-b6e7-75eb4f000000'
OP2_ID = '00000000-ee51-4700-873d-e10911000000'


def test_ensure_clean_workspace(scd_session):
  for op_id in (OP1_ID, OP2_ID):
    resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
      resp = scd_session.get('/operation_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 404, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_request_1(scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(OP1_ID), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.delete('/operation_references/{}'.format(OP1_ID))
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_request_2(scd_session):
  with open('./scd/resources/op_request_2.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(OP2_ID), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_query_degenerate_polygon(scd_session):
  with open('./scd/resources/op_request_3.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_query_not_area_too_large(scd_session):
  with open('./scd/resources/op_request_4.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 200, resp.content


# ID conversion bug exposure
# Reproduces issue #314
@default_scope(SCOPE_SC)
def test_id_conversion_bug(scd_session):
  sub_uuid = uuid.uuid4()
  time_ref = datetime.datetime.utcnow() + datetime.timedelta(days=1)
  time_start = datetime.datetime(time_ref.year, time_ref.month, time_ref.day, 1, 30)
  time_end = datetime.datetime(time_ref.year, time_ref.month, time_ref.day, 22, 15)
  req = {
    "extents": {
      "volume": {
        "outline_polygon": {
          "vertices": [
            { "lng": -91.49723052978516, "lat": 41.70085834502109 },
            { "lng": -91.50341033935547, "lat": 41.6770148220322 },
            { "lng": -91.47989273071289, "lat": 41.67509157220958 },
            { "lng": -91.4663314819336, "lat": 41.69329603398001 }
          ]
        },
        "altitude_upper": {"units": "M", "reference": "W84", "value": 764.79037},
        "altitude_lower": {"units": "M", "reference": "W84", "value": 23.24352}
      },
      "time_start": {"value": time_start.isoformat() + "Z", "format": "RFC3339"},
      "time_end": {"value": time_end.isoformat() + "Z", "format": "RFC3339"}
    },
    "old_version": 0,
    "uss_base_url": "http://localhost:12012/services/uss/public/uss/v1/",
    "notify_for_constraints": True
  }
  resp = scd_session.put('/subscriptions/{}'.format(sub_uuid), json=req)
  assert resp.status_code == 200, resp.content

  req["extents"]["time_start"]["value"] = (time_start + datetime.timedelta(hours=1)).isoformat() + "Z"
  req["old_version"] = 1
  resp = scd_session.put('/subscriptions/{}'.format(sub_uuid), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.delete('/subscriptions/{}'.format(sub_uuid))
  assert resp.status_code == 200, resp.content
