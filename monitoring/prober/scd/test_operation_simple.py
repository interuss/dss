"""Basic Operation tests:

  - create the Operation with a 60 minute length
  - get by ID
  - search with earliest_time and latest_time
  - delete
"""

import datetime

from . import common


def test_op_does_not_exist_get(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 404, resp.content


def test_op_does_not_exist_query(scd_session, op1_uuid):
  if scd_session is None:
    return
  time_now = datetime.datetime.utcnow()
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(time_now, time_now, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [op['id'] for op in resp.json().get('operation_references', [])]


def test_create_op_single_extent(scd_session, op1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = scd_session.put(
    '/operation_references/{}'.format(op1_uuid),
    json={
      'extents': common.make_vol4(time_start, time_end, 0, 120, common.make_circle(-56, 178, 50)),
      'old_version': 0,
      'state': 'Accepted',
      'uss_base_url': 'https://example.com/dss',
      'new_subscription': {
        'uss_base_url': 'https://example.com/dss',
        'notify_for_constraints': False
      }
    })
  assert resp.status_code == 400, resp.content


def test_create_op(scd_session, op1_uuid):
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = scd_session.put(
      '/operation_references/{}'.format(op1_uuid),
      json={
          'extents': [common.make_vol4(time_start, time_end, 0, 120, common.make_circle(-56, 178, 50))],
          'old_version': 0,
          'state': 'Accepted',
          'uss_base_url': 'https://example.com/dss',
          'new_subscription': {
              'uss_base_url': 'https://example.com/dss',
              'notify_for_constraints': False
          }
      })
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == op1_uuid
  assert op['uss_base_url'] == 'https://example.com/dss'
  assert op['time_start']['value'] == time_start.strftime(common.DATE_FORMAT)
  assert op['time_end']['value'] == time_end.strftime(common.DATE_FORMAT)
  assert op['version'] == 1
  assert 'state' not in op


def test_get_op_by_id(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content

  data = resp.json()
  op = data['operation_reference']
  assert op['id'] == op1_uuid
  assert op['uss_base_url'] == 'https://example.com/dss'
  assert op['version'] == 1
  assert 'state' not in op


def test_get_op_by_search_missing_params(scd_session):
  resp = scd_session.post('/operation_references/query')
  assert resp.status_code == 400, resp.content


def test_get_op_by_search(scd_session, op1_uuid):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json().get('operation_references', [])]


def test_get_op_by_search_earliest_time_included(scd_session, op1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=59)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json()['operation_references']]


def test_get_op_by_search_earliest_time_excluded(scd_session, op1_uuid):
  earliest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=61)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(earliest_time, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]


def test_get_op_by_search_latest_time_included(scd_session, op1_uuid):
  latest_time = datetime.datetime.utcnow() + datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid in [x['id'] for x in resp.json()['operation_references']]


def test_get_op_by_search_latest_time_excluded(scd_session, op1_uuid):
  latest_time = datetime.datetime.utcnow() - datetime.timedelta(minutes=1)
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, latest_time, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]


def test_delete_op(scd_session, op1_uuid):
  resp = scd_session.delete('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 200, resp.content


def test_get_deleted_op_by_id(scd_session, op1_uuid):
  resp = scd_session.get('/operation_references/{}'.format(op1_uuid))
  assert resp.status_code == 404, resp.content


def test_get_deleted_op_by_search(scd_session, op1_uuid):
  resp = scd_session.post('/operation_references/query', json={
    'area_of_interest': common.make_vol4(None, None, 0, 5000, common.make_circle(-56, 178, 300))
  })
  assert resp.status_code == 200, resp.content
  assert op1_uuid not in [x['id'] for x in resp.json()['operation_references']]

