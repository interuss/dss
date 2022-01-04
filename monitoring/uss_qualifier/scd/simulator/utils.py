from typing import List, Union
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
from pathlib import Path
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest

class GeometryGenerationRule(ImplicitDict):
    """ A class to hold configuration for developing treatment flight path """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    ''' An object to hold generated flight path, is_control is a nomenclature used to see if the generated path is the first one '''
    geometry: Union[LineString, Polygon]
    is_control: bool
    geometry_generation_rule: GeometryGenerationRule
  
class RequiredResults(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    expected_response: str # One of "Planned", "Rejected", "ConflictWithFlight" or "Failed"
    authorisation_data_validation_checks: List[str]
    operational_intent_validation_checks: List[str]

class TestInjectionRequiredResult(ImplicitDict):
    test_injection: InjectFlightRequest
    required_result: RequiredResults