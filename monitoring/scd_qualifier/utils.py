from typing import List, Union
from geojson import Feature
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict
import arrow

class StringBasedDateTime(str):
  """String that only allows values which describe a datetime."""
  def __new__(cls, value):
    if isinstance(value, str):
      t = arrow.get(value).datetime
    else:
      t = value
    str_value = str.__new__(cls, arrow.get(t).to('UTC').format('YYYY-MM-DDTHH:mm:ss.SSSSSS') + 'Z')
    str_value.datetime = t
    return str_value

class LatLngPoint(ImplicitDict):
    '''A class to hold information about a location as Latitude / Longitude pair'''
    lat: float
    lng: float

class Radius(ImplicitDict):
    ''' A class to hold the radius of a circle for the outline_circle object as specified per the ASTM standard '''
    value: float
    units:str

class VolumePolygon(ImplicitDict):
    ''' A class to hold the polygon object, used in the outline_polygon of the Volume3D object as specified in the ASTM standard '''
    vertices: List[LatLngPoint] # A minimum of three LatLngPoints

class Circle(ImplicitDict):
    ''' Hold the details of a circle object used in the outline_circle object as specified in the ASTM standard '''
    center: LatLngPoint 
    radius: Radius


class Altitude(ImplicitDict):
    ''' A class to hold altitude per ASTM standard '''
    value:int
    reference:str 
    units: str 

class Time(ImplicitDict):
    ''' A class to hold Time per ASTM standard'''
    value: str
    format:str 

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

class VolumeGenerationRules(ImplicitDict):
    """ A class to hold configuration for developing rules """
    intersect_altitude: bool = 0
    intersect_time: bool = 0
    is_control: bool = 0
    expected_result: str

class GeoJSONFeature(ImplicitDict):
    ''' A class to hold GeoJSON Feature flight plan, at this moment we are using a LineString but it could be a LineString or Polygon '''
    geojson: Feature
    
class SCDVolume4D(ImplicitDict):
    ''' A class to hold a volume 4D with path and time details and GeoJSON flight plan, USSP will have to transform flight plan to a  '''
    flight_plan: GeoJSONFeature
    time_start: StringBasedDateTime
    time_end: StringBasedDateTime
    altitude_lower: Altitude
    altitude_upper: Altitude
    
class GeometryGenerationRules(ImplicitDict):
    """ A class to hold configuration for developing treatment flight path """
    intersect_space:bool = 0
    
class GeometryPayload(ImplicitDict):
    ''' An object to hold generated flight path, is_control is a nomenclature used to see if the generated path is the first one '''
    geometry: Union[LineString, Polygon]
    is_control: bool
    geometry_generation_rules: GeometryGenerationRules
  
