"""Basic subscription tests:

  - create the subscription with a 60 minute expiry
  - get by ID
  - get by search
  - delete
"""

import datetime
import re

import common


def test_sub_does_not_exist(session, sub1_uuid):
    resp = session.get("/subscriptions/{}".format(sub1_uuid))
    assert resp.status_code == 404
    assert resp.json()["message"] == "resource not found: {}".format(sub1_uuid)


def test_create_sub(session, sub1_uuid):
    time_start = datetime.datetime.utcnow()
    time_end = time_start + datetime.timedelta(minutes=60)

    resp = session.put(
        "/subscriptions/{}".format(sub1_uuid),
        json={
            "extents": {
                "spatial_volume": {
                    "footprint": {"vertices": common.VERTICES},
                    "altitude_lo": 20,
                    "altitude_hi": 400,
                },
                "time_start": time_start.strftime(common.DATE_FORMAT),
                "time_end": time_end.strftime(common.DATE_FORMAT),
            },
            "callbacks": {"identification_service_area_url": "https://example.com/foo"},
        },
    )
    assert resp.status_code == 200

    data = resp.json()
    assert data["subscription"]["id"] == sub1_uuid
    assert data["subscription"]["notification_index"] == 0
    assert data["subscription"]["callbacks"] == {
        "identification_service_area_url": "https://example.com/foo"
    }
    assert data["subscription"]["time_start"] == time_start.strftime(common.DATE_FORMAT)
    assert data["subscription"]["time_end"] == time_end.strftime(common.DATE_FORMAT)
    assert re.match(r"[a-z0-9]{10,}$", data["subscription"]["version"])
    assert "service_areas" in data


def test_get_sub_by_id(session, sub1_uuid):
    resp = session.get("/subscriptions/{}".format(sub1_uuid))
    assert resp.status_code == 200

    data = resp.json()
    assert data["subscription"]["id"] == sub1_uuid
    assert data["subscription"]["notification_index"] == 0
    assert data["subscription"]["callbacks"] == {
        "identification_service_area_url": "https://example.com/foo"
    }


def test_get_sub_by_search(session, sub1_uuid):
    resp = session.get("/subscriptions?area={}".format(common.GEO_POLYGON_STRING))
    assert resp.status_code == 200
    assert sub1_uuid in [x["id"] for x in resp.json()["subscriptions"]]


def test_get_sub_by_searching_huge_area(session, sub1_uuid):
    resp = session.get("/subscriptions?area={}".format(common.HUGE_GEO_POLYGON_STRING))
    assert resp.status_code == 413


def test_delete_sub(session, sub1_uuid):
    # GET the sub first to find its version.
    resp = session.get("/subscriptions/{}".format(sub1_uuid))
    assert resp.status_code == 200
    version = resp.json()["subscription"]["version"]

    # Then delete it.
    resp = session.delete("/subscriptions/{}/{}".format(sub1_uuid, version))
    assert resp.status_code == 200


def test_get_deleted_sub_by_id(session, sub1_uuid):
    resp = session.get("/subscriptions/{}".format(sub1_uuid))
    assert resp.status_code == 404


def test_get_deleted_sub_by_search(session, sub1_uuid):
    resp = session.get("/subscriptions?area={}".format(common.GEO_POLYGON_STRING))
    assert resp.status_code == 200
    assert sub1_uuid not in [x["id"] for x in resp.json()["subscriptions"]]
