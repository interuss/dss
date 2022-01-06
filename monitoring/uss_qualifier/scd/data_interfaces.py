from typing import List
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest

class RequiredResults(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    expected_result: str 
    # Why we expected a response of rejected (add a dict / enum)

class TestInjectionRequiredResult(ImplicitDict):
    test_injection: InjectFlightRequest
    required_result: RequiredResults