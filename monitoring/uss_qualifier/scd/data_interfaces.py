from typing import List
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest



class RequiredResults(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    expected_response: str 
    authorisation_data_fields_to_evaluate: List[str]  # In the 
    operational_intent_processing_result: str

class TestInjectionRequiredResult(ImplicitDict):
    test_injection: InjectFlightRequest
    required_result: RequiredResults