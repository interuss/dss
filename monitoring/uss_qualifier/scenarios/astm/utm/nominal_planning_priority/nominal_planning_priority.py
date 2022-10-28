from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
    Capability,
)
from monitoring.uss_qualifier.resources.astm.f3548.v21 import DSSInstanceResource
from monitoring.uss_qualifier.resources.astm.f3548.v21.dss import DSSInstance
from monitoring.uss_qualifier.resources.flight_planning import (
    FlightIntentsResource,
    FlightPlannersResource,
)
from monitoring.uss_qualifier.resources.flight_planning.flight_planner import (
    FlightPlanner,
)
from monitoring.uss_qualifier.scenarios.astm.utm.test_steps import (
    validate_shared_operational_intent,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from monitoring.uss_qualifier.scenarios.flight_planning.test_steps import (
    clear_area,
    check_capabilities,
    inject_successful_flight_intent,
    cleanup_flights,
)


class NominalPlanningPriority(TestScenario):
    first_flight: InjectFlightRequest
    priority_flight: InjectFlightRequest
    uss1: FlightPlanner
    uss2: FlightPlanner
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
        self.first_flight, self.priority_flight = flight_intents

        self.dss = dss.dss

    def run(self):
        self.begin_test_scenario()

        self.record_note(
            "First USS",
            f"{self.uss1.config.participant_id}",
        )
        self.record_note(
            "Priority USS",
            f"{self.uss2.config.participant_id}",
        )

        self.begin_test_case("Setup")
        if not self._setup():
            return
        self.end_test_case()

        self.begin_test_case("Plan first flight")
        if not self._plan_first_flight():
            return
        self.end_test_case()

        self.begin_test_case("Plan priority flight")
        if not self._plan_priority_flight():
            return
        self.end_test_case()

        self.begin_test_case("Activate priority flight")
        if not self._activate_priority_flight():
            return
        self.end_test_case()

        self.end_test_scenario()

    def _setup(self) -> bool:
        if not check_capabilities(
            self,
            "Check for necessary capabilities",
            required_capabilities=[
                ([self.uss1, self.uss2], Capability.BasicStrategicConflictDetection)
            ],
            prerequisite_capabilities=[(self.uss2, Capability.HighPriorityFlights)],
        ):
            return False

        if not clear_area(
            self,
            "Area clearing",
            [self.first_flight, self.priority_flight],
            [self.uss1, self.uss2],
        ):
            return False

        return True

    def _plan_first_flight(self) -> bool:
        resp = inject_successful_flight_intent(
            self, "Inject flight intent", self.uss1, self.first_flight
        )
        if resp is None:
            return False
        op_intent_id = resp.operational_intent_id

        self.begin_test_step("Validate flight creation")
        # TODO
        self.end_test_step()  # Validate flight creation

        validate_shared_operational_intent(
            self, "Validate flight sharing", self.first_flight, op_intent_id
        )

        return True

    def _plan_priority_flight(self) -> bool:
        resp = inject_successful_flight_intent(
            self, "Inject flight intent", self.uss2, self.priority_flight
        )
        if resp is None:
            return False
        op_intent_id = resp.operational_intent_id

        self.begin_test_step("Validate flight creation")
        # TODO
        self.end_test_step()  # Validate flight creation

        validate_shared_operational_intent(
            self, "Validate flight sharing", self.priority_flight, op_intent_id
        )

        return True

    def _activate_priority_flight(self) -> bool:

        return True

    def cleanup(self):
        self.begin_cleanup()
        cleanup_flights(self, (self.uss2, self.uss1))
        self.end_cleanup()
