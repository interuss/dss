from datetime import datetime, timedelta
import time
from typing import List, Tuple
import uuid

import flask
import requests.exceptions
import yaml

from monitoring.monitorlib import scd, versioning
from monitoring.monitorlib.clients import scd as scd_client
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    InjectFlightResponse,
    SCOPE_SCD_QUALIFIER_INJECT,
    InjectFlightResult,
    DeleteFlightResponse,
    DeleteFlightResult,
    ClearAreaRequest,
    ClearAreaOutcome,
    ClearAreaResponse,
    Capability,
    CapabilitiesResponse,
)
from implicitdict import ImplicitDict, StringBasedDateTime
from monitoring.mock_uss import config, resources, webapp
from monitoring.mock_uss.auth import requires_scope
from monitoring.mock_uss.scdsc import database
from monitoring.mock_uss.scdsc.database import db
from monitoring.monitorlib.uspace import problems_with_flight_authorisation


DEADLOCK_TIMEOUT = timedelta(seconds=60)


def query_operational_intents(
    area_of_interest: scd.Volume4D,
) -> List[scd.OperationalIntent]:
    """Retrieve a complete set of operational intents in an area, including details.

    :param area_of_interest: Area where intersecting operational intents must be discovered
    :return: Full definition for every operational intent discovered
    """
    op_intent_refs = scd_client.query_operational_intent_references(
        resources.utm_client, area_of_interest
    )
    tx = db.value
    get_details_for = []
    for op_intent_ref in op_intent_refs:
        if (
            op_intent_ref.id not in tx.cached_operations
            or tx.cached_operations[op_intent_ref.id].reference.version
            != op_intent_ref.version
        ):
            get_details_for.append(op_intent_ref)

    updated_op_intents = []
    for op_intent_ref in get_details_for:
        updated_op_intents.append(
            scd_client.get_operational_intent_details(
                resources.utm_client, op_intent_ref.uss_base_url, op_intent_ref.id
            )
        )

    with db as tx:
        for op_intent in updated_op_intents:
            tx.cached_operations[op_intent.reference.id] = op_intent
        return [
            tx.cached_operations[op_intent_ref.id] for op_intent_ref in op_intent_refs
        ]


@webapp.route("/scdsc/v1/status", methods=["GET"])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_injection_status() -> Tuple[str, int]:
    """Implements USS status in SCD automated testing injection API."""
    json, code = injection_status()
    return flask.jsonify(json), code


def injection_status() -> Tuple[dict, int]:
    return (
        {"status": "Ready", "version": versioning.get_code_version()},
        200,
    )


@webapp.route("/scdsc/v1/capabilities", methods=["GET"])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_scd_capabilities() -> Tuple[str, int]:
    """Implements USS capabilities in SCD automated testing injection API."""
    json, code = scd_capabilities()
    return flask.jsonify(json), code


def scd_capabilities() -> Tuple[dict, int]:
    return (
        CapabilitiesResponse(
            capabilities=[
                Capability.BasicStrategicConflictDetection,
                Capability.FlightAuthorisationValidation,
                Capability.HighPriorityFlights,
            ]
        ),
        200,
    )


@webapp.route("/scdsc/v1/flights/<flight_id>", methods=["PUT"])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_inject_flight(flight_id: str) -> Tuple[str, int]:
    """Implements flight injection in SCD automated testing injection API."""
    print(f"[inject_flight:{flight_id}] Starting handler")
    try:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        req_body: InjectFlightRequest = ImplicitDict.parse(json, InjectFlightRequest)
    except ValueError as e:
        msg = "Create flight {} unable to parse JSON: {}".format(flight_id, e)
        return msg, 400
    json, code = inject_flight(flight_id, req_body)
    return flask.jsonify(json), code


