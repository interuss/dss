import traceback

from monitoring.monitorlib.clients.scd_automated_testing import QueryError
from monitoring.monitorlib.scd import bounding_vol4
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    InjectFlightResult,
    DeleteFlightResult,
    Capability,
)
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.astm.f3548.v21 import DSSInstanceResource
from monitoring.uss_qualifier.resources.astm.f3548.v21.dss import DSSInstance
from monitoring.uss_qualifier.resources.flight_planning import (
    FlightIntentsResource,
    FlightPlannersResource,
)
from monitoring.uss_qualifier.resources.flight_planning.target import TestTarget
from monitoring.uss_qualifier.scd.executor.test_steps.inject_flight import (
    validate_op_intent_details,
)
from monitoring.uss_qualifier.scenarios import TestScenario


class NominalPlanning(TestScenario):
    first_flight: InjectFlightRequest
    conflicting_flight: InjectFlightRequest
    uss1: TestTarget
    uss2: TestTarget
    dss: DSSInstance

    def __init__(
        self,
        flight_intents: FlightIntentsResource,
        flight_planners: FlightPlannersResource,
        dss: DSSInstanceResource,
    ):
        super().__init__()
        if len(flight_planners.flight_planners) != 2:
            raise ValueError(
                f"`{self.me()}` TestScenario requires exactly 2 flight_planners; found {len(flight_planners.flight_planners)}"
            )
        self.uss1, self.uss2 = flight_planners.flight_planners

        flight_intents = flight_intents.get_flight_intents()
        if len(flight_intents) < 2:
            raise ValueError(
                f"`{self.me()}` TestScenario requires at least 2 flight_intents; found {len(flight_intents)}"
            )
        self.first_flight, self.conflicting_flight = flight_intents

        self.dss = dss.dss

    def run(self):
        information = "\n".join(
            [
                f"First-mover USS: {self.uss1.config.participant_id} at {self.uss2.config.injection_base_url}",
                f"Second USS: {self.uss2.config.participant_id} at {self.uss2.config.injection_base_url}",
            ]
        )
        self.begin_test_scenario(information=information)

        self.begin_test_case("Setup")
        if not self._setup():
            return
        self.end_test_case()

        self.begin_test_case("Plan first flight")
        if not self._plan_first_flight():
            return
        self.end_test_case()

        self.begin_test_case("Attempt second flight")
        if not self._attempt_second_flight():
            return
        self.end_test_case()

        self.end_test_scenario()

    def _setup(self) -> bool:
        self.begin_test_step("Check for necessary capabilities")

        for uss in (self.uss1, self.uss2):
            try:
                uss_info = uss.get_target_information()
            except QueryError as e:
                stacktrace = "".join(
                    traceback.format_exception(
                        etype=type(e), value=e, tb=e.__traceback__
                    )
                )
                self.record_failed_check(
                    name="Valid responses",
                    summary="Failed to query planner information",
                    severity=Severity.Medium,
                    relevant_participants=[uss.participant_id],
                    details=stacktrace,
                )
                continue
            self.record_query(uss_info.version_query)
            self.record_query(uss_info.capabilities_query)
            if Capability.BasicStrategicConflictDetection not in uss_info.capabilities:
                self.record_failed_check(
                    name="Support BasicStrategicConflictDetection",
                    summary="Planner does not support basic strategic conflict detection",
                    severity=Severity.High,
                    relevant_participants=[uss.participant_id],
                    details=f"Reported capabilities: ({', '.join(uss_info.capabilities)})",
                    query_timestamps=[uss_info.capabilities_query.request.timestamp],
                )
                return False

        self.end_test_step()  # Check for necessary capabilities

        self.begin_test_step("Area clearing")

        extent = bounding_vol4(
            self.first_flight.operational_intent.volumes
            + self.first_flight.operational_intent.off_nominal_volumes
            + self.conflicting_flight.operational_intent.volumes
            + self.conflicting_flight.operational_intent.off_nominal_volumes
        )
        for uss in (self.uss1, self.uss2):
            resp, query = uss.clear_area(extent)
            self.record_query(query)
            if query.status_code != 200:
                self.record_failed_check(
                    name="Area cleared successfully",
                    summary="Error occurred attempting to clear area",
                    severity=Severity.High,
                    relevant_participants=[uss.participant_id],
                    details=f"Status code {query.status_code}",
                    query_timestamps=[query.request.timestamp],
                )
                return False
            if not resp.outcome.success:
                self.record_failed_check(
                    name="Area cleared successfully",
                    summary="Area could not be cleared",
                    severity=Severity.High,
                    relevant_participants=[uss.participant_id],
                    details=f'Participant indicated "{resp.outcome.message}"',
                    query_timestamps=[query.request.timestamp],
                )
                return False

        self.end_test_step()  # Area clearing
        return True

    def _plan_first_flight(self) -> bool:
        self.begin_test_step("Inject flight intent")

        resp, query, flight_id = self.uss1.request_flight(self.first_flight)
        self.record_query(query)
        if resp.result == InjectFlightResult.ConflictWithFlight:
            self.record_failed_check(
                name="No conflict",
                summary="Conflict-free flight not created due to conflict",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f'{self.uss1.participant_id} indicated ConflictWithFlight: "{resp.notes}"',
                query_timestamps=[query.request.timestamp],
            )
            return False
        if resp.result == InjectFlightResult.Rejected:
            self.record_failed_check(
                name="Rejection",
                summary="Valid flight rejected",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f'{self.uss1.participant_id} indicated Rejected: "{resp.notes}"',
                query_timestamps=[query.request.timestamp],
            )
            return False
        if resp.result == InjectFlightResult.Failed:
            self.record_failed_check(
                name="Failure",
                summary="Failed to create flight",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f'{self.uss1.participant_id} Failed to process the user flight intent: "{resp.notes}"',
                query_timestamps=[query.request.timestamp],
            )
            return False
        op_intent_id = resp.operational_intent_id

        self.end_test_step()  # Inject flight intent

        self.begin_test_step("Validate flight creation")
        # TODO
        self.end_test_step()  # Validate flight creation

        self.begin_test_step("Validate flight sharing")
        extent = bounding_vol4(
            self.first_flight.operational_intent.volumes
            + self.first_flight.operational_intent.off_nominal_volumes
        )
        op_intent_refs, query = self.dss.find_op_intent(extent)
        self.record_query(query)
        if query.status_code != 200:
            self.record_failed_check(
                name="DSS response",
                summary="Failed to query DSS for operational intents",
                severity=Severity.High,
                relevant_participants=[self.dss.participant_id],
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
            self.record_failed_check(
                name="Operational intent shared correctly",
                summary="Operational intent reference not found in DSS",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f"USS {self.uss1.participant_id} indicated that it created an operational intent with ID {op_intent_id}, but no operational intent references with that ID were found in the DSS in the area of the flight intent",
                query_timestamps=[query.request.timestamp],
            )
            return False
        op_intent_ref = matching_op_intent_refs[0]

        op_intent, query = self.dss.get_full_op_intent(op_intent_ref)
        if query.status_code != 200:
            self.record_failed_check(
                name="Operational intent shared correctly",
                summary="Operational intent details could not be retrieved from USS",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f"Received status code {query.status_code} from {self.uss1.participant_id} when querying for details of operational intent {op_intent_id}",
                query_timestamps=[query.request.timestamp],
            )
            return False

        error_text = validate_op_intent_details(op_intent, extent)
        if error_text:
            self.record_failed_check(
                name="Correct operational intent details",
                summary="Operational intent details do not match user flight intent",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=error_text,
                query_timestamps=[query.request.timestamp],
            )
            return False

        self.end_test_step()  # Validate flight sharing

        return True

    def _attempt_second_flight(self):
        self.begin_test_step("Inject flight intent")

        resp, query, flight_id = self.uss2.request_flight(self.conflicting_flight)
        self.record_query(query)
        if resp.result == InjectFlightResult.Planned:
            self.record_failed_check(
                name="Incorrectly planned",
                summary="Conflict-free flight not created due to conflict",
                severity=Severity.High,
                relevant_participants=[self.uss2.participant_id],
                details="The user's intended flight conflicts with an existing operational intent so the result of attempting to fulfill this flight intent should not be a successful planning of the flight.",
                query_timestamps=[query.request.timestamp],
            )
            return False
        if resp.result == InjectFlightResult.Failed:
            self.record_failed_check(
                name="Failure",
                summary="Failed to create flight",
                severity=Severity.High,
                relevant_participants=[self.uss1.participant_id],
                details=f'{self.uss1.participant_id} Failed to process the user flight intent: "{resp.notes}"',
                query_timestamps=[query.request.timestamp],
            )
            return False

        self.end_test_step()  # Inject flight intent
        return True

    def cleanup(self):
        self.begin_cleanup()

        for uss in (self.uss2, self.uss1):
            while uss.created_flight_ids:
                name = next(iter(uss.created_flight_ids))
                flight_id = uss.created_flight_ids.pop(name)
                resp, query = uss.cleanup_flight(flight_id)
                self.record_query(query)
                if resp.result != DeleteFlightResult.Closed:
                    self.record_failed_check(
                        name="Successful flight deletion",
                        summary="Failed to delete flight",
                        severity=Severity.Medium,
                        relevant_participants=[uss.participant_id],
                        details="",
                        query_timestamps=[query.request.timestamp],
                    )

        self.end_cleanup()
