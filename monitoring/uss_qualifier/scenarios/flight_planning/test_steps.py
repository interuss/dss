import traceback
from datetime import datetime
from typing import List, Dict, Union, Optional, Tuple

from monitoring.monitorlib.clients.scd_automated_testing import QueryError
from monitoring.monitorlib.scd import bounding_vol4
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    Capability,
    InjectFlightResult,
    InjectFlightResponse,
    DeleteFlightResult,
)
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.flight_planning.target import TestTarget
from monitoring.uss_qualifier.scenarios.scenario import TestScenarioType
from monitoring.uss_qualifier.scenarios.astm.utm.evaluation import (
    validate_op_intent_details,
)


def clear_area(
    scenario: TestScenarioType,
    test_step: str,
    flight_intents: List[InjectFlightRequest],
    flight_planners: List[TestTarget],
) -> bool:
    """Perform a test step to clear the area that will be used in the scenario.

    This function assumes:
    * `scenario` is ready to execute a test step
    * "Area cleared successfully" check declared for specified test step in `scenario`'s documentation

    Returns: False if the scenario should stop, True otherwise.
    """
    scenario.begin_test_step(test_step)

    volumes = []
    for flight_intent in flight_intents:
        volumes += flight_intent.operational_intent.volumes
        volumes += flight_intent.operational_intent.off_nominal_volumes
    extent = bounding_vol4(volumes)
    for uss in flight_planners:
        resp, query = uss.clear_area(extent)
        scenario.record_query(query)
        if query.status_code != 200:
            scenario.record_failed_check(
                name="Area cleared successfully",
                summary="Error occurred attempting to clear area",
                severity=Severity.High,
                relevant_participants=[uss.participant_id],
                details=f"Status code {query.status_code}",
                query_timestamps=[query.request.timestamp],
            )
            return False
        if not resp.outcome.success:
            scenario.record_failed_check(
                name="Area cleared successfully",
                summary="Area could not be cleared",
                severity=Severity.High,
                relevant_participants=[uss.participant_id],
                details=f'Participant indicated "{resp.outcome.message}"',
                query_timestamps=[query.request.timestamp],
            )
            return False

    scenario.end_test_step()


OneOrMoreTestTargets = Union[TestTarget, List[TestTarget]]
OneOrMoreCapabilities = Union[Capability, List[Capability]]


