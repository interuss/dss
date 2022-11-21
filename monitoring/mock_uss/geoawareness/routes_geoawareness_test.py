import uuid

import pytest
from flask import json

from monitoring.mock_uss import webapp
from monitoring.monitorlib.auth import NoAuth

TEST_DATASET_URL = "https://raw.githubusercontent.com/interuss/dss/517595ad4074bdb621feb4ab81c2d2f4fc11eff1/monitoring/uss_qualifier/scenarios/uspace/geo_awareness/design/CHE/geo-awareness-che-1.json"


@pytest.fixture()
def app():
    app = webapp
    app.config.update({"TESTING": True})
    yield app


@pytest.fixture()
def client(app):
    test_client = app.test_client()
    return test_client


@pytest.fixture()
def client_options(app):
    auth = NoAuth()
    token = auth.issue_token("localhost", ["geo-awareness.test"])
    return {"headers": {"Authorization": f"Bearer {token}"}}


def test_status_unauthenticated(client):
    response = client.get("/geoawareness/status")
    assert response.status_code == 401


def test_status(client, client_options):
    response = client.get("/geoawareness/status", **client_options)
    assert response.status_code == 200
    assert json.loads(response.data)["status"] == "Ready"


def test_geosource(client, client_options):
    id = uuid.uuid4()

    # Creation
    response = client.put(
        f"/geoawareness/geozone_sources/{id}",
        json={
            "https_source": {
                "url": TEST_DATASET_URL,
                "format": "ED-269",
            }
        },
        **client_options,
    )
    assert response.status_code == 200
    assert response.json["result"] == "Ready"
    assert "message" not in response.json.keys()

    # Status
    response = client.get(f"/geoawareness/geozone_sources/{id}", **client_options)
    assert response.status_code == 200
    assert response.json["result"] == "Ready"

    # Delete
    response = client.delete(f"/geoawareness/geozone_sources/{id}", **client_options)
    assert response.status_code == 200
    assert response.json["result"] == "Deactivating"

    # Deleted
    response = client.get(f"/geoawareness/geozone_sources/{id}", **client_options)
    assert response.status_code == 404


def test_geosource_unsupported_source(client, client_options):
    id = uuid.uuid4()

    # Creation with invalid payload
    response = client.put(
        f"/geoawareness/geozone_sources/{id}",
        json={"unsupported_source": {}},
        **client_options,
    )
    assert response.status_code == 400


def test_geosource_not_found(client, client_options):
    id = uuid.uuid4()

    # Source not present
    response = client.get(f"/geoawareness/geozone_sources/{id}", **client_options)
    assert response.status_code == 404


def test_geosource_unsupported_format(client, client_options):
    id = uuid.uuid4()

    # Creation with unsupported format
    response = client.put(
        f"/geoawareness/geozone_sources/{id}",
        json={
            "https_source": {
                "url": TEST_DATASET_URL,
                "format": "Unsupported",
            }
        },
        **client_options,
    )
    assert response.status_code == 400


def test_geosource_bad_url(client, client_options):
    id = uuid.uuid4()

    # Creation with invalid url
    response = client.put(
        f"/geoawareness/geozone_sources/{id}",
        json={
            "https_source": {
                "url": "/not_found",
                "format": "ED-269",
            }
        },
        **client_options,
    )
    assert response.status_code == 200
    assert response.json["result"] == "Error"
    assert response.json["message"].startswith(
        "Unable to download and parse /not_found"
    )


def test_geozone_simple_check(client, client_options):
    id = uuid.uuid4()

    # Data source creation
    response = client.put(
        f"/geoawareness/geozone_sources/{id}",
        json={
            "https_source": {
                "url": TEST_DATASET_URL,
                "format": "ED-269",
            }
        },
        **client_options,
    )
    assert response.status_code == 200
    assert response.json["result"] == "Ready"  # Synchronous load
    assert "message" not in response.json.keys()

    test_positions = {
        "montreux": {
            "uomDimensions": "M",
            "verticalReferenceType": "AGL",
            "height": 100,
            "longitude": 6.913146,
            "latitude": 46.430758,
        },
        "reichenbach": {
            "uomDimensions": "M",
            "verticalReferenceType": "AGL",
            "height": 100,
            "longitude": 7.67807,
            "latitude": 46.612893,
        },
        "lausanne": {
            "uomDimensions": "M",
            "verticalReferenceType": "AGL",
            "height": 100,
            "longitude": 6.616329788594015,
            "latitude": 46.54017511057239,
        },
        "gantrisch": {
            "uomDimensions": "M",
            "verticalReferenceType": "AGL",
            "height": 100,
            "longitude": 7.4190915566346405,
            "latitude": 46.78810934830412,
        },
        "geneve": {  # Outside of all geozones
            "uomDimensions": "M",
            "verticalReferenceType": "AGL",
            "height": 100,
            "longitude": 6.157840457872254,
            "latitude": 46.19334706610729,
        },
    }

    # Checks: Tuple(filterset, True if present)
    checks = [
        # Test permanent applicability of one of the geozone.
        (
            {"filterSets": [{"before": "2020-01-01T00:00:00Z"}]},
            True,
        ),
        # Test Reichenbach zone time applicability. (not before)
        (
            {
                "filterSets": [
                    {
                        "before": "2020-01-01T00:00:00Z",
                        "position": test_positions["reichenbach"],
                    }
                ]
            },
            False,
        ),
        # Test Reichenbach zone time applicability. (not after)
        (
            {
                "filterSets": [
                    {
                        "after": "2026-01-01T00:00:00Z",
                        "position": test_positions["reichenbach"],
                    }
                ],
            },
            False,
        ),
        # Test Reichenbach zone time applicability. (during)
        (
            {
                "filterSets": [
                    {
                        "after": "2022-06-20T00:00:00Z",
                        "position": test_positions["reichenbach"],
                    }
                ]
            },
            True,
        ),
        # Test Reichenbach zone restriction. (not REQ_AUTHORISATION)
        (
            {
                "filterSets": [
                    {
                        "after": "2022-06-20T00:00:00Z",
                        "position": test_positions["reichenbach"],
                        "ed269": {"acceptableRestrictions": ["REQ_AUTHORISATION"]},
                    }
                ]
            },
            False,
        ),
        # Test Reichenbach zone restrictions. (is PROHIBITED)
        (
            {
                "filterSets": [
                    {
                        "after": "2022-06-20T00:00:00Z",
                        "position": test_positions["reichenbach"],
                        "ed269": {
                            "acceptableRestrictions": [
                                "REQ_AUTHORISATION",
                                "PROHIBITED",
                            ]
                        },
                    }
                ]
            },
            True,
        ),
        # Test Montreux zone. (Circle)
        ({"filterSets": [{"position": test_positions["montreux"]}]}, True),
        # Test Gantrisch zone restrictions. (Polygon)
        ({"filterSets": [{"position": test_positions["gantrisch"]}]}, True),
        # Test Gantrisch zone uSpaceClass. (is SPECIFIC)
        (
            {
                "filterSets": [
                    {
                        "position": test_positions["gantrisch"],
                        "ed269": {"uSpaceClass": "SPECIFIC"},
                    }
                ]
            },
            True,
        ),
        # Test location without zone (Geneve).
        ({"filterSets": [{"position": test_positions["geneve"]}]}, False),
    ]

    response = client.post(
        f"/geoawareness/check",
        json={"checks": [c[0] for c in checks]},
        **client_options,
    )
    assert response.status_code == 200
    assert response.json["applicableGeozone"] == [
        "Present" if c[1] else "Absent" for c in checks
    ]
