import flask

from monitoring.monitorlib import scd
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.scdsc.database import db
from loguru import logger
import os
import json
import monitoring.mock_uss.scdsc.request_validator as request_validator
from monitoring.mock_uss.scdsc import report_settings
import monitoring.messagesigning.message_signer as signer

import traceback


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
    response = None
    status_code = None
    # If requested operational intent doesn't exist, return 404
    if flight is None:
        response = scd.ErrorResponse(
            message='Operational intent {} not known by this USS'.format(
                entityid
            )
        )
        status_code = 404
    else:
        # Return nominal response with details
        response = scd.GetOperationalIntentDetailsResponse(
            operational_intent=scd.OperationalIntent(
                reference=flight.op_intent_reference,
                details=scd.OperationalIntentDetails(
                    volumes=flight.op_intent_injection.volumes,
                    off_nominal_volumes=flight.op_intent_injection.off_nominal_volumes,
                    priority=flight.op_intent_injection.priority
                )
            )
        )
        status_code = 200

    # Check message signing headers only if the message signing feature is on.
    if os.environ.get('MESSAGE_SIGNING', None) == "true":
        try:
            analysis_result = request_validator.validate_message_signing_headers()
            test_context = analysis_result['test_context']
            query = analysis_result['query']
            results = analysis_result['results']
            if results['validation_passed']:
                query['response'] = {
                    'code': 200,
                    'json': response
                }
                status_code = 200
            else:
                failure_reasons = results['validation_issue']
                error_message = "{}: {}".format(failure_reasons['summary'], failure_reasons['details'])
                response = scd.ErrorResponse(message=error_message)
                query['response'] = {
                    'code': 403,
                    'json': response
                }
                status_code = 403
            resp = sign_response(flask.jsonify(response))
            query['response']['headers'] = json.dumps({k: v for k, v in resp.headers.items()})
            interaction_id = report_settings.reprt_recorder.capture_interaction(
             query,
        'Checking that the message signing headers in the incoming {} request to the mock uss endpoint {} are valid.'.format(flask.request.method, flask.request.path),
        test_context=test_context)
            if not results['validation_passed']:
                results['validation_issue']['interactions'] = [interaction_id]
                report_settings.reprt_recorder.capture_issue(results['validation_issue'])
        except Exception as e:
            logger.error("Exception raised while validating message signing headers: " + str(e))
            logger.error(traceback.format_exc())

    return response, status_code


@webapp.route("/mock/scd/uss/v1/operational_intents", methods=["POST"])
@requires_scope([scd.SCOPE_SC])
def notify_operational_intent_details_changed():
    """Implements notifyOperationalIntentDetailsChanged in ASTM SCD API."""
    # Check message signing headers only if the message signing feature is on.
    response = flask.jsonify('')
    status_code = 204
    if os.environ.get('MESSAGE_SIGNING', None) == "true":
        try:
            analysis_result = request_validator.validate_message_signing_headers()
            test_context = analysis_result['test_context']
            query = analysis_result['query']
            results = analysis_result['results']
            if results['validation_passed']:
                query['response'] = {
                    'code': 204,
                    'json': None
                }
                status_code = 204
            else:
                failure_reasons = results['validation_issue']
                error_message = "{}: {}".format(failure_reasons['summary'], failure_reasons['details'])
                query['response'] = {
                    'code': 403,
                    'json': {'message': error_message}
                }
                response = flask.jsonify(scd.ErrorResponse(message=error_message))
                status_code = 403
            resp = sign_response(response)
            query['response']['headers'] = json.dumps({k: v for k, v in resp.headers.items()})
            interaction_id = report_settings.reprt_recorder.capture_interaction(
                query,
            'Checking that the message signing headers in the incoming {} request to the mock uss endpoint {} are valid.'.format(flask.request.method, flask.request.path),
            test_context=test_context)
            if not results['validation_passed']:
                results['validation_issue']['interactions'] = [interaction_id]
                report_settings.reprt_recorder.capture_issue(results['validation_issue'])
        except Exception as e:
            logger.error("Exception raised while validating message signing headers: " + str(e))
            logger.error(traceback.format_exc())
    return response, status_code


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

def sign_response(response):
    try:
        type_of_response = str(type(response))
        if 'None' not in type_of_response and os.environ.get('MESSAGE_SIGNING', None) == "true":
            signed_headers = signer.get_signed_headers(response)
            response.headers.update(signed_headers)
    except Exception as e:
        logger.error("Could not sign response: " + str(e))
    return response
