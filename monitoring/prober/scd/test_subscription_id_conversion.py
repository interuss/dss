"""ID conversion bug exposure
Reproduces issue #314
"""

import datetime
import uuid

def test_put_sub1(scd_session):
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
            { "lng": -91.4663314819336, "lat": 41.69329603398001 },
            { "lng": -91.49723052978516, "lat": 41.70085834502109 }
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
