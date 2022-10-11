import flask

from monitoring.monitorlib import scd
from monitoring.mock_uss import webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.scdsc.database import db
from monitoring.mock_uss.scdsc import report_settings
import json
import datetime

@webapp.route("/mock/scd/uss/v1/operational_intents/<entityid>", methods=["GET"])
@requires_scope([scd.SCOPE_SC])
def get_operational_intent_details(entityid: str):
    """Implements getOperationalIntentDetails in ASTM SCD API."""

    req_info = {
        'method': flask.request.method,
        'url': flask.request.url,
        'initiated_at': datetime.datetime.utcnow().isoformat(),
        'headers': json.dumps({k: v for k, v in flask.request.headers.items()})
    }

    req_info['body'] = flask.request.data.decode('utf-8')
    query = {'request': req_info}

    test_context = {
        'test_name': 'Message signing in GET Op req from UUT',
        'test_case': 'Uss posts Op with Message Signing Expect 200'}
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
        query['reponse'] = {'status_code': 404, 'body': json.dumps(err_resp)}
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
    query['response'] = {'status_code': 200, 'body': json.dumps(response)}
    report_settings.reprt_recorder.capture_interaction(query,
                                                        "Checking message signing for GET /mock/scd/uss/v1/operational_intents/<entityid>",
                                                        test_context=test_context)
    return flask.jsonify(response), 200


@webapp.route("/mock/scd/uss/v1/operational_intents", methods=["POST"])
@requires_scope([scd.SCOPE_SC])
def notify_operational_intent_details_changed():
    """Implements notifyOperationalIntentDetailsChanged in ASTM SCD API."""
    req_info = {
        'method': flask.request.method,
        'url': flask.request.url,
        'initiated_at': datetime.datetime.utcnow().isoformat(),
        'headers': json.dumps({k: v for k, v in flask.request.headers.items()})
    }

    req_info['body'] = flask.request.data.decode('utf-8')
    query = {'request': req_info}

    test_context = {
        'test_name': 'Message signing in POST Op from UUT',
        'test_case': 'Uss posts Op with Message Signing Expect 204'}

    ms_headers = [
        "x-utm-message-signature",
        "x-utm-message-signature-input",
        "x-utm-jws-header ",
        "content-digest"
    ]

    missing_ms_header = []
    for ms_header in ms_headers:
        if not flask.request.headers.has_key(ms_header):
            missing_ms_header.append(ms_header)

    if len(missing_ms_header) > 0:
        resp_invalid_ms = {"error_type": "Missing Message Signature header", "reason found": "Missing header"}
        query['response'] = {'status_code': 403, 'body': json.dumps(resp_invalid_ms)}

        interact_id = report_settings.reprt_recorder.capture_interaction(
            query, "Checking message signing for POST /mock/scd/uss/v1/operational_intents/<entityid>",
            test_context=test_context)
        issue = {
            'context': test_context,
            'uss_role': "Notified USS",
            'target': "USS",
            'summary': "POST Request to mock_uss was missing valid message signing",
            'details': "POST request sent back 403. " + json.dumps(resp_invalid_ms),
            'interactions': [interact_id]
        }
        report_settings.reprt_recorder.capture_issue(issue)
        return flask.jsonify(resp_invalid_ms), 403

    else:
        query['response'] = {'status_code': 204, 'body': ""}
        report_settings.reprt_recorder.capture_interaction(
            query,
            "Checking message signing for POST /mock/scd/uss/v1/operational_intents/<entityid>",
            test_context=test_context)
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
