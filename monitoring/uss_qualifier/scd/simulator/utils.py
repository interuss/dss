from pathlib import Path
from typing import List, Union, Literal
from geojson import Feature
import enum
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime
from monitoring.monitorlib.scd import Altitude, VolumePolygon, Volume4D, Volume3D
from pathlib import Path

class OutputSubDirectories(ImplicitDict):
    ''' A class to hold information about output directories that will be generated when writing the files to disk. '''
    partial_operational_intent_references: Path
    scd_rules: Path

class OperationalIntentReference(ImplicitDict):
    """Class for keeping track of an operational intent reference"""
    id: str

class Volume3D(ImplicitDict):
    '''A class to hold Volume3D objects '''
    
    outline_polygon: VolumePolygon
    altitude_lower: Altitude
    altitude_upper: Altitude

class Volume4D(ImplicitDict):
    '''A class to hold ASTM Volume4D objects '''
    volume: Volume3D
    time_start: StringBasedDateTime
    time_end: StringBasedDateTime

class OperationalIntentDetails(ImplicitDict):
    """Class for holding details of an operational intent reference """
    volumes: List[Volume4D]
    priority: int

class VolumeGenerationRule(ImplicitDict):
    """ A class to hold configuration for holding rule for volume generation """
    intersect_altitude: bool = 0
    intersect_time: bool = 0
    is_control: bool = 0
    expected_result: str
    
class SCDVolume4D(ImplicitDict):
    ''' A class to hold a volume 4D with path and time details and GeoJSON flight plan, USSP will have to transform flight plan to a  '''
    flight_plan: Feature
    time_start: StringBasedDateTime
    time_end: StringBasedDateTime
    altitude_lower: Altitude
    altitude_upper: Altitude
    
class GeometryGenerationRule(ImplicitDict):
    """ A class to hold configuration for developing treatment flight path """
    intersect_space:bool = 0
    
class GeneratedGeometry(ImplicitDict):
    ''' An object to hold generated flight path, is_control is a nomenclature used to see if the generated path is the first one '''
    geometry: Union[LineString, Polygon]
    is_control: bool
    geometry_generation_rule: GeometryGenerationRule
  
class FlightAuthPartialRequestDetails(ImplicitDict):
    ''' An object to hold flight authorization details '''
    uas_serial_number:str
    operation_mode: str
    operation_category:str
    uas_class:str
    identification_technologies:List = []
    connectivity_methods:List = []
    endurance_minutes: int 
    emergency_procedure_url: str
    operator_id:str

class OperationalIntentState(str, enum.Enum):
    ''' A enum to hold all states of a negotiation proposal state '''
    Accepted = 'Accepted'
    Activated = 'Activated'
    Nonconforming = 'Nonconforming'
    Contingent = 'Contingent'

class ImplicitSubscriptionParameters(ImplicitDict):
    uss_base_url: str
    notify_for_constraints: bool

class PartialOperationalIntentReferenceDetails(ImplicitDict):
    ''' A class to hold data about operational intent creation that the USS submits to the DSS '''
    extents: List[Volume4D]
    key: List[str]
    state: Literal[OperationalIntentState.Accepted,OperationalIntentState.Activated,OperationalIntentState.Nonconforming, OperationalIntentState.Contingent]
    uss_base_url: str
    # the following two fields are not used beccuase they are specific to the operator but are a part of the Opreational Intent Reference creation details.
    # subsription_id: str
    # new_subscription: ImplicitSubscriptionParameters