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
from monitoring.uss_qualifier.resources.flight_planning.target import TestTarget
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from monitoring.uss_qualifier.scenarios.flight_planning.test_steps import (
    clear_area,
    check_capabilities,
    inject_successful_flight_intent,
    cleanup_flights,
)


class Validation(TestScenario):
    flight_intents: List[InjectFlightRequest]
    uss: TestTarget

    def __init__(
        self,
        flight_intents: FlightIntentsResource,
        flight_planners: FlightPlannersResource,
    ):
        super().__init__()
        if len(flight_planners.flight_planners) != 1:
            raise ValueError(
                f"`{self.me()}` TestScenario requires exactly 1 flight_planner; found {len(flight_planners.flight_planners)}"
            )
        self.uss = flight_planners.flight_planners[0]

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
                (self.uss, Capability.FlightAuthorisationValidation)
            ],
        ):
            return False

        if not clear_area(
            self,
            "Area clearing",
            self.flight_intents,
            [self.uss],
        ):
            return False

        return True

    def _attempt_invalid_flights(self) -> bool:
        self.begin_test_step("Inject invalid flight intent")

        for flight_intent in self.flight_intents[0:-1]:
            resp, query, flight_id = self.uss.request_flight(flight_intent)
            self.record_query(query)
            if resp.result == InjectFlightResult.Planned:
                problems = ", ".join(
                    problems_with_flight_authorisation(
                        flight_intent.flight_authorisation
                    )
                )
                self.record_failed_check(
                    name="Incorrectly planned",
                    summary="Flight planned with invalid flight authorisation",
                    severity=Severity.Medium,
                    relevant_participants=[self.uss.participant_id],
                    details=f"Flight intent resulted in successful flight planning even though the flight authorisation had: {problems}",
                    query_timestamps=[query.request.timestamp],
                )
            if resp.result == InjectFlightResult.Failed:
                self.record_failed_check(
                    name="Failure",
                    summary="Failed to create flight",
                    severity=Severity.Medium,
                    relevant_participants=[self.uss.participant_id],
                    details=f'{self.uss.participant_id} Failed to process the user flight intent: "{resp.notes}"',
                    query_timestamps=[query.request.timestamp],
                )

            self.end_test_step()  # Inject flight intent

        return True

    def _plan_valid_flight(self) -> bool:
        resp = inject_successful_flight_intent(
            self, "Inject valid flight intent", self.uss, self.flight_intents[-1]
        )
        if resp is None:
            return False

        return True

    def cleanup(self):
        self.begin_cleanup()

        flights = {self.uss: list(self.uss.created_flight_ids.values())}
        flights = cleanup_flights(self, flights)

        names_to_remove = [
            k for k, v in self.uss.created_flight_ids if v in flights[self.uss]
        ]
        for name in names_to_remove:
            del self.uss.created_flight_ids[name]

        self.end_cleanup()
