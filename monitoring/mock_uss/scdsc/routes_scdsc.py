import flask

from monitoring.monitorlib import scd
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.scdsc.database import db
from loguru import logger


@webapp.route("/mock/scd/uss/v1/operational_intents/<entityid>", methods=["GET"])
@requires_scope([scd.SCOPE_SC])
def get_operational_intent_details(entityid: str):
    """Implements getOperationalIntentDetails in ASTM SCD API."""
    # Look up entityid in database
    tx = db.value
    flight = None
    for f in tx.flights.values():
        if f.op_intent_reference.id == entityid:
            flight = f
            break

    # If requested operational intent doesn't exist, return 404
    if flight is None:
        err_resp = scd.ErrorResponse(
            message='Operational intent {} not known by this USS'.format(
                entityid
            )
        )
        return flask.jsonify(err_resp), 404

    # Return nominal response with details
    response = scd.GetOperationalIntentDetailsResponse(
        operational_intent=scd.OperationalIntent(
            reference=flight.op_intent_reference,
            details=scd.OperationalIntentDetails(
                volumes=flight.op_intent_injection.volumes,
                off_nominal_volumes=flight.op_intent_injection.off_nominal_volumes,
                priority=flight.op_intent_injection.priority,
            ),
        )
    )
    return flask.jsonify(response), 200


@webapp.route("/mock/scd/uss/v1/operational_intents", methods=["POST"])
@requires_scope([scd.SCOPE_SC])
def notify_operational_intent_details_changed():
    """Implements notifyOperationalIntentDetailsChanged in ASTM SCD API."""
    return '', 204


@webapp.route("/mock/scd/uss/v1/reports", methods=["POST"])
@requires_scope(
    [scd.SCOPE_SC, scd.SCOPE_CP, scd.SCOPE_CM, scd.SCOPE_CM_SA, scd.SCOPE_AA]
)
def make_uss_report():
    """Implements makeUssReport in ASTM SCD API."""

    return flask.jsonify({"message": "Not yet implemented"}), 500

    # Parse the request
    # TODO: Implement

    # Construct the ErrorReport object, primarily from the request
    # TODO: Implement

    # Do not store the ErrorReport (in this diagnostic implementation)

    # Return the ErrorReport as the nominal response
    # TODO: Implement
