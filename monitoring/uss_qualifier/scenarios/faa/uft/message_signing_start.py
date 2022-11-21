from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.interuss.mock_uss import (
    MockUSSResource,
    MockUSSClient,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario


class StartMessageSigningReport(TestScenario):
    _mock_uss: MockUSSClient

    def __init__(self, mock_uss: MockUSSResource):
        super().__init__()
        self._mock_uss = mock_uss.mock_uss

    def run(self):
        self.begin_test_scenario()

        self.begin_test_case("Start message signing")

        self.begin_test_step("Check mock USS readiness")

        query = self._mock_uss.get_status()
        self.record_query(query)

        with self.check(
            "Status ok", participants=[self._mock_uss.participant_id]
        ) as check:
            if query.status_code != 200:
                check.record_failed(
                    summary="Failed to get status from mock USS",
                    details=f"Status code {query.status_code}",
                    severity=Severity.High,
                    query_timestamps=[query.request.timestamp],
                )
                return  # Return if this scenario cannot continue

        with self.check("Ready", participants=[self._mock_uss.participant_id]) as check:
            status = query.response.get("json", {}).get("status", "<No status found>")
            if status != "Ready":
                check.record_failed(
                    summary="Mock USS SCD functionality not ready",
                    details=f"Status indicated as: {status}",
                    severity=Severity.High,
                    query_timestamps=[query.request.timestamp],
                )
                return

        self.end_test_step()  # Check mock USS readiness

        self.begin_test_step("Signal mock USS")

        # TODO: Add call to mock USS to start message signing report
        with self.check(
            "Successful start", participants=[self._mock_uss.participant_id]
        ) as check:
            if False:  # TODO: Insert appropriate check
                check.record_failed(
                    summary="Failed to start message signing report",
                    details="TODO",
                    severity=Severity.High,
                    query_timestamps=[],
                )
                return

        self.end_test_step()  # Signal mock USS

        self.end_test_case()  # Start message signing

        self.end_test_scenario()
