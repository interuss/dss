from typing import List, Tuple
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


def query_operational_intents(area_of_interest: scd.Volume4D) -> List[scd.OperationalIntent]:
    """Retrieve a complete set of operational intents in an area, including details.

    :param area_of_interest: Area where intersecting operational intents must be discovered
    :return: Full definition for every operational intent discovered
    """
    op_intent_refs = scd_client.query_operational_intent_references(resources.utm_client, area_of_interest)
    tx = db.value
    get_details_for = []
    for op_intent_ref in op_intent_refs:
        if op_intent_ref.id not in tx.cached_operations or tx.cached_operations[op_intent_ref.id].reference.version != op_intent_ref.version:
            get_details_for.append(op_intent_ref)

    updated_op_intents = []
    for op_intent_ref in get_details_for:
        updated_op_intents.append(scd_client.get_operational_intent_details(resources.utm_client, op_intent_ref.uss_base_url, op_intent_ref.id))

    with db as tx:
        for op_intent in updated_op_intents:
            tx.cached_operations[op_intent.reference.id] = op_intent
        return [tx.cached_operations[op_intent_ref.id] for op_intent_ref in op_intent_refs]


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

    if webapp.config[config.KEY_BEHAVIOR_LOCALITY].is_uspace_applicable:
        # Validate flight authorisation
        pass  # TODO: Implement

    # Check for operational intents in the DSS
    start_time = scd.start_of(req_body.operational_intent.volumes)
    end_time = scd.end_of(req_body.operational_intent.volumes)
    area = scd.rect_bounds_of(req_body.operational_intent.volumes)
    alt_lo, alt_hi = scd.meter_altitude_bounds_of(req_body.operational_intent.volumes)
    vol4 = scd.make_vol4(start_time, end_time, alt_lo, alt_hi, polygon=scd.make_polygon(latlngrect=area))
    try:
        op_intents = query_operational_intents(vol4)
    except (ValueError, scd_client.OperationError) as e:
        notes = 'Error querying operational intents: {}'.format(e)
        return flask.jsonify(InjectFlightResponse(
            result=InjectFlightResult.Failed, notes=notes)), 200

    # Check for intersections
    v1 = req_body.operational_intent.volumes
    for op_intent in op_intents:
        if req_body.operational_intent.priority > op_intent.details.priority:
            continue
        if webapp.config[config.KEY_BEHAVIOR_LOCALITY].allow_same_priority_intersections:
            continue
        v2a = op_intent.details.volumes
        v2b = op_intent.details.off_nominal_volumes
        if scd.vol4s_intersect(v1, v2a) or scd.vol4s_intersect(v1, v2b):
            notes = 'Requested flight intersected {}\'s operational intent {}'.format(op_intent.reference.manager, op_intent.reference.id)
            return flask.jsonify(InjectFlightResponse(
                result=InjectFlightResult.ConflictWithFlight, notes=notes)), 200

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
        notes = 'Error creating operational intent: {}'.format(e)
        return flask.jsonify(InjectFlightResponse(
            result=InjectFlightResult.Failed, notes=notes)), 200
    scd_client.notify_subscribers(
        resources.utm_client, result.operational_intent_reference.id,
        scd.OperationalIntent(
            reference=result.operational_intent_reference,
            details=req_body.operational_intent),
        result.subscribers)

    # Store flight in database
    record = database.FlightRecord(
        op_intent_reference=result.operational_intent_reference,
        op_intent_injection=req_body.operational_intent,
        flight_authorisation=req_body.flight_authorisation)
    with db as tx:
        tx.flights[flight_id] = record

    return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Planned, operational_intent_id=id))
