from monitoring.monitorlib.rid_automated_testing import injection_api
from implicitdict import ImplicitDict
from uas_standards.interuss.automated_testing.flight_planning.v1.api import \
    InjectFlightRequest, ClearAreaRequest


# Each request descriptor in this file is expected to implement a static
# request_type_name() method which indicates the type of request corresponding
# with the descriptor.  Handler clients will use this type name to determine
# what kind of query each query is.

# --- RID observation (interfaces/automated_testing/rid/observation.yaml) ---
class RIDObservationGetDisplayDataRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'rid.observation.getDisplayData'

    view: str


class RIDObservationGetDetailsRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'rid.observation.getDetails'

    id: str


# --- RID injection (interfaces/automated_testing/rid/injection.yaml) ---
class RIDInjectionCreateTestRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'rid.injection.createTest'

    test_id: str
    request_body: injection_api.CreateTestParameters


class RIDInjectionDeleteTestRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'rid.injection.deleteTest'

    test_id: str
    version: str


# --- SCD injection (interfaces/automated_testing/scd/v1/scd.yaml) ---
class SCDInjectionStatusRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'scd.getStatus'


class SCDInjectionCapabilitiesRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'scd.getCapabilities'


class SCDInjectionPutFlightRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'scd.putFlight'

    flight_id: str
    request_body: InjectFlightRequest


class SCDInjectionDeleteFlightRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'scd.deleteFlight'

    flight_id: str


class SCDInjectionClearAreaRequest(ImplicitDict):
    @staticmethod
    def request_type_name() -> str:
        return 'scd.createClearAreaRequest'

    request_body: ClearAreaRequest
