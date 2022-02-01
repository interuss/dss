import flask

from monitoring.monitorlib import scd
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope


@webapp.route('/mock/scd/uss/v1/operational_intents/<entityid>', methods=['GET'])
@requires_scope([scd.SCOPE_SC])
def get_operational_intent_details(entityid: str):
    """Implements getOperationalIntentDetails in ASTM SCD API."""

    return flask.jsonify({'message': 'Not yet implemented'}), 500

    # Look up entityid in database
    # TODO: Implement

    # If requested operational intent doesn't exist, return 404
    # TODO: Implement

    # Return nominal response with details
    # TODO: Implement


@webapp.route('/mock/scd/uss/v1/operational_intents', methods=['POST'])
@requires_scope([scd.SCOPE_SC])
def notify_operational_intent_details_changed():
    """Implements notifyOperationalIntentDetailsChanged in ASTM SCD API."""

    # Do nothing because this USS is unsophisticated and polls the DSS for every
    # change in its operational intents
    return '', 204


@webapp.route('/mock/scd/uss/v1/reports', methods=['POST'])
@requires_scope([scd.SCOPE_SC, scd.SCOPE_CP, scd.SCOPE_CM, scd.SCOPE_CM_SA, scd.SCOPE_AA])
def make_uss_report():
    """Implements makeUssReport in ASTM SCD API."""

    return flask.jsonify({'message': 'Not yet implemented'}), 500

    # Parse the request
    # TODO: Implement

    # Construct the ErrorReport object, primarily from the request
    # TODO: Implement

    # Do not store the ErrorReport (in this diagnostic implementation)

    # Return the ErrorReport as the nominal response
    # TODO: Implement
