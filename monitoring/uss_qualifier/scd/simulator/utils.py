from typing import List, Union
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
from pathlib import Path
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest

class GeometryGenerationRule(ImplicitDict):
    """ A class to hold configuration for developing flight paths for testing """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    ''' An object to hold generated flight path and the associated rule '''
    geometry: Union[LineString, Polygon]    
    geometry_generation_rule: GeometryGenerationRule
  
class RequiredResults(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    expected_response: str 
    authorisation_data_fields_to_evaluate: List[str]
    operational_intent_processing_result: str

class TestInjectionRequiredResult(ImplicitDict):
    test_injection: InjectFlightRequest
    required_result: RequiredResults