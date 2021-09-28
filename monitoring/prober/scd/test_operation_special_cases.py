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

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type


OP1_TYPE = register_resource_type(210, 'Operational intent 1')
OP2_TYPE = register_resource_type(211, 'Operational intent 2')
SUB_TYPE = register_resource_type(212, 'Subscription')


@for_api_versions(scd.API_0_3_5)
def test_ensure_clean_workspace_v5(ids, scd_api, scd_session):
  for op_id in map(ids, (OP1_TYPE, OP2_TYPE)):
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


@for_api_versions(scd.API_0_3_15)
def test_ensure_clean_workspace_v15(ids, scd_api, scd_session):
  for op_id in map(ids, (OP1_TYPE, OP2_TYPE)):
    resp = scd_session.get(
      '/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
    if resp.status_code == 200:
      resp = scd_session.delete(
        '/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 200, resp.content
      resp = scd_session.get(
        '/operational_intent_references/{}'.format(op_id), scope=SCOPE_SC)
      assert resp.status_code == 404, resp.content
    elif resp.status_code == 404:
      # As expected.
      pass
    else:
      assert False, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_request_1_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP1_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.delete('/operation_references/{}'.format(ids(OP1_TYPE)))
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_op_request_1_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP1_TYPE)), json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.delete('/operational_intent_references/{}'.format(ids(OP1_TYPE)))
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_request_2_v5(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_2.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(ids(OP2_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_op_request_2_v15(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_2_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP2_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_query_degenerate_polygon_v5(scd_api, scd_session):
  with open('./scd/resources/op_request_3.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_op_query_degenerate_polygon_v15(scd_api, scd_session):
  with open('./scd/resources/op_request_3_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_references/query', json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_5)
@default_scope(SCOPE_SC)
def test_op_query_not_area_too_large_v5(scd_api, scd_session):
  with open('./scd/resources/op_request_4.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_op_query_not_area_too_large_v15(scd_api, scd_session):
  with open('./scd/resources/op_request_4_v15.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_references/query', json=req)
  assert resp.status_code == 200, resp.content


# ID conversion bug exposure
# Reproduces issue #314
@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_id_conversion_bug_v5(ids, scd_api, scd_session):
  sub_uuid = ids(SUB_TYPE)
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
