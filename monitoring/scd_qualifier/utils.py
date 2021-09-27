from typing import List, NamedTuple

from shapely.geometry.linestring import LineString
from monitoring.monitorlib.typing import ImplicitDict
import arrow
import uuid

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

class LatLngPoint(NamedTuple):
    '''A clas to hold information about a location as Latitude / Longitude pair'''
    lat: float
    lng: float

class Radius(NamedTuple):
    ''' A class to hold the radius of a circle for the outline_circle object '''
    value: float
    units:str

class Polygon(NamedTuple):
    ''' A class to hold the polygon object, used in the outline_polygon of the Volume3D object '''
    vertices: List[LatLngPoint] # A minimum of three LatLngPoints

class Circle(NamedTuple):
    ''' Hold the details of a circle object used in the outline_circle object'''
    center: LatLngPoint 
    radius: Radius


class Altitude(NamedTuple):
    ''' A class to hold altitude '''
    value:int
    reference:str
    units: str


class OperationalIntentReference(NamedTuple):
    """Class for keeping track of an operational intent reference"""
    id: str


class Volume3D(NamedTuple):
    '''A class to hold Volume3D objects'''
    
    outline_polygon: Polygon
    altitude_lower: Altitude
    altitude_upper: Altitude

class Volume4D(NamedTuple):
    '''A class to hold ASTM Volume4D objects '''
    volume: Volume3D
    time_start: StringBasedDateTime
    time_end: StringBasedDateTime


class OperationalIntentDetails(NamedTuple):
    """Class for keeping track of an operational intent reference """
    volumes: List[Volume4D]
    priority: int

class TreatmentVolumeOptions(NamedTuple):
    """ A class to hold configuration for developing treatments """
    intersect_altitude: bool = 0
    is_control: bool = 0
    
class TreatmentPathOptions(ImplicitDict):
    """ A class to hold configuration for developing treatment flight path as GeoJSON """
    intersect_space:bool = 0
    
class PathPayload(ImplicitDict):
    path: LineString
    is_control: bool
    path_options: TreatmentPathOptions
    
    