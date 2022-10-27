from typing import List

from implicitdict import ImplicitDict, StringBasedDateTime, StringBasedTimeDelta

from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
)
from monitoring.uss_qualifier.fileio import FileReference


class FlightIntent(ImplicitDict):
    reference_time: StringBasedDateTime
    """The time that all other times in the FlightInjectionAttempt are relative to. If this FlightInjectionAttempt is initiated by uss_qualifier at t_test, then each t_volume_original timestamp within test_injection should be adjusted to t_volume_adjusted such that t_volume_adjusted = t_test + planning_time when t_volume_original = reference_time"""

    request: InjectFlightRequest
    """Definition of the flight the user wants to create."""


class FlightIntentCollection(ImplicitDict):
    intents: List[FlightIntent]
    """Flights that users want to create."""


class FlightIntentsSpecification(ImplicitDict):
    planning_time: StringBasedTimeDelta
    """Time delta between the time uss_qualifier initiates this FlightInjectionAttempt and when a timestamp within the test_injection equal to reference_time occurs"""

    file_source: FileReference
    """Location of file to load"""
