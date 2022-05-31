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
from monitoring.prober.scd import actions


OP1_TYPE = register_resource_type(210, 'Operational intent 1')
OP2_TYPE = register_resource_type(211, 'Operational intent 2')
SUB_TYPE = register_resource_type(212, 'Subscription')


@for_api_versions(scd.API_0_3_17)
def test_ensure_clean_workspace(ids, scd_api, scd_session):
  for op_id in map(ids, (OP1_TYPE, OP2_TYPE)):
      actions.delete_operation_if_exists(op_id, scd_session, scd_api)
  actions.delete_subscription_if_exists(ids(SUB_TYPE), scd_session, scd_api)


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_request_1(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP1_TYPE)), json=req)
  assert resp.status_code == 200, resp.content
  data = resp.json()
  assert 'operational_intent_reference'  in data, data
  assert 'ovn' in resp.json()['operational_intent_reference'], data
  ovn = data['operational_intent_reference']['ovn']

  resp = scd_session.delete('/operational_intent_references/{}/{}'.format(ids(OP1_TYPE), ovn))
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_request_2(ids, scd_api, scd_session):
  with open('./scd/resources/op_request_2.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operational_intent_references/{}'.format(ids(OP2_TYPE)), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_query_degenerate_polygon(scd_api, scd_session):
  with open('./scd/resources/op_request_3.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_references/query', json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_op_query_not_area_too_large(scd_api, scd_session):
  with open('./scd/resources/op_request_4.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operational_intent_references/query', json=req)
  assert resp.status_code == 200, resp.content


# ID conversion bug exposure
# Reproduces issue #314
@for_api_versions(scd.API_0_3_17)
@default_scope(SCOPE_SC)
def test_id_conversion_bug(ids, scd_api, scd_session):
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
    "uss_base_url": "https://example.com/uss/v1/",
    "notify_for_constraints": True
  }
  resp = scd_session.put('/subscriptions/{}'.format(sub_uuid), json=req)
  assert resp.status_code == 200, resp.content

  req["extents"]["time_start"]["value"] = (time_start + datetime.timedelta(hours=1)).isoformat() + "Z"
  if scd_api == scd.API_0_3_17:
    resp = scd_session.put('/subscriptions/{}/{}'.format(sub_uuid, resp.json()['subscription']['version']), json=req)
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content

  if scd_api == scd.API_0_3_17:
    resp = scd_session.delete('/subscriptions/{}/{}'.format(sub_uuid, resp.json()['subscription']['version']))
  else:
    raise NotImplementedError('Unsupported API version {}'.format(scd_api))
  assert resp.status_code == 200, resp.content


@for_api_versions(scd.API_0_3_17)
def test_final_cleanup(ids, scd_api, scd_session):
    test_ensure_clean_workspace(ids, scd_api, scd_session)