def check_capabilities(
    scenario: TestScenarioType,
    test_step: str,
    required_capabilities: Optional[
        List[Tuple[OneOrMoreTestTargets, OneOrMoreCapabilities]]
    ] = None,
    prerequisite_capabilities: Optional[
        List[Tuple[OneOrMoreTestTargets, OneOrMoreCapabilities]]
    ] = None,
) -> bool:
    """Perform a check that flight planners support certain capabilities.

    This function assumes:
    * `scenario` is ready to execute a test step
    *  If `required_capabilities` is specified:
      * "Valid responses" check declared for specified test step in `scenario`'s documentation
      * "Support {required_capability}" check declared for specified test in step`scenario`'s documentation

    Args:
      required_capabilities: The specified USSs must support these capabilities.
        If a capability is not supported, a "Valid responses" failed check will
        be created.
      prerequisite_capabilities: If any of the specified USSs do not support
        this capabilities, a "Prerequisite capabilities" note will be added and
        the scenario will be indicated to stop, but no failed check will be
        created.

    Returns:
      False if the scenario should stop, True otherwise.
    """
    scenario.begin_test_step(test_step)

    if required_capabilities is None:
        required_capabilities = []
    if prerequisite_capabilities is None:
        prerequisite_capabilities = []

    # Collect all the flight planners that need to be queried
    all_flight_planners: List[TestTarget] = []
    for flight_planner_list in [p for p, _ in required_capabilities] + [
        p for p, _ in prerequisite_capabilities
    ]:
        if not isinstance(flight_planner_list, list):
            flight_planner_list = [flight_planner_list]
        for flight_planner in flight_planner_list:
            if flight_planner not in all_flight_planners:
                all_flight_planners.append(flight_planner)

    # Query all the flight planners and collect key results
    flight_planner_capabilities: List[Tuple[TestTarget, List[Capability]]] = []
    flight_planner_capability_query_timestamps: List[Tuple[TestTarget, datetime]] = []
    for flight_planner in all_flight_planners:
        try:
            uss_info = flight_planner.get_target_information()
        except QueryError as e:
            stacktrace = "".join(
                traceback.format_exception(etype=type(e), value=e, tb=e.__traceback__)
            )
            scenario.record_failed_check(
                name="Valid responses",
                summary=f"Failed to query {flight_planner.participant_id} for information",
                severity=Severity.Medium,
                relevant_participants=[flight_planner.participant_id],
                details=stacktrace,
            )
            continue
        scenario.record_query(uss_info.version_query)
        scenario.record_query(uss_info.capabilities_query)
        flight_planner_capabilities.append((flight_planner, uss_info.capabilities))
        flight_planner_capability_query_timestamps.append(
            (flight_planner, uss_info.capabilities_query.request.timestamp)
        )

    # Check for required capabilities
    for flight_planners, capabilities in required_capabilities:
        if not isinstance(flight_planners, list):
            flight_planners = [flight_planners]
        if not isinstance(capabilities, list):
            capabilities = [capabilities]
        for flight_planner in flight_planners:
            for required_capability in capabilities:
                available_capabilities = [
                    c for p, c in flight_planner_capabilities if p is flight_planner
                ][0]
                if required_capability not in available_capabilities:
                    timestamp = [
                        t
                        for p, t in flight_planner_capability_query_timestamps
                        if p is flight_planner
                    ][0]
                    scenario.record_failed_check(
                        name=f"Support {required_capability}",
                        summary=f"Flight planner {flight_planner.participant_id} does not support {required_capability}",
                        severity=Severity.High,
                        relevant_participants=[flight_planner.participant_id],
                        details=f"Reported capabilities: ({', '.join(available_capabilities)})",
                        query_timestamps=[timestamp],
                    )
                    return False

    # Check for prerequisite capabilities
    unsupported_prerequisites: List[str] = []
    for flight_planners, capabilities in prerequisite_capabilities:
        if not isinstance(flight_planners, list):
            flight_planners = [flight_planners]
        if not isinstance(capabilities, list):
            capabilities = [capabilities]
        for flight_planner in flight_planners:
            available_capabilities = [
                c for p, c in flight_planner_capabilities if p is flight_planner
            ][0]
            unmet_capabilities = ", ".join(
                c for c in capabilities if c not in available_capabilities
            )
            if unmet_capabilities:
                unsupported_prerequisites.append(
                    f"* {flight_planner.participant_id}: {unmet_capabilities}"
                )
    if unsupported_prerequisites:
        scenario.record_note(
            "Unsupported prerequisite capabilities",
            "\n".join(unsupported_prerequisites),
        )
        return False

    scenario.end_test_step()
    return True


def inject_successful_flight_intent(
    scenario: TestScenarioType,
    test_step: str,
    flight_planner: TestTarget,
    flight_intent: InjectFlightRequest,
) -> Optional[InjectFlightResponse]:
    """Inject a flight intent that should result in success.

    This function assumes:
    * `scenario` is currently ready to execute a test step
    * "Successful planning" check is declared for specified test step in `scenario`'s documentation


    Returns: None if a check failed, otherwise the injection response.
    """
    scenario.begin_test_step(test_step)
    resp, query, flight_id = flight_planner.request_flight(flight_intent)
    scenario.record_query(query)
    if resp.result == InjectFlightResult.ConflictWithFlight:
        scenario.record_failed_check(
            name="Successful planning",
            summary="Conflict-free flight not created due to conflict",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=f'{scenario.uss1.participant_id} indicated ConflictWithFlight: "{resp.notes}"',
            query_timestamps=[query.request.timestamp],
        )
        return None
    if resp.result == InjectFlightResult.Rejected:
        scenario.record_failed_check(
            name="Successful planning",
            summary="Valid flight rejected",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=f'{scenario.uss1.participant_id} indicated Rejected: "{resp.notes}"',
            query_timestamps=[query.request.timestamp],
        )
        return None
    if resp.result == InjectFlightResult.Failed:
        scenario.record_failed_check(
            name="Successful planning",
            summary="Failed to create flight",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=f'{scenario.uss1.participant_id} Failed to process the user flight intent: "{resp.notes}"',
            query_timestamps=[query.request.timestamp],
        )
        return None
    scenario.end_test_step()
    return resp


