from typing import Tuple
import flask
from implicitdict import ImplicitDict

from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.geoawareness.geozone_sources import (
    get_geozone_source,
    create_geozone_source,
    delete_geozone_source,
)
from monitoring.monitorlib.geoawareness_automated_testing import api as geoawareness_api


@webapp.route(
    "/geoawareness/geozone_sources/<geozone_source_id>",
    methods=["GET", "PUT", "DELETE"],
)
@requires_scope([geoawareness_api.SCOPE_GEOAWARENESS_TEST])
def geozone_sources(geozone_source_id: str) -> Tuple[str, int]:
    def _request_body() -> geoawareness_api.GeozoneSourceDefinition:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        req_body: geoawareness_api.GeozoneSourceDefinition = ImplicitDict.parse(
            json, geoawareness_api.GeozoneSourceDefinition
        )
        return req_body

    if flask.request.method == "GET":
        return get_geozone_source(geozone_source_id)

    elif flask.request.method == "PUT":
        try:
            body = _request_body()
        except ValueError as e:
            msg = "Create geozone source {} unable to parse JSON: {}".format(
                geozone_source_id, e
            )
            return msg, 400

        return create_geozone_source(geozone_source_id, body)

    elif flask.request.method == "DELETE":
        return delete_geozone_source(geozone_source_id)

    else:
        return "Unsupported Method", 405
