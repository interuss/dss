from typing import Union
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
import random
from itertools import cycle
import string

class GeometryGenerationRule(ImplicitDict):
    """A class to hold configuration for developing flight paths for testing """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    """An object to hold generated flight path and the associated rule """
    geometry: Union[LineString, Polygon]    
    geometry_generation_rule: GeometryGenerationRule
   
class FlightNameIncorrectField(ImplicitDict):
    """A class to hold name of a flight and the associated incorrect field within it, used when generating data for flight authorisation data checks """
    flight_name: str
    incorrect_field:str = None