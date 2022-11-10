from enum import Enum
from uas_standards.ansi_cta_2063_a import SerialNumber
from uas_standards.en4709_02 import OperatorRegistrationNumber
from implicitdict import ImplicitDict, StringBasedDateTime
from typing import List, Optional
from monitoring.monitorlib.scd import Volume4D

SCOPE_SCD_QUALIFIER_INJECT = "utm.inject_test_data"

## Definitions around operational intent data that need to be submitted to the test injection interface


class OperationalIntentTestInjection(ImplicitDict):
    """A class to hold data for operational intent data that will be submitted to the SCD testing interface."""

    state: str
    priority: int = 0
    volumes: List[Volume4D]
    off_nominal_volumes: List[Volume4D] = []


### End of definitions around operational intent data

## Definitions around flight authorisation data that need to be submitted to the test injection interface


class UASClass(str, Enum):
    Other = "Other"
    C0 = "C0"
    C1 = "C1"
    C2 = "C2"
    C3 = "C3"
    C4 = "C4"
    C5 = "C5"
    C6 = "C6"


class OperationMode(str, Enum):
    Undeclared = "Undeclared"
    Vlos = "Vlos"
    Bvlos = "Bvlos"


class OperationCategory(str, Enum):
    Unknown = "Unknown"
    Open = "Open"
    Specific = "Specific"
    Certified = "Certified"


class FlightAuthorisationData(ImplicitDict):
    """A class to hold information about Flight Authorisation Test"""

    uas_serial_number: SerialNumber
    operation_mode: OperationMode
    operation_category: Optional[OperationCategory]
    uas_class: UASClass
    identification_technologies: List[str]
    uas_type_certificate: Optional[str]
    connectivity_methods: List[str]
    endurance_minutes: int
    emergency_procedure_url: str
    operator_id: OperatorRegistrationNumber
    uas_id: Optional[str]


### End of definitions around flight authorisation data


class InjectFlightRequest(ImplicitDict):
    """A class to hold the details of a test injection payload"""

    operational_intent: OperationalIntentTestInjection
    flight_authorisation: FlightAuthorisationData


class InjectFlightResult(str, Enum):
    Planned = "Planned"
    Rejected = "Rejected"
    ConflictWithFlight = "ConflictWithFlight"
    Failed = "Failed"


class InjectFlightResponse(ImplicitDict):
    """A class to hold test flight submission response"""

    result: InjectFlightResult
    notes: Optional[str]
    operational_intent_id: Optional[str]


class DeleteFlightResult(str, Enum):
    Closed = "Closed"
    Failed = "Failed"


class DeleteFlightResponse(ImplicitDict):
    result: DeleteFlightResult
    notes: Optional[str]


class ClearAreaRequest(ImplicitDict):
    request_id: str
    extent: Volume4D


class ClearAreaOutcome(ImplicitDict):
    success: bool
    message: Optional[str]
    timestamp: StringBasedDateTime


class ClearAreaResponse(ImplicitDict):
    outcome: ClearAreaOutcome


class Capability(str, Enum):
    FlightAuthorisationValidation = "FlightAuthorisationValidation"
    BasicStrategicConflictDetection = "BasicStrategicConflictDetection"
    HighPriorityFlights = "HighPriorityFlights"


class CapabilitiesResponse(ImplicitDict):
    capabilities: Optional[List[Capability]]


class Status(str, Enum):
    Starting = "Starting"
    Ready = "Ready"


class StatusResponse(ImplicitDict):
    status: str
    version: str
