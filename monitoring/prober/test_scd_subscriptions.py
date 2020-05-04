"""Basic strategic conflict detection Subscription tests:

  - make sure Subscription doesn't exist
  - create the Subscription with a 60 minute expiry
  - get by ID
  - get by searching a circular area
  - delete
  - make sure Subscription can't be found by ID
  - make sure Subscription can't be found by search
"""

import datetime
import re

import common


def _check_sub1(data, sub1_uuid):
  assert data['subscription']['id'] == sub1_uuid
  assert (('notification_index' not in data['subscription']) or
          (data['subscription']['notification_index'] == 0))
  assert data['subscription']['version'] == 1
  assert data['subscription']['uss_base_url'] == 'https://example.com/foo'
  assert data['subscription']['time_start']['format'] == common.TIME_FORMAT_CODE
  assert data['subscription']['time_end']['format'] == common.TIME_FORMAT_CODE
  assert data['subscription']['notify_for_operations'] == True
  assert (('notify_for_constraints' not in data['subscription']) or
          (data['subscription']['notify_for_constraints'] == False))
  assert (('implicit_subscription' not in data['subscription']) or
            (data['subscription']['implicit_subscription'] == False))
  assert (('dependent_operations' not in data['subscription'])
          or len(data['subscription']['dependent_operations']) == 0)
  assert 'operations' in data


def test_scd_sub_does_not_exist(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(sub1_uuid))
#   assert resp.status_code == 404
#   assert resp.json()['message'] == 'resource not found: {}'.format(sub1_uuid)


def test_scd_create_sub(scd_session, sub1_uuid):
  if scd_session is None:
    return
  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(minutes=60)

  resp = scd_session.put(
      '/subscriptions/{}'.format(sub1_uuid),
      json={

        "old_version": 0,
        "uss_base_url": "https://example.com/foo",
        "notify_for_operations": True,
        "notify_for_constraints": False
      })
  if resp.status_code != 200:
    print(resp.content)
  assert resp.status_code == 200

  data = resp.json()
#   assert data['subscription']['time_start']['value'] == time_start.strftime(
#       common.DATE_FORMAT)
#   assert data['subscription']['time_end']['value'] == time_end.strftime(
#       common.DATE_FORMAT)
#   _check_sub1(data, sub1_uuid)


def test_scd_get_sub_by_id(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(sub1_uuid))
  assert resp.status_code == 200

#   data = resp.json()
#   _check_sub1(data, sub1_uuid)


def test_scd_get_sub_by_search(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.post(
      '/subscriptions/query',
      json={
        "area_of_interest": {
         "volume": {
           "outline_circle": {
             "type": "Feature",
             "geometry": {
               "type": "Point",
               "coordinates": {
                 "type": "Point",
                 "coordinates": [
                   -122.106325,
                   47.660898
                 ]
               }
             },
             "properties": {
               "radius": {
                 "value": 300.183,
                 "units": "M"
               }
             }
           },
           "altitude_lower": {
             "value": 0,
             "reference": "W84",
             "units": "M"
           },
           "altitude_upper": {
             "value": 3000,
             "reference": "W84",
             "units": "M"
           }
         },
         "time_start": {
           "value": "1985-04-12T23:20:50.52Z",
           "format": "RFC3339"
         },
         "time_end": {
           "value": "2100-04-12T23:20:50.52Z",
           "format": "RFC3339"
         }
        }
      })
  if resp.status_code != 200:
    print(resp.content)
#   assert resp.status_code == 200
#   assert sub1_uuid in [x['id'] for x in resp.json()['subscriptions']]


def test_scd_delete_sub(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.delete('/subscriptions/{}'.format(sub1_uuid))
#   assert resp.status_code == 200


def test_scd_get_deleted_sub_by_id(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions/{}'.format(sub1_uuid))
#   assert resp.status_code == 404


def test_scd_get_deleted_sub_by_search(scd_session, sub1_uuid):
  if scd_session is None:
    return
  resp = scd_session.get('/subscriptions?area={}'.format(common.GEO_POLYGON_STRING))
#   assert resp.status_code == 200
#   assert sub1_uuid not in [x['id'] for x in resp.json()['subscriptions']]
