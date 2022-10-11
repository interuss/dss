from datetime import datetime
from typing import List, Tuple
from urllib.parse import urlparse
import uuid

import flask
import requests.exceptions
import yaml

from monitoring.monitorlib import scd, versioning
from monitoring.monitorlib.clients import scd as scd_client
from monitoring.monitorlib.scd_automated_testing import scd_injection_api
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResponse, SCOPE_SCD_QUALIFIER_INJECT, InjectFlightResult, DeleteFlightResponse, DeleteFlightResult, ClearAreaRequest, ClearAreaOutcome, ClearAreaResponse, Capability, CapabilitiesResponse
from implicitdict import ImplicitDict, StringBasedDateTime
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
    return flask.jsonify({'status': 'Ready', 'version': versioning.get_code_version()})


@webapp.route('/scdsc/v1/capabilities', methods=['GET'])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scd_capabilities() -> Tuple[str, int]:
    """Implements USS capabilities in SCD automated testing injection API."""
    return flask.jsonify(CapabilitiesResponse(capabilities=[
        Capability.BasicStrategicConflictDetection,
        Capability.FlightAuthorisationValidation,
        Capability.HighPriorityFlights]))


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
        flight_auth = req_body.flight_authorisation
        if not flight_auth.uas_serial_number.valid:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid serial number'))
        if not flight_auth.operator_id.valid:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid operator ID'))
        if flight_auth.uas_class == scd_injection_api.UASClass.Other:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid UAS class'))
        if flight_auth.operation_category == scd_injection_api.OperationCategory.Unknown:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid operation category'))
        if flight_auth.endurance_minutes < 1 or flight_auth.endurance_minutes > 10 * 24 * 60:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid endurance'))
        if sum(1 if len(m) > 0 else 0 for m in flight_auth.connectivity_methods) == 0:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid connectivity methods'))
        if sum(1 if len(m) > 0 else 0 for m in flight_auth.identification_technologies) == 0:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid identification technologies'))
        try:
            urlparse(flight_auth.emergency_procedure_url)
        except ValueError:
            return flask.jsonify(InjectFlightResponse(result=InjectFlightResult.Rejected, notes='Invalid emergency procedure URL'))

    # Check for operational intents in the DSS
    start_time = scd.start_of(req_body.operational_intent.volumes)
    end_time = scd.end_of(req_body.operational_intent.volumes)
    area = scd.rect_bounds_of(req_body.operational_intent.volumes)
    alt_lo, alt_hi = scd.meter_altitude_bounds_of(req_body.operational_intent.volumes)
    vol4 = scd.make_vol4(start_time, end_time, alt_lo, alt_hi, polygon=scd.make_polygon(latlngrect=area))
    try:
        op_intents = query_operational_intents(vol4)
    except (ValueError, scd_client.OperationError, requests.exceptions.ConnectionError, ConnectionError) as e:
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
    except (ValueError, scd_client.OperationError, requests.exceptions.ConnectionError, ConnectionError) as e:
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


@webapp.route('/scdsc/v1/flights/<flight_id>', methods=['DELETE'])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def delete_flight(flight_id: str) -> Tuple[str, int]:
    """Implements flight deletion in SCD automated testing injection API."""

    with db as tx:
        flight = tx.flights.pop(flight_id, None)

    if flight is None:
        return flask.jsonify(DeleteFlightResponse(
            result=DeleteFlightResult.Failed,
            notes='Flight {} does not exist'.format(flight_id))), 200

    # Delete operational intent from DSS
    try:
        result = scd_client.delete_operational_intent_reference(
            resources.utm_client,
            flight.op_intent_reference.id,
            flight.op_intent_reference.ovn)
    except (ValueError, scd_client.OperationError, requests.exceptions.ConnectionError, ConnectionError) as e:
        notes = 'Error deleting operational intent: {}'.format(e)
        return flask.jsonify(DeleteFlightResponse(
            result=DeleteFlightResult.Failed, notes=notes)), 200
    scd_client.notify_subscribers(
        resources.utm_client, result.operational_intent_reference.id,
        None, result.subscribers)

    return flask.jsonify(DeleteFlightResponse(result=DeleteFlightResult.Closed))


@webapp.route('/scdsc/v1/clear_area_requests', methods=['POST'])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def clear_area() -> Tuple[str, int]:
    try:
        json = flask.request.json
        if json is None:
            raise ValueError('Request did not contain a JSON payload')
        req = ImplicitDict.parse(json, ClearAreaRequest)
    except ValueError as e:
        msg = 'Unable to parse ClearAreaRequest JSON request: {}'.format(e)
        return msg, 400

    # Find operational intents in the DSS
    start_time = scd.start_of([req.extent])
    end_time = scd.end_of([req.extent])
    area = scd.rect_bounds_of([req.extent])
    alt_lo, alt_hi = scd.meter_altitude_bounds_of([req.extent])
    vol4 = scd.make_vol4(start_time, end_time, alt_lo, alt_hi, polygon=scd.make_polygon(latlngrect=area))
    try:
        op_intent_refs = scd_client.query_operational_intent_references(resources.utm_client, vol4)
    except (ValueError, scd_client.OperationError, requests.exceptions.ConnectionError, ConnectionError) as e:
        msg = 'Error querying operational intents: {}'.format(e)
        return flask.jsonify(ClearAreaResponse(
            outcome=ClearAreaOutcome(
                success=False, message=msg,
                timestamp=StringBasedDateTime(datetime.utcnow())),
            request=req)), 200

    # Try to delete every operational intent found
    dss_deletion_results = {}
    deleted = set()
    for op_intent_ref in op_intent_refs:
        try:
            scd_client.delete_operational_intent_reference(resources.utm_client, op_intent_ref.id, op_intent_ref.ovn)
            dss_deletion_results[op_intent_ref.id] = 'Deleted from DSS'
            deleted.add(op_intent_ref.id)
        except scd_client.OperationError as e:
            dss_deletion_results[op_intent_ref.id] = str(e)

    # Delete corresponding flight injections and cached operational intents
    with db as tx:
        flights_to_delete = []
        for flight_id, record in tx.flights.items():
            if record.op_intent_reference.id in deleted:
                flights_to_delete.append(flight_id)
        for flight_id in flights_to_delete:
            del tx.flights[flight_id]

        cache_deletions = []
        for op_intent_id in deleted:
            if op_intent_id in tx.cached_operations:
                del tx.cached_operations[op_intent_id]
                cache_deletions.append(op_intent_id)

    msg = yaml.dump({
        'dss_deletions': dss_deletion_results,
        'flight_deletions': flights_to_delete,
        'cache_deletions': cache_deletions,
    })
    return flask.jsonify(ClearAreaResponse(
        outcome=ClearAreaOutcome(
            success=True, message=msg,
            timestamp=StringBasedDateTime(datetime.utcnow())),
        request=req)), 200
