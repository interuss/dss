from monitoring.monitorlib.scd import bounding_vol4
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
)
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.scenarios.astm.utm.evaluation import (
    validate_op_intent_details,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenarioType


def validate_shared_operational_intent(
    scenario: TestScenarioType,
    test_step: str,
    flight_intent: InjectFlightRequest,
    op_intent_id: str,
) -> bool:
    """Validate that operational intent information was correctly shared for a flight intent.

    This function implements the test step described in
    validate_shared_operational_intent.md.

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
    with scenario.check("DSS response", [scenario.dss.participant_id]) as check:
        if query.status_code != 200:
            check.record_failed(
                summary="Failed to query DSS for operational intents",
                severity=Severity.High,
                details=f"Received status code {query.status_code} from the DSS",
                query_timestamps=[query.request.timestamp],
            )
            return False

    matching_op_intent_refs = [
        op_intent_ref
        for op_intent_ref in op_intent_refs
        if op_intent_ref.id == op_intent_id
    ]
    with scenario.check(
        "Operational intent shared correctly", [scenario.uss1.participant_id]
    ) as check:
        if not matching_op_intent_refs:
            check.record_failed(
                summary="Operational intent reference not found in DSS",
                severity=Severity.High,
                details=f"USS {scenario.uss1.participant_id} indicated that it created an operational intent with ID {op_intent_id}, but no operational intent references with that ID were found in the DSS in the area of the flight intent",
                query_timestamps=[query.request.timestamp],
            )
            return False
    op_intent_ref = matching_op_intent_refs[0]

    op_intent, query = scenario.dss.get_full_op_intent(op_intent_ref)
    with scenario.check(
        "Operational intent shared correctly", [scenario.uss1.participant_id]
    ) as check:
        if query.status_code != 200:
            check.record_failed(
                summary="Operational intent details could not be retrieved from USS",
                severity=Severity.High,
                details=f"Received status code {query.status_code} from {scenario.uss1.participant_id} when querying for details of operational intent {op_intent_id}",
                query_timestamps=[query.request.timestamp],
            )
            return False

    error_text = validate_op_intent_details(op_intent, extent)
    with scenario.check(
        "Correct operational intent details", [scenario.uss1.participant_id]
    ) as check:
        if error_text:
            check.record_failed(
                summary="Operational intent details do not match user flight intent",
                severity=Severity.High,
                details=error_text,
                query_timestamps=[query.request.timestamp],
            )
            return False

    scenario.end_test_step()
    return True
