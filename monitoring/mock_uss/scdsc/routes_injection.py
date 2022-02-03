from typing import Tuple
import uuid

import flask

from monitoring.monitorlib import scd
from monitoring.monitorlib.clients import scd as scd_client
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResponse, SCOPE_SCD_QUALIFIER_INJECT, InjectFlightResult
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.mock_uss import config, resources, webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.scdsc import database
from monitoring.mock_uss.scdsc.database import db


@webapp.route('/scdsc/v1/status', methods=['GET'])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_injection_status() -> Tuple[str, int]:
    """Implements USS status in SCD automated testing injection API."""
    return flask.jsonify({'status': 'Ready'})


@webapp.route('/scdsc/v1/flights/<flight_id>', methods=['PUT'])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def inject_flight(flight_id: str) -> Tuple[str, int]:
    """Implements flight injection in SCD automated testing injection API."""

    try:
        json = flask.request.json
        if json is None:
            raise ValueError('Request did not contain a JSON payload')
        req_body: InjectFlightRequest = ImplicitDict.parse(
            json, InjectFlightRequest)
    except ValueError as e:
        msg = 'Create flight {} unable to parse JSON: {}'.format(flight_id, e)
        return msg, 400

    # Validate flight authorisation
    # TODO: Implement

    # Check for operational intents in the DSS
    start_time = scd.start_of(req_body.operational_intent.volumes)
    end_time = scd.end_of(req_body.operational_intent.volumes)
    area = scd.rect_bounds_of(req_body.operational_intent.volumes)
    vol4 = scd.make_vol4(start_time, end_time, 0, 3048, polygon=scd.make_polygon(latlngrect=area))
    try:
        op_intents = scd_client.query_operational_intents(resources.utm_client, vol4, db.cached_operations)
    except (ValueError, scd_client.OperationError) as e:
        # TODO: Add text message to InjectFlightResponse to document reason for outcome
        print('Error querying operational intents: {}'.format(e))
        return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Failed)), 200

    # Check for intersections
    v1 = req_body.operational_intent.volumes
    for op_intent in op_intents:
        v2a = op_intent.details.volumes
        v2b = op_intent.details.off_nominal_volumes
        if scd.vol4s_intersect(v1, v2a) or scd.vol4s_intersect(v1, v2b):
            # TODO: Add text message to InjectFlightResponse to document reason for outcome
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.ConflictWithFlight)), 200

    # Create operational intent in DSS
    base_url = '{}/mock/scd'.format(webapp.config[config.KEY_BASE_URL])
    req = scd.PutOperationalIntentReferenceParameters(
        extents=req_body.operational_intent.volumes,
        key=[op.reference.ovn for op in op_intents],
        state=req_body.operational_intent.state,
        uss_base_url=base_url,
        new_subscription=scd.ImplicitSubscriptionParameters(
            uss_base_url=base_url
        )
    )
    id = str(uuid.uuid4())
    try:
        result = scd_client.create_operational_intent_reference(resources.utm_client, id, req)
    except (ValueError, scd_client.OperationError) as e:
        # TODO: Add text message to InjectFlightResponse to document reason for outcome
        print('Error creating operational intent: {}'.format(e))
        return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Failed)), 200
    # TODO: Notify subscribers

    # Store flight in database
    record = database.FlightRecord(
        op_intent_reference=result.operational_intent_reference,
        op_intent_injection=req_body.operational_intent,
        flight_authorisation=req_body.flight_authorisation)
    db.flights[flight_id] = record

    return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Planned, operational_intent_id=id))