def inject_flight(flight_id: str, req_body: InjectFlightRequest) -> Tuple[dict, int]:
    locality = webapp.config[config.KEY_BEHAVIOR_LOCALITY]

    if locality.is_uspace_applicable():
        # Validate flight authorisation
        print(f"[inject_flight:{flight_id}] Validating flight authorisation")
        problems = problems_with_flight_authorisation(req_body.flight_authorisation)
        if problems:
            return (
                InjectFlightResponse(
                    result=InjectFlightResult.Rejected, notes=", ".join(problems)
                ),
                200,
            )

    # Check if this is an existing flight being modified
    deadline = datetime.utcnow() + DEADLOCK_TIMEOUT
    while True:
        with db as tx:
            if flight_id in tx.flights:
                # This is an existing flight being modified
                existing_flight = tx.flights[flight_id]
                if existing_flight and not existing_flight.locked:
                    print(
                        f"[inject_flight:{flight_id}] Existing flight locked for update"
                    )
                    existing_flight.locked = True
                    break
            else:
                print(f"[inject_flight:{flight_id}] Request is for a new flight")
                tx.flights[flight_id] = None
                existing_flight = None
                break
        # We found an existing flight but it was locked; wait for it to become
        # available
        time.sleep(0.5)

        if datetime.utcnow() > deadline:
            raise RuntimeError(
                f"Deadlock in inject_flight while attempting to gain access to flight {flight_id}"
            )

    try:
        # Check for operational intents in the DSS
        print(
            f"[inject_flight:{flight_id}] Checking for operational intents in the DSS"
        )
        start_time = scd.start_of(req_body.operational_intent.volumes)
        end_time = scd.end_of(req_body.operational_intent.volumes)
        area = scd.rect_bounds_of(req_body.operational_intent.volumes)
        alt_lo, alt_hi = scd.meter_altitude_bounds_of(
            req_body.operational_intent.volumes
        )
        vol4 = scd.make_vol4(
            start_time,
            end_time,
            alt_lo,
            alt_hi,
            polygon=scd.make_polygon(latlngrect=area),
        )
        try:
            op_intents = query_operational_intents(vol4)
        except (
            ValueError,
            scd_client.OperationError,
            requests.exceptions.ConnectionError,
            ConnectionError,
        ) as e:
            notes = "Error querying operational intents: {}".format(e)
            return (
                InjectFlightResponse(result=InjectFlightResult.Failed, notes=notes),
                200,
            )

        # Check for intersections
        print(
            f"[inject_flight:{flight_id}] Checking for intersections with {', '.join(op_intent.reference.id for op_intent in op_intents)}"
        )
        v1 = req_body.operational_intent.volumes
        for op_intent in op_intents:
            if (
                existing_flight
                and existing_flight.op_intent_reference.id == op_intent.reference.id
            ):
                # Don't consider intersections with a past version of this flight
                continue
            if req_body.operational_intent.priority > op_intent.details.priority:
                # Don't consider intersections with lower-priority operational intents
                continue
            if (
                req_body.operational_intent.priority == op_intent.details.priority
                and locality.allows_same_priority_intersections(
                    req_body.operational_intent.priority
                )
            ):
                # Don't consider intersections with same-priority operational intents if they're allowed
                continue
            v2a = op_intent.details.volumes
            v2b = op_intent.details.off_nominal_volumes
            if scd.vol4s_intersect(v1, v2a) or scd.vol4s_intersect(v1, v2b):
                notes = f"Requested flight (priority {req_body.operational_intent.priority}) intersected {op_intent.reference.manager}'s operational intent {op_intent.reference.id} (priority {op_intent.details.priority})"
                return (
                    InjectFlightResponse(
                        result=InjectFlightResult.ConflictWithFlight, notes=notes
                    ),
                    200,
                )

        # Create operational intent in DSS
        print(f"[inject_flight:{flight_id}] Sharing operational intent with DSS")
        base_url = "{}/mock/scd".format(webapp.config[config.KEY_BASE_URL])
        req = scd.PutOperationalIntentReferenceParameters(
            extents=req_body.operational_intent.volumes,
            key=[op.reference.ovn for op in op_intents],
            state=req_body.operational_intent.state,
            uss_base_url=base_url,
            new_subscription=scd.ImplicitSubscriptionParameters(uss_base_url=base_url),
        )
        try:
            if existing_flight:
                id = existing_flight.op_intent_reference.id
                result = scd_client.update_operational_intent_reference(
                    resources.utm_client,
                    id,
                    existing_flight.op_intent_reference.ovn,
                    req,
                )
            else:
                id = str(uuid.uuid4())
                result = scd_client.create_operational_intent_reference(
                    resources.utm_client, id, req
                )
        except (
            ValueError,
            scd_client.OperationError,
            requests.exceptions.ConnectionError,
            ConnectionError,
        ) as e:
            notes = "Error creating operational intent: {}".format(e)
            print(f"[inject_flight:{flight_id}] {notes}")
            return (
                InjectFlightResponse(result=InjectFlightResult.Failed, notes=notes),
                200,
            )
        print(
            f"[inject_flight:{flight_id}] Notifying subscribers {', '.join(s.uss_base_url for s in result.subscribers)}"
        )
        scd_client.notify_subscribers(
            resources.utm_client,
            result.operational_intent_reference.id,
            scd.OperationalIntent(
                reference=result.operational_intent_reference,
                details=req_body.operational_intent,
            ),
            result.subscribers,
        )

        # Store flight in database
        print(f"[inject_flight:{flight_id}] Storing flight in database")
        record = database.FlightRecord(
            op_intent_reference=result.operational_intent_reference,
            op_intent_injection=req_body.operational_intent,
            flight_authorisation=req_body.flight_authorisation,
        )
        with db as tx:
            tx.flights[flight_id] = record

        print(f"[inject_flight:{flight_id}] Complete.")
        return (
            InjectFlightResponse(
                result=InjectFlightResult.Planned, operational_intent_id=id
            ),
            200,
        )
    finally:
        with db as tx:
            if tx.flights[flight_id]:
                # FlightRecord was a true existing flight
                print(
                    f"[inject_flight] Releasing placeholder for flight_id {flight_id}"
                )
                tx.flights[flight_id].locked = False
            else:
                # FlightRecord was just a placeholder for a new flight
                print(
                    f"[inject_flight] Releasing lock on existing flight_id {flight_id}"
                )
                del tx.flights[flight_id]


