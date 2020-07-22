"""Basic Operation tests:

  - make sure the Operation doesn't exist with get or query
  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - mutate
  - delete
"""

import json

from ..infrastructure import default_scope
from .common import SCOPE_SC, SCOPE_CI, SCOPE_CM


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_request_1(scd_session):
  with open('./scd/resources/op_request_1.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/2df6b920-b6ee-4082-b6e7-75eb4fde25d1', json=req)
  assert resp.status_code == 200, resp.content

  resp = scd_session.delete('/operation_references/2df6b920-b6ee-4082-b6e7-75eb4fde25d1')
  assert resp.status_code == 200, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_request_2(scd_session, op1_uuid):
  with open('./scd/resources/op_request_2.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.put('/operation_references/{}'.format(op1_uuid), json=req)
  assert resp.status_code == 400, resp.content


# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_query_degenerate_polygon(scd_session):
  with open('./scd/resources/op_request_3.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 200, resp.content

# Preconditions: None
# Mutations: None
@default_scope(SCOPE_SC)
def test_op_query_not_area_too_large(scd_session):
  with open('./scd/resources/op_request_4.json', 'r') as f:
    req = json.load(f)
  resp = scd_session.post('/operation_references/query', json=req)
  assert resp.status_code == 200, resp.content
