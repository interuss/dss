from pathlib import Path
from typing import List, Union, Literal
from geojson import Feature
import enum
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime
from monitoring.monitorlib.scd import Altitude, Volume4D,  OperationalIntentState
from pathlib import Path

class OutputSubDirectories(ImplicitDict):
    ''' A class to hold information about output directories that will be generated when writing the files to disk. '''
    injection_payloads: Path 
    scd_rules: Path

class InjectionStatus(str, enum.Enum):
    ''' A enum to hold results of flight processing as defined by the SCD test API '''
    Planned = 'Planned'
    Rejected = 'Rejected'
    ConflictWithFlight = 'ConflictWithFlight'
    Failed = 'Failed'

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
    
class OperationalIntentInjection(ImplicitDict):
    ''' A class to hold data for operational intent data that will be submitted to the SCD testing interface.  For more information see : https://github.com/interuss/dss/blob/master/interfaces/automated-testing/scd/scd.yaml#L286'''
    state: Literal[OperationalIntentState.Accepted,OperationalIntentState.Activated,OperationalIntentState.Nonconforming, OperationalIntentState.Contingent] 
    priorty: int = 0
    volumes: List[Volume4D]
    off_nominal_volumes: List[Volume4D]= []


## Definitions around flight authorization data that needs to be submitted to the test injection interface
    
class OperationMode(str, enum.Enum):
    ''' A enum to hold all modes for an operation, for more information see: https://github.com/interuss/dss/blob/master/interfaces/automated-testing/scd/scd.yaml#L393 '''
    Undeclared = 'Undeclared'
    Vlos = 'Vlos'
    Bvlos = 'Bvlos'

class OperationCategory(str, enum.Enum):
    ''' A enum to hold all categories for an operation '''
    Open = 'Open'
    Specific = 'Specific'
    Certified = 'Certified'

class UASClass(str, enum.Enum):
    ''' A object to hold all UAS Classes per EASA defintions'''
    Other = 'Other'
    C0 = 'C0'
    C1 = 'C1'
    C2 = 'C2'
    C3 = 'C3'
    C4 = 'C4'
    C5 = 'C5'
    C6 = 'C6'

class IDTechnology(str, enum.Enum):
    ''' A enum to hold ID technologies for an operation '''
    Network = 'network'
    Broadcast = 'broadcast'

class FlightAuthorizationData(ImplicitDict):
    '''A class to hold information about Flight Authorization Test, for more information see https://github.com/interuss/dss/blob/master/interfaces/automated-testing/scd/scd.yaml#L317'''
    
    uas_serial_number: str
    operation_mode: Literal[OperationMode.Undeclared, OperationMode.Vlos, OperationMode.Bvlos]
    operation_category: Literal[OperationCategory.Open, OperationCategory.Specific, OperationCategory.Certified]
    uas_class: Literal[UASClass.Other,UASClass.C0, UASClass.C1, UASClass.C2, UASClass.C3, UASClass.C4,UASClass.C5, UASClass.C6]
    identification_technologies: List[str]
    uas_type_certificate: str
    connectivity_methods: List[str]
    endurance_minutes: int
    emergency_procedure_url: str
    operator_id: str
    uas_id: str    

class InjectionResults(ImplicitDict):
    result: Literal[InjectionStatus.Planned, InjectionStatus.Rejected, InjectionStatus.ConflictWithFlight, InjectionStatus]
    operational_intent_id: str