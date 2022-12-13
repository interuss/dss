from typing import List

from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    InjectFlightResult,
    Capability,
)
from monitoring.monitorlib.uspace import problems_with_flight_authorisation
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.flight_planning import (
    FlightIntentsResource,
    FlightPlannersResource,
)
from monitoring.uss_qualifier.resources.flight_planning.flight_planner import (
    FlightPlanner,
)
from monitoring.uss_qualifier.resources.flight_planning.flight_planners import (
    FlightPlannerResource,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from monitoring.uss_qualifier.scenarios.flight_planning.test_steps import (
    clear_area,
    check_capabilities,
    inject_successful_flight_intent,
    cleanup_flights,
)


class Validation(TestScenario):
    flight_intents: List[InjectFlightRequest]
    ussp: FlightPlanner

    def __init__(
        self,
        flight_intents: FlightIntentsResource,
        flight_planner: FlightPlannerResource,
    ):
        super().__init__()
        self.ussp = flight_planner.flight_planner

        intents = flight_intents.get_flight_intents()
        if len(intents) < 2:
            raise ValueError(
                f"`{self.me()}` TestScenario requires at least 2 flight_intents; found {len(intents)}"
            )
        for i, flight_intent in enumerate(intents[0:-1]):
            problems = problems_with_flight_authorisation(
                flight_intent.flight_authorisation
            )
            if not problems:
                raise ValueError(
                    f"`{self.me()}` TestScenario requires all flight intents except the last to have invalid flight authorisation data.  Instead, intent {i+1}/{len(intents)} had valid flight authorisation data."
                )
        problems = problems_with_flight_authorisation(intents[-1].flight_authorisation)
        if problems:
            problems = ", ".join(problems)
            raise ValueError(
                f"`{self.me()}` TestScenario requires the last flight intent to be valid.  Instead, the flight authorisation data had: {problems}"
            )
        self.flight_intents = intents

    def run(self):
        self.begin_test_scenario()

        self.record_note("Planner", self.ussp.participant_id)

        self.begin_test_case("Setup")
        if not self._setup():
            return
        self.end_test_case()

        self.begin_test_case("Attempt invalid flights")
        if not self._attempt_invalid_flights():
            return
        self.end_test_case()

        self.begin_test_case("Plan valid flight")
        if not self._plan_valid_flight():
            return
        self.end_test_case()

        self.end_test_scenario()

    def _setup(self) -> bool:
        if not check_capabilities(
            self,
            "Check for necessary capabilities",
            required_capabilities=[
                (self.ussp, Capability.FlightAuthorisationValidation)
            ],
        ):
            return False

        clear_area(
            self,
            "Area clearing",
            self.flight_intents,
            [self.ussp],
        )

        return True

    def _attempt_invalid_flights(self) -> bool:
        self.begin_test_step("Inject invalid flight intents")

        for flight_intent in self.flight_intents[0:-1]:
            resp, query, flight_id = self.ussp.request_flight(flight_intent)
            self.record_query(query)
            with self.check("Incorrectly planned", [self.ussp.participant_id]) as check:
                if resp.result == InjectFlightResult.Planned:
                    problems = ", ".join(
                        problems_with_flight_authorisation(
                            flight_intent.flight_authorisation
                        )
                    )
                    check.record_failed(
                        summary="Flight planned with invalid flight authorisation",
                        severity=Severity.Medium,
                        details=f"Flight intent resulted in successful flight planning even though the flight authorisation had: {problems}",
                        query_timestamps=[query.request.timestamp],
                    )
            with self.check("Failure", [self.ussp.participant_id]) as check:
                if resp.result == InjectFlightResult.Failed:
                    check.record_failed(
                        summary="Failed to create flight",
                        severity=Severity.Medium,
                        details=f'{self.ussp.participant_id} Failed to process the user flight intent: "{resp.notes}"',
                        query_timestamps=[query.request.timestamp],
                    )

            self.end_test_step()  # Inject flight intents

        return True

    def _plan_valid_flight(self) -> bool:
        resp, _ = inject_successful_flight_intent(
            self, "Inject valid flight intent", self.ussp, self.flight_intents[-1]
        )
        if resp is None:
            return False

        return True

    def cleanup(self):
        self.begin_cleanup()
        cleanup_flights(self, [self.ussp])
        self.end_cleanup()
