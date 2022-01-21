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
   
class OutputSubDirectories(ImplicitDict):
    ''' A class to hold information about output directories that will be generated when writing the files to disk. '''
    autmoated_test_base_path: Path
