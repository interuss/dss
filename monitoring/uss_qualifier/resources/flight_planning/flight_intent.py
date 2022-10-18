from typing import List, Optional

from implicitdict import ImplicitDict, StringBasedDateTime, StringBasedTimeDelta

from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
)


class FlightIntent(ImplicitDict):
    reference_time: StringBasedDateTime
    """The time that all other times in the FlightInjectionAttempt are relative to. If this FlightInjectionAttempt is initiated by uss_qualifier at t_test, then each t_volume_original timestamp within test_injection should be adjusted to t_volume_adjusted such that t_volume_adjusted = t_test + planning_time when t_volume_original = reference_time"""

    request: InjectFlightRequest
    """Definition of the flight the user wants to create."""


class FlightIntentCollection(ImplicitDict):
    intents: List[FlightIntent]
    """Flights that users want to create."""


class FlightIntentsJSONFileConfiguration(ImplicitDict):
    path: str
    """Path to a file containing a JSON representation of an instance of FlightRecordCollection.  This may be a local file or a web URL."""


class FlightIntentsSpecification(ImplicitDict):
    planning_time: StringBasedTimeDelta
    """Time delta between the time uss_qualifier initiates this FlightInjectionAttempt and when a timestamp within the test_injection equal to reference_time occurs"""

    json_file_source: Optional[FlightIntentsJSONFileConfiguration] = None
    """When this field is populated, flight intents will be loaded from a JSON file containing an object with the FlightIntentCollection schema"""
