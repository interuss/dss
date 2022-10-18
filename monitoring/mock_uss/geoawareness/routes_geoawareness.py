from typing import Tuple
import flask
from implicitdict import ImplicitDict

from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.geoawareness.check import check_geozones
from monitoring.mock_uss.geoawareness.geozone_sources import (
    get_geozone_source,
    create_geozone_source,
    delete_geozone_source,
)
from monitoring.monitorlib.geoawareness_automated_testing import api as geoawareness_api
from monitoring.monitorlib.geoawareness_automated_testing.api import (
    GeozonesCheckResponse,
)


@webapp.route(
    "/geoawareness/geozone_sources/<geozone_source_id>",
    methods=["GET"],
)
@requires_scope([geoawareness_api.SCOPE_GEOAWARENESS_TEST])
def get_geozone_sources(geozone_source_id: str) -> Tuple[str, int]:
    return get_geozone_source(geozone_source_id)


@webapp.route(
    "/geoawareness/geozone_sources/<geozone_source_id>",
    methods=["PUT"],
)
@requires_scope([geoawareness_api.SCOPE_GEOAWARENESS_TEST])
def put_geozone_sources(geozone_source_id: str) -> Tuple[str, int]:
    try:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        body: geoawareness_api.GeozoneSourceDefinition = ImplicitDict.parse(
            json, geoawareness_api.GeozoneSourceDefinition
        )
    except ValueError as e:
        msg = "Create geozone source {} unable to parse JSON: {}".format(
            geozone_source_id, e
        )
        return msg, 400

    return create_geozone_source(geozone_source_id, body)


@webapp.route(
    "/geoawareness/geozone_sources/<geozone_source_id>",
    methods=["DELETE"],
)
@requires_scope([geoawareness_api.SCOPE_GEOAWARENESS_TEST])
def delete_geozone_sources(geozone_source_id: str) -> Tuple[str, int]:
    return delete_geozone_source(geozone_source_id)


@webapp.route("/geoawareness/check", methods=["POST"])
@requires_scope([geoawareness_api.SCOPE_GEOAWARENESS_TEST])
def check():
    try:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        body: geoawareness_api.GeozonesCheckRequest = ImplicitDict.parse(
            json, geoawareness_api.GeozonesCheckRequest
        )
    except ValueError as e:
        msg = "Geozone check unable to parse JSON: {}".format(e)
        return msg, 400
    applicable_geozone = check_geozones(body)

    return GeozonesCheckResponse(applicableGeozone=applicable_geozone), 200
