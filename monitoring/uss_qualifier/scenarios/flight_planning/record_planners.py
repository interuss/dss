from monitoring.uss_qualifier.resources.flight_planning import FlightPlannersResource
from monitoring.uss_qualifier.scenarios import TestScenario


class RecordPlanners(TestScenario):
    _flight_planners: FlightPlannersResource

    def __init__(self, flight_planners: FlightPlannersResource):
        super().__init__()
        self._flight_planners = flight_planners

    def run(self):
        self.begin_test_scenario()

        self.record_note(
            "Available flight planners",
            "\n".join(
                f"* {fp.config.participant_id}: {fp.config.injection_base_url}"
                for fp in self._flight_planners.flight_planners
            ),
        )

        self.end_test_scenario()
