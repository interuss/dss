from time import timezone

from datetime import datetime, timedelta, timezone
from typing import Optional

from monitoring.monitorlib import scd
from monitoring.monitorlib.clients.scd_automated_testing import QueryError
from monitoring.monitorlib.fetch import scd as fetch_scd
from monitoring.monitorlib.scd import OperationalIntent, Volume4D
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightResponse,
    InjectFlightResult,
)
from implicitdict import ImplicitDict, StringBasedDateTime
from monitoring.uss_qualifier.common_data_definitions import (
    Severity,
    SubjectType,
    IssueSubject,
)
from monitoring.uss_qualifier.resources.flight_planning.automated_test import TestStep
from monitoring.uss_qualifier.scd.executor.errors import TestRunnerError
from monitoring.uss_qualifier.resources.flight_planning.target import TestTarget
from monitoring.uss_qualifier.scd.reports import Issue, TestStepReference


NUMERIC_PRECISION = 0.001


def execute(
    runner, step: TestStep, step_ref: TestStepReference, target: TestTarget
) -> None:
    print(
        "[SCD]     Step: Inject flight {} to {}".format(
            step.inject_flight.name, target.name
        )
    )
    try:
        t_test = datetime.now(timezone.utc)
        t_delta = (
            t_test - step.inject_flight.reference_time.datetime
        ) + step.inject_flight.planning_time.timedelta

        # add the delta to the reference time so in the next iteration the time delta is not the same than in the previous iteration
        step.inject_flight.reference_time = StringBasedDateTime(
            step.inject_flight.reference_time.datetime + t_delta
        )

        for volume in step.inject_flight.test_injection.operational_intent.volumes:
            t_start_adjusted = (volume.time_start.value.datetime + t_delta).replace(
                tzinfo=None
            )
            t_end_adjusted = (volume.time_end.value.datetime + t_delta).replace(
                tzinfo=None
            )

            volume.time_start.value = StringBasedDateTime(t_start_adjusted.isoformat())
            volume.time_end.value = StringBasedDateTime(t_end_adjusted.isoformat())
            break
        resp, query, flight_id = target.inject_flight(step.inject_flight)
    except QueryError as e:
        interaction_id = runner.report_recorder.capture_interaction(
            step_ref, e.query, "Inject flight into USS"
        )
        issue = runner.report_recorder.capture_injection_unknown_issue(
            interaction_id,
            summary="Injection request was unsuccessful",
            details="Injection attempt failed with status {}.".format(
                e.query.status_code
            ),
            target_name=target.name,
            attempt=step.inject_flight,
        )
        raise TestRunnerError(
            "Unsuccessful attempt to inject flight {}".format(step.inject_flight.name),
            issue,
        )
    interaction_id = runner.report_recorder.capture_interaction(
        step_ref, query, "Inject flight into USS"
    )
    runner.evaluate_inject_flight_response(
        interaction_id, target, step.inject_flight, resp
    )

    if (
        runner.dss_target
        and resp.result == InjectFlightResult.Planned
        and InjectFlightResult.Planned
        in step.inject_flight.known_responses.acceptable_results
    ):
        _verify_operational_intent(runner, step, step_ref, target, resp, flight_id)


