from typing import Tuple

import flask

from monitoring.monitorlib.scd_automated_testing import scd_injection_api
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope


@webapp.route('/scdsc/v1/status', methods=['GET'])
@requires_scope([scd_injection_api.SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_injection_status() -> Tuple[str, int]:
    """Implements USS status in SCD automated testing injection API."""
    return flask.jsonify({'status': 'Ready'})


@webapp.route('/scdsc/v1/flights/<flight_id>', methods=['PUT'])
@requires_scope([scd_injection_api.SCOPE_SCD_QUALIFIER_INJECT])
def inject_flight(flight_id: str) -> Tuple[str, int]:
    """Implements flight injection in SCD automated testing injection API."""

    try:
        json = flask.request.json
        if json is None:
            raise ValueError('Request did not contain a JSON payload')
        req_body: scd_injection_api.InjectFlightRequest = ImplicitDict.parse(
            json, scd_injection_api.InjectFlightRequest)
    except ValueError as e:
        msg = 'Create flight {} unable to parse JSON: {}'.format(flight_id, e)
        return msg, 400

    return flask.jsonify(scd_injection_api.InjectFlightResponse(
        result=scd_injection_api.InjectFlightResult.Failed))

    # Validate flight authorisation
    # TODO: Implement

    # Check for operational intents in the DSS
    # TODO: Implement

    # Check for intersections
    # TODO: Implement

    # Create operational intent in DSS
    # TODO: Implement

    # Store flight in database
    # record = database.FlightRecord(
    #     op_intent_id=op_intent_id,
    #     op_intent=req_body.operational_intent,
    #     flight_authorisation=req_body.flight_authorisation)
    # db.tests[flight_id] = record
    #
    # return flask.jsonify(scd_injection_api.InjectFlightResponse(result=, operational_intent_id=op_intent_id))