@webapp.route("/scdsc/v1/flights/<flight_id>", methods=["DELETE"])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_delete_flight(flight_id: str) -> Tuple[str, int]:
    """Implements flight deletion in SCD automated testing injection API."""
    json, code = delete_flight(flight_id)
    return flask.jsonify(json), code


def delete_flight(flight_id) -> Tuple[dict, int]:
    deadline = datetime.utcnow() + DEADLOCK_TIMEOUT
    while True:
        with db as tx:
            if flight_id in tx.flights:
                flight = tx.flights[flight_id]
                if flight and not flight.locked:
                    # FlightRecord was a true existing flight not being mutated anywhere else
                    del tx.flights[flight_id]
                    break
            else:
                # No FlightRecord found
                flight = None
                break
        # There is a race condition with another handler to create or modify the requested flight; wait for that to resolve
        time.sleep(0.5)

        if datetime.utcnow() > deadline:
            raise RuntimeError(
                f"Deadlock in delete_flight while attempting to gain access to flight {flight_id}"
            )

    if flight is None:
        return (
            DeleteFlightResponse(
                result=DeleteFlightResult.Failed,
                notes="Flight {} does not exist".format(flight_id),
            ),
            200,
        )

    # Delete operational intent from DSS
    try:
        result = scd_client.delete_operational_intent_reference(
            resources.utm_client,
            flight.op_intent_reference.id,
            flight.op_intent_reference.ovn,
        )
    except (
        ValueError,
        scd_client.OperationError,
        requests.exceptions.ConnectionError,
        ConnectionError,
    ) as e:
        notes = "Error deleting operational intent: {}".format(e)
        return (
            DeleteFlightResponse(result=DeleteFlightResult.Failed, notes=notes),
            200,
        )
    scd_client.notify_subscribers(
        resources.utm_client,
        result.operational_intent_reference.id,
        None,
        result.subscribers,
    )

    return DeleteFlightResponse(result=DeleteFlightResult.Closed), 200