def _verify_operational_intent(
    runner,
    step: TestStep,
    step_ref: TestStepReference,
    target: TestTarget,
    resp: InjectFlightResponse,
    flight_id: str,
) -> None:
    # Verify that the operation actually does exist in the system

    # Check the DSS for a reference
    op_intent = step.inject_flight.test_injection.operational_intent
    vol4s = op_intent.volumes + op_intent.off_nominal_volumes
    rect = scd.rect_bounds_of(vol4s)
    t0 = scd.start_of(vol4s)
    t1 = scd.end_of(vol4s)
    alts = scd.meter_altitude_bounds_of(vol4s)
    op_refs = fetch_scd.operational_intent_references(
        runner.dss_target.client, rect, t0, t1, alts[0], alts[1]
    )
    dss_interaction_id = runner.report_recorder.capture_interaction(
        step_ref, op_refs, "Check if injected operational intent exists in DSS"
    )
    if not op_refs.success:
        # The DSS call didn't even succeed
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=[],
            severity=Severity.Critical,
            subject=None,
            summary="Error querying operational intent references from DSS",
            details=op_refs.error,
            interactions=[dss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)
    try:
        ImplicitDict.parse(
            op_refs.json_result, scd.QueryOperationalIntentReferenceResponse
        )
    except ValueError as e:
        # The DSS returned an invalid result
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=[],
            severity=Severity.Critical,
            subject=None,
            summary="Error in operational intent reference data format from DSS",
            details=str(e),
            interactions=[dss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)
    if resp.operational_intent_id not in op_refs.references_by_id:
        # The expected operational intent reference wasn't present
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=["F3548-21 USS0005"],
            severity=Severity.High,
            subject=IssueSubject(
                subject_type=SubjectType.OperationalIntent,
                subject=resp.operational_intent_id,
            ),
            summary="Operational intent not present in DSS",
            details="When queried for the operational intent {} for flight {} in the DSS, it was not found in the Volume4D of interest".format(
                resp.operational_intent_id, flight_id
            ),
            interactions=[dss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)
    op_ref = op_refs.references_by_id[resp.operational_intent_id]

    # Check the USS for the operational intent details
    op = fetch_scd.operational_intent(
        op_ref["uss_base_url"], resp.operational_intent_id, runner.dss_target.client
    )
    uss_interaction_id = runner.report_recorder.capture_interaction(
        step_ref, op, "Inspect operational intent details"
    )
    if not op.success:
        # The USS call didn't succeed
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=["F3548-21 USS0105"],
            severity=Severity.High,
            subject=IssueSubject(
                subject_type=SubjectType.OperationalIntent,
                subject=resp.operational_intent_id,
            ),
            summary="Error querying operational intent details from USS",
            details=op.error,
            interactions=[uss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)
    try:
        op_resp: scd.GetOperationalIntentDetailsResponse = ImplicitDict.parse(
            op.json_result, scd.GetOperationalIntentDetailsResponse
        )
    except ValueError as e:
        # The USS returned an invalid result
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=["F3548-21 USS0105"],
            severity=Severity.High,
            subject=IssueSubject(
                subject_type=SubjectType.OperationalIntent,
                subject=resp.operational_intent_id,
            ),
            summary="Error in operational intent data format from USS",
            details=str(e),
            interactions=[uss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)

    error_text = validate_op_intent_details(
        op_resp.operational_intent,
        scd.bounding_vol4(op_intent.volumes + op_intent.off_nominal_volumes),
    )
    if error_text:
        # The USS's flight details are incorrect
        issue = Issue(
            context=runner.context,
            check_code="CREATED_FLIGHT_EXISTS_IN_SYSTEM",
            uss_role=step.inject_flight.injection_target.uss_role,
            target=target.name,
            relevant_requirements=[],
            severity=Severity.High,
            subject=IssueSubject(
                subject_type=SubjectType.OperationalIntent,
                subject=resp.operational_intent_id,
            ),
            summary="Operational intent details does not match injected flight",
            details=error_text,
            interactions=[uss_interaction_id],
        )
        raise TestRunnerError(issue.summary, issue)


def validate_op_intent_details(
    operational_intent: OperationalIntent, expected_extent: Volume4D
) -> Optional[str]:
    # Check that the USS is providing reasonable details
    resp_vol4s = (
        operational_intent.details.volumes
        + operational_intent.details.off_nominal_volumes
    )
    resp_alts = scd.meter_altitude_bounds_of(resp_vol4s)
    resp_start = scd.start_of(resp_vol4s)
    resp_end = scd.end_of(resp_vol4s)
    error_text = None
    if resp_alts[0] > expected_extent.volume.altitude_lower.value + NUMERIC_PRECISION:
        error_text = "Lower altitude specified by USS in operational intent details ({} m WGS84) is above the lower altitude in the injected flight ({} m WGS84)".format(
            resp_alts[0], expected_extent.volume.altitude_lower.value
        )
    elif resp_alts[1] < expected_extent.volume.altitude_upper.value - NUMERIC_PRECISION:
        error_text = "Upper altitude specified by USS in operational intent details ({} m WGS84) is below the upper altitude in the injected flight ({} m WGS84)".format(
            resp_alts[1], expected_extent.volume.altitude_upper.value
        )
    elif resp_start > expected_extent.time_start.value.datetime + timedelta(
        seconds=NUMERIC_PRECISION
    ):
        error_text = "Start time specified by USS in operational intent details ({}) is past the start time of the injected flight ({})".format(
            resp_start.isoformat(), expected_extent.time_start.value
        )
    elif resp_end < expected_extent.time_end.value.datetime - timedelta(
        seconds=NUMERIC_PRECISION
    ):
        error_text = "End time specified by USS in operational intent details ({}) is prior to the end time of the injected flight ({})".format(
            resp_end.isoformat(), expected_extent.time_end.value
        )
    return error_text
