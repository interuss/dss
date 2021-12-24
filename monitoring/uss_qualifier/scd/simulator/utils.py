from typing import List, Union, Literal
from geojson import Feature
import enum
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime
from monitoring.monitorlib.scd import Altitude
from monitoring.monitorlib.scd_automated_testing.scd_injection_observation_api import InjectionStatus
from pathlib import Path

## Definitions around operational intent data that needs to be submitted to the test injection interface

class GeometryGenerationRule(ImplicitDict):
    """ A class to hold configuration for generation of geometries """
    intersect_space:bool = 0 # This flag specifies if the genrated geometry should interect with the first one (used only if )
    
class VolumeGenerationRule(ImplicitDict):
    """ A class to hold configuration rules for flight volume generation """
    intersect_altitude: bool = 0 # The simulator has the ability to generate multiple volumes, this is a flag to ensure that a intersection in altitude must be ensured
    intersect_time: bool = 0 # The simulator has the ability to generate multiple volumes, this is a flag to ensure that a intersection of flight paths in time is ensured 
    is_control: bool = 0 # When multiple volumes are generated the first one is called control against which the intersection rules are applied. 
    expected_result: InjectionStatus.ConflictWithFlight # If the flag for intersection is set then the 
    
class GeneratedGeometry(ImplicitDict):
    ''' An object to hold generated flight path, is_control is a nomenclature used to see if the generated path is the first one (against which the geometry and vole generation rules are applied) '''
    geometry: Union[LineString, Polygon]
    is_control: bool
    geometry_generation_rule: GeometryGenerationRule
  
class GeneratedFlightPlan(ImplicitDict):
    ''' A class used internally in the simulator with altiude and time details and GeoJSON as a flight plan the GeoJSON flight plan is converted in to a Volume 4D object and written to disk subsequently '''
    flight_plan: Feature 
    time_start: StringBasedDateTime
    time_end: StringBasedDateTime
    altitude_lower: Altitude
    altitude_upper: Altitude
    
class InjectFlightResponse(ImplicitDict):
    ''' A class to hold test status response '''
    result: Literal[InjectionStatus.Planned, InjectionStatus.Rejected, InjectionStatus.ConflictWithFlight, InjectionStatus.Failed]
    operational_intent_id: str