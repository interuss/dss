from typing import Union
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
from pathlib import Path

class GeometryGenerationRule(ImplicitDict):
    """ A class to hold configuration for developing treatment flight path """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    ''' An object to hold generated flight path, is_control is a nomenclature used to see if the generated path is the first one '''
    geometry: Union[LineString, Polygon]
    is_control: bool
    geometry_generation_rule: GeometryGenerationRule
  