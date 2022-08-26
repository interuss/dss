from monitoring.monitorlib.rid_automated_testing import injection_api
from monitoring.monitorlib.typing import ImplicitDict


# --- RID observation ---
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


# --- RID injection ---
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
