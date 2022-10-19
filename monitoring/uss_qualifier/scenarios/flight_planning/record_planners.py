from monitoring.uss_qualifier.resources.flight_planning import FlightPlannersResource
from monitoring.uss_qualifier.scenarios import TestScenario


class RecordPlanners(TestScenario):
    _flight_planners: FlightPlannersResource

    def __init__(self, flight_planners: FlightPlannersResource):
        super().__init__()
        self._flight_planners = flight_planners

    def run(self):
        information = "Available flight planners:\n" + "\n".join(
            f"* {fp.config.participant_id}: {fp.config.injection_base_url}"
            for fp in self._flight_planners.flight_planners
        )
        self.begin_test_scenario(information)
        self.end_test_scenario()
