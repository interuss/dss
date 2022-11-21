from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.interuss.mock_uss import (
    MockUSSResource,
    MockUSSClient,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario


class FinalizeMessageSigningReport(TestScenario):
    _mock_uss: MockUSSClient

    def __init__(self, mock_uss: MockUSSResource):
        super().__init__()
        self._mock_uss = mock_uss.mock_uss

    def run(self):
        self.begin_test_scenario()

        self.begin_test_case("Finalize message signing")

        self.begin_test_step("Signal mock USS")

        # TODO: Add call to mock USS to finalize message signing report
        with self.check(
            "Successful finalization", participants=[self._mock_uss.participant_id]
        ) as check:
            if False:  # TODO: Insert appropriate check
                check.record_failed(
                    summary="Failed to finalize message signing report",
                    details="TODO",
                    severity=Severity.High,
                    query_timestamps=[],
                )
                return

        self.end_test_step()  # Signal mock USS

        self.end_test_case()  # Start message signing

        self.end_test_scenario()
