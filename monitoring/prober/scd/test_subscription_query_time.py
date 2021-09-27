"""Strategic conflict detection Subscription put query tests:

  - query with different time formats.
"""

import datetime

from monitoring.monitorlib.infrastructure import default_scope
from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import SCOPE_SC
from monitoring.prober.infrastructure import for_api_versions, register_resource_type

BASE_URL = 'https://example.com/uss'

SUB_TYPE = register_resource_type(219, 'Subscription')


def _make_sub_req(time_start, time_end, alt_start, alt_end, radius, scd_api):
  req = {
    "extents": scd.make_vol4(time_start, time_end, alt_start, alt_end, scd.make_circle(-56, 178, radius)),
    "old_version": 0,
    "uss_base_url": BASE_URL,

    "notify_for_constraints": False
  }
  if scd_api == scd.API_0_3_5:
    req["notify_for_operations"] = True
  elif scd_api == scd.API_0_3_15:
    req["notify_for_operational_intents"] = True
  return req


@for_api_versions(scd.API_0_3_5, scd.API_0_3_15)
@default_scope(SCOPE_SC)
def test_subscription_with_invalid_start_time(ids, scd_api, scd_session):
  if scd_session is None:
    return

  time_start = datetime.datetime.utcnow()
  time_end = time_start + datetime.timedelta(hours=2.5)
  req = _make_sub_req(time_start, time_end, 200, 1000, 500, scd_api)
  req['extents']['time_start']['value'] = 'something-invalid'

  resp = scd_session.put('/subscriptions/{}'.format(ids(SUB_TYPE)), json=req)
  assert resp.status_code == 400, resp.content
