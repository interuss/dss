from enum import Enum

from monitoring.monitorlib.rid_automated_testing import injection_api
from implicitdict import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import \
    InjectFlightRequest
from uas_standards.interuss.automated_testing.flight_planning.v1.api import (
    ClearAreaRequest,
)


class RequestType(str, Enum):
    RID_Observation_GetDisplayData = "rid.observation.getDisplayData"
    RID_Observation_GetDetails = "rid.observation.getDetails"
    RID_Injection_CreateTest = "rid.injection.createTest"
    RID_Injection_DeleteTest = "rid.injection.deleteTest"
    SCD_GetStatus = "scd.getStatus"
    SCD_GetCapabilities = "scd.getCapabilities"
    SCD_PutFlight = "scd.putFlight"
    SCD_DeleteFlight = "scd.deleteFlight"
    SCD_CreateClearAreaRequest = "scd.createClearAreaRequest"


SCD_REQUESTS = {RequestType.SCD_GetStatus, RequestType.SCD_GetCapabilities, RequestType.SCD_PutFlight, RequestType.SCD_DeleteFlight, RequestType.SCD_CreateClearAreaRequest}


# Each request descriptor in this file is expected to implement a static
# request_type_name() method which indicates the type of request corresponding
# with the descriptor.  Handler clients will use this type name to determine
# what kind of query each query is.

# --- RID observation (interfaces/automated_testing/rid/observation.yaml) ---
class RIDObservationGetDisplayDataRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.RID_Observation_GetDisplayData

    view: str


class RIDObservationGetDetailsRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.RID_Observation_GetDetails

    id: str


# --- RID injection (interfaces/automated_testing/rid/injection.yaml) ---
class RIDInjectionCreateTestRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.RID_Injection_CreateTest

    test_id: str
    request_body: injection_api.CreateTestParameters


class RIDInjectionDeleteTestRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.RID_Injection_DeleteTest

    test_id: str
    version: str


# --- SCD injection (interfaces/automated_testing/scd/v1/scd.yaml) ---
class SCDInjectionStatusRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.SCD_GetStatus


class SCDInjectionCapabilitiesRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.SCD_GetCapabilities


class SCDInjectionPutFlightRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.SCD_PutFlight

    flight_id: str
    request_body: InjectFlightRequest


class SCDInjectionDeleteFlightRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> RequestType:
        return RequestType.SCD_DeleteFlight

    flight_id: str


class SCDInjectionClearAreaRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return RequestType.SCD_CreateClearAreaRequest

    request_body: ClearAreaRequest
