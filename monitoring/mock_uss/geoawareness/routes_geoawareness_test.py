import uuid

import pytest
from flask import json

from monitoring.mock_uss import webapp
from monitoring.mock_uss.config import KEY_TOKEN_AUDIENCE
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
    assert response.json["result"] == "Activating"

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
