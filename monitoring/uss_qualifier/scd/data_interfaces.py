from typing import Optional, List, Dict
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest
from monitoring.uss_qualifier.common_data_definitions import Severity


class KnownIssueFields(ImplicitDict):
    """Information, which can be defined at the time of test design, about a problem detected by an automated test when a USS provides a response that is not same as the expected result"""
    test_code: str
    """Code corresponding to check generating this issue"""

    relevant_requirements: List[str] = []
    """Requirements that this issue relates to"""

    severity: Severity
    """How severe the issue is"""

    subject: Optional[str]
    """Identifier of the subject of this issue, if applicable. This may be a UAS serial number, or any field or other object central to the issue."""

    summary: str
    """Human-readable summary of the issue"""

    details: str
    """Human-readable description of the issue"""


class KnownResponses(ImplicitDict):
    """Mapping of the flight injection attempt's USS response to test outcome"""
    acceptable_results: List[str]
    """Acceptable values in the result data field of InjectFlightResponse. The flight injection attempt will be considered successful if the USS under test reports one of these as the result of attempting to inject the flight."""

    incorrect_result_details: Dict[str, KnownIssueFields]
    """For each case where the USS provides an InjectFlightResponse `result` value that is not in the acceptable results, this field contains information about how the Issue should be described"""


class InjectionTarget(ImplicitDict):
    """The means to identify a particular USS within an AutomatedTest"""
    uss_role: str
    """The role of the USS that is the target of a flight injection attempt (e.g., 'Querying USS').  The test executor will assign a USS from the pool of USSs to be tested to each role defined in an AutomatedTest before executing that AutomatedTest."""


class FlightInjectionAttempt(ImplicitDict):
    """All information necessary to attempt to create a flight in a USS and to evaluate the outcome of that attempt"""
    reference_time: str
    """A time to hold when the test data was created, it will be updated to reflect when the test data was submitted """
    name: str
    """Name of this flight, used to refer to the flight later in the automated test"""

    test_injection: InjectFlightRequest
    """Definition of the flight to be injected"""

    known_responses: KnownResponses
    """Details about what the USS under test should report after processing the test data"""

    injection_target: InjectionTarget
    """The particular USS to which the flight injection attempt should be directed"""


class FlightDeletionAttempt(ImplicitDict):
    """All information necessary to attempt to close a flight previously injected into a USS"""
    flight_name: str
    """Name of the flight previously injected into the USS to delete"""


class TestStep(ImplicitDict):
    """The action taken in one step of a sequence of steps constituting an automated test"""
    name: str
    """Human-readable name/summary of this step"""

    inject_flight: Optional[FlightInjectionAttempt]
    """If populated, the test driver should attempt to inject a flight for this step"""

    delete_flight: Optional[FlightDeletionAttempt]
    """If populated, the test driver should attempt to delete the specified flight for this step"""


class AutomatedTest(ImplicitDict):
    """Definition of a complete automated test involving some subset of USSs under test"""
    name: str
    """Human-readable name of this test (e.g., 'Nominal planning')"""

    steps: List[TestStep]
    """Actions to be performed for this test"""
