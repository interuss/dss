from uas_standards.astm.f3548.v21.api import OperationalIntentState

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
        "Operational intent details retrievable", [scenario.uss1.participant_id]
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

    with scenario.check("Off-nominal volumes", [scenario.uss1.participant_id]) as check:
        if (
            op_intent.reference.state == OperationalIntentState.Accepted
            or op_intent.reference.state == OperationalIntentState.Activated
        ) and op_intent.details.get("off_nominal_volumes", None):
            check.record_failed(
                summary="Accepted or Activated operational intents are not allowed off-nominal volumes",
                severity=Severity.Medium,
                details=f"Operational intent {op_intent.reference.id} was {op_intent.reference.state} and had {len(op_intent.details.off_nominal_volumes)} off-nominal volumes",
                query_timestamps=[query.request.timestamp],
            )

    all_volumes = op_intent.details.get("volumes", []) + op_intent.details.get(
        "off_nominal_volumes", []
    )

    def volume_vertices(v4):
        if "outline_circle" in v4.volume:
            return 1
        if "outline_polygon" in v4.volume:
            return len(v4.volume.outline_polygon.vertices)

    n_vertices = sum(volume_vertices(v) for v in all_volumes)
    with scenario.check("Vertices", [scenario.uss1.participant_id]) as check:
        if n_vertices > 10000:
            check.record_failed(
                summary="Too many vertices",
                severity=Severity.Medium,
                details=f"Operational intent {op_intent.reference.id} had {n_vertices} vertices total",
                query_timestamps=[query.request.timestamp],
            )

    scenario.end_test_step()
    return True
