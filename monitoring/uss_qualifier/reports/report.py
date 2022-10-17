from datetime import datetime
import traceback
from typing import List, Optional

from implicitdict import ImplicitDict, StringBasedDateTime

from monitoring.monitorlib import fetch, inspection
from monitoring.uss_qualifier.common_data_definitions import Severity


class FailedCheck(ImplicitDict):
    name: str
    """Name of the check that failed"""

    documentation_url: str
    """URL at which the check which failed is described"""

    timestamp: StringBasedDateTime
    """Time the issue was discovered"""

    summary: str
    """Human-readable summary of the issue"""

    details: str
    """Human-readable description of the issue"""

    relevant_requirements: List[str]
    """Requirements that this issue relates to"""

    severity: Severity
    """How severe the issue is"""

    relevant_participants: List[str]
    """Participant IDs of actors or organizations to which this failure may be relevant"""

    query_report_timestamps: Optional[List[str]]
    """List of the `report` timestamp field for queries relevant to this failed check"""

    additional_data: Optional[dict]
    """Additional data, structured according to the checks' needs, that may be relevant for understanding this failed check"""


class TestStepReport(ImplicitDict):
    name: str
    """Name of this test step"""

    documentation_url: str
    """URL at which this test step is described"""

    start_time: StringBasedDateTime
    """Time at which the test step started"""

    queries: Optional[List[fetch.Query]]
    """Description of HTTP requests relevant to this issue"""

    failed_checks: List[FailedCheck]
    """The checks which failed in this test step"""

    end_time: Optional[StringBasedDateTime]
    """Time at which the test step completed or encountered an error"""


class TestCaseReport(ImplicitDict):
    name: str
    """Name of this test case"""

    documentation_url: str
    """URL at which this test case is described"""

    start_time: StringBasedDateTime
    """Time at which the test case started"""

    end_time: Optional[StringBasedDateTime]
    """Time at which the test case completed or encountered an error"""

    steps: List[TestStepReport]
    """Reports for each of the test steps in this test case"""


class ErrorReport(ImplicitDict):
    type: str
    """Type of error"""

    message: str
    """Error message"""

    timestamp: StringBasedDateTime
    """Time at which the error was logged"""

    stacktrace: str
    """Full stack trace of error"""

    @staticmethod
    def create_from_exception(e: Exception):
        return ErrorReport(
            type=str(inspection.fullname(e.__class__)),
            message=str(e),
            timestamp=StringBasedDateTime(datetime.utcnow()),
            stacktrace="".join(
                traceback.format_exception(etype=type(e), value=e, tb=e.__traceback__)
            ),
        )


class TestScenarioReport(ImplicitDict):
    name: str
    """Name of this test scenario"""

    documentation_url: str
    """URL at which this test scenario is described"""

    start_time: StringBasedDateTime
    """Time at which the test scenario started"""

    end_time: Optional[StringBasedDateTime]
    """Time at which the test scenario completed or encountered an error"""

    successful: bool = False
    """True iff test scenario completed normally with no failed checks"""

    cases: List[TestCaseReport]
    """Reports for each of the test cases in this test scenario"""

    execution_error: Optional[ErrorReport]
    """If there was an error while executing this test scenario, this field describes the error"""


class TestSuiteActionReport(ImplicitDict):
    test_suite: Optional["TestSuiteReport"]
    """If this action was a test suite, this field will hold its report"""

    test_scenario: Optional[TestScenarioReport]
    """If this action was a test scenario, this field will hold its report"""

    def successful(self) -> bool:
        test_suite = "test_suite" in self and self.test_suite is not None
        test_scenario = "test_scenario" in self and self.test_scenario is not None
        if sum(1 if case else 0 for case in [test_suite, test_scenario]) != 1:
            raise ValueError(
                "Exactly one of `test_suite` or `test_scenario` must be populated"
            )
        if test_suite:
            return self.test_suite.successful
        if test_scenario:
            return self.test_scenario.successful

        # This line should not be possible to reach
        raise RuntimeError("Case selection logic failed for TestSuiteActionReport")


class TestSuiteReport(ImplicitDict):
    name: str
    """Name of this test suite"""

    documentation_url: str
    """URL at which this test suite is described"""

    start_time: StringBasedDateTime
    """Time at which the test suite started"""

    actions: List[TestSuiteActionReport]
    """Reports from test scenarios and test suites comprising the test suite for this report"""

    end_time: Optional[StringBasedDateTime]
    """Time at which the test suite completed"""

    successful: bool = False
    """True iff test suite completed normally with no failed checks"""