def validate_shared_operational_intent(
    scenario: TestScenarioType,
    test_step: str,
    flight_intent: InjectFlightRequest,
    op_intent_id: str,
) -> bool:
    """Validate that operational intent information was correctly shared for a flight intent.

    This function assumes:
    * `scenario` is ready to execute a test step
    * "DSS response" check declared for specified test step in `scenario`'s documentation
    * "Operational intent shared correctly" check declared for specified test step in `scenario`'s documentation
    * "Correct operational intent details" check declared for specified test in step`scenario`'s documentation

    Returns:
      False if the scenario should stop, True otherwise.
    """
    scenario.begin_test_step(test_step)
    extent = bounding_vol4(
        flight_intent.operational_intent.volumes
        + flight_intent.operational_intent.off_nominal_volumes
    )
    op_intent_refs, query = scenario.dss.find_op_intent(extent)
    scenario.record_query(query)
    if query.status_code != 200:
        scenario.record_failed_check(
            name="DSS response",
            summary="Failed to query DSS for operational intents",
            severity=Severity.High,
            relevant_participants=[scenario.dss.participant_id],
            details=f"Received status code {query.status_code} from the DSS",
            query_timestamps=[query.request.timestamp],
        )
        return False

    matching_op_intent_refs = [
        op_intent_ref
        for op_intent_ref in op_intent_refs
        if op_intent_ref.id == op_intent_id
    ]
    if not matching_op_intent_refs:
        scenario.record_failed_check(
            name="Operational intent shared correctly",
            summary="Operational intent reference not found in DSS",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=f"USS {scenario.uss1.participant_id} indicated that it created an operational intent with ID {op_intent_id}, but no operational intent references with that ID were found in the DSS in the area of the flight intent",
            query_timestamps=[query.request.timestamp],
        )
        return False
    op_intent_ref = matching_op_intent_refs[0]

    op_intent, query = scenario.dss.get_full_op_intent(op_intent_ref)
    if query.status_code != 200:
        scenario.record_failed_check(
            name="Operational intent shared correctly",
            summary="Operational intent details could not be retrieved from USS",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=f"Received status code {query.status_code} from {scenario.uss1.participant_id} when querying for details of operational intent {op_intent_id}",
            query_timestamps=[query.request.timestamp],
        )
        return False

    error_text = validate_op_intent_details(op_intent, extent)
    if error_text:
        scenario.record_failed_check(
            name="Correct operational intent details",
            summary="Operational intent details do not match user flight intent",
            severity=Severity.High,
            relevant_participants=[scenario.uss1.participant_id],
            details=error_text,
            query_timestamps=[query.request.timestamp],
        )
        return False

    scenario.end_test_step()
    return True


def cleanup_flights(
    self: TestScenarioType, flights: Dict[TestTarget, List[str]]
) -> Dict[TestTarget, List[str]]:
    """Remove flights during a cleanup test step.

    This function assumes:
    * `scenario` is currently cleaning up (cleanup has started)
    * "Successful flight deletion" check declared for cleanup phase in `scenario`'s documentation

    Returns:
      False if the scenario should stop, True otherwise.
    """
    removed_flights: Dict[TestTarget, List[str]] = {}
    for flight_planner, flight_ids in flights.items():
        removed = []
        for flight_id in flight_ids:
            resp, query = flight_planner.cleanup_flight(flight_id)
            self.record_query(query)
            if resp.result == DeleteFlightResult.Closed:
                removed.append(flight_id)
            else:
                self.record_failed_check(
                    name="Successful flight deletion",
                    summary="Failed to delete flight",
                    severity=Severity.Medium,
                    relevant_participants=[flight_planner.participant_id],
                    details="",
                    query_timestamps=[query.request.timestamp],
                )
        removed_flights[flight_planner] = removed
    return removed_flights
