from typing import Tuple

import flask

from implicitdict import ImplicitDict
from monitoring.mock_uss import webapp
from .behavior import ServiceProviderBehavior
from .database import db


@webapp.route("/ridsp/behavior", methods=["PUT"])
def ridsp_set_dp_behavior() -> Tuple[str, int]:
    """Set the behavior of the mock Display Provider."""
    try:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        dp_behavior = ImplicitDict.parse(json, ServiceProviderBehavior)
    except ValueError as e:
        msg = "Change behavior for Service Provider unable to parse JSON: {}".format(e)
        return msg, 400

    with db as tx:
        tx.behavior = dp_behavior

    return flask.jsonify(dp_behavior)


@webapp.route("/ridsp/behavior", methods=["GET"])
def ridsp_get_dp_behavior() -> Tuple[str, int]:
    """Get the behavior of the mock Display Provider."""
    return flask.jsonify(db.value.behavior)