@webapp.route("/scdsc/v1/clear_area_requests", methods=["POST"])
@requires_scope([SCOPE_SCD_QUALIFIER_INJECT])
def scdsc_clear_area() -> Tuple[str, int]:
    try:
        json = flask.request.json
        if json is None:
            raise ValueError("Request did not contain a JSON payload")
        req = ImplicitDict.parse(json, ClearAreaRequest)
    except ValueError as e:
        msg = "Unable to parse ClearAreaRequest JSON request: {}".format(e)
        return msg, 400
    json, code = clear_area(req)
    return flask.jsonify(json), code


def clear_area(req: ClearAreaRequest) -> Tuple[dict, int]:
    # Find operational intents in the DSS
    start_time = scd.start_of([req.extent])
    end_time = scd.end_of([req.extent])
    area = scd.rect_bounds_of([req.extent])
    alt_lo, alt_hi = scd.meter_altitude_bounds_of([req.extent])
    vol4 = scd.make_vol4(
        start_time, end_time, alt_lo, alt_hi, polygon=scd.make_polygon(latlngrect=area)
    )
    try:
        op_intent_refs = scd_client.query_operational_intent_references(
            resources.utm_client, vol4
        )
    except (
        ValueError,
        scd_client.OperationError,
        requests.exceptions.ConnectionError,
        ConnectionError,
    ) as e:
        msg = "Error querying operational intents: {}".format(e)
        return (
            ClearAreaResponse(
                outcome=ClearAreaOutcome(
                    success=False,
                    message=msg,
                    timestamp=StringBasedDateTime(datetime.utcnow()),
                ),
                request=req,
            ),
            200,
        )

    # Try to delete every operational intent found
    dss_deletion_results = {}
    deleted = set()
    for op_intent_ref in op_intent_refs:
        try:
            scd_client.delete_operational_intent_reference(
                resources.utm_client, op_intent_ref.id, op_intent_ref.ovn
            )
            dss_deletion_results[op_intent_ref.id] = "Deleted from DSS"
            deleted.add(op_intent_ref.id)
        except scd_client.OperationError as e:
            dss_deletion_results[op_intent_ref.id] = str(e)

    # Delete corresponding flight injections and cached operational intents
    deadline = datetime.utcnow() + DEADLOCK_TIMEOUT
    while True:
        pending_flights = set()
        with db as tx:
            flights_to_delete = []
            for flight_id, record in tx.flights.items():
                if record is None or record.locked:
                    pending_flights.add(flight_id)
                    continue
                if record.op_intent_reference.id in deleted:
                    flights_to_delete.append(flight_id)
            for flight_id in flights_to_delete:
                del tx.flights[flight_id]

            cache_deletions = []
            for op_intent_id in deleted:
                if op_intent_id in tx.cached_operations:
                    del tx.cached_operations[op_intent_id]
                    cache_deletions.append(op_intent_id)

        if not pending_flights:
            break
        time.sleep(0.5)

        if datetime.utcnow() > deadline:
            raise RuntimeError(
                f"Deadlock in clear_area while attempting to gain access to flight(s) {', '.join(pending_flights)}"
            )

    msg = yaml.dump(
        {
            "dss_deletions": dss_deletion_results,
            "flight_deletions": flights_to_delete,
            "cache_deletions": cache_deletions,
        }
    )
    return (
        ClearAreaResponse(
            outcome=ClearAreaOutcome(
                success=True,
                message=msg,
                timestamp=StringBasedDateTime(datetime.utcnow()),
            ),
            request=req,
        ),
        200,
    )
