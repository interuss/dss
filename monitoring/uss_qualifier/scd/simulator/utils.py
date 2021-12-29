from typing import List, Union, Literal
from geojson import Feature
import enum
from shapely.geometry import LineString, Polygon
from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime
from monitoring.monitorlib.scd import Altitude
from pathlib import Path

    
class InjectFlightResponse(ImplicitDict):
    ''' A class to hold test flight submission response '''
    result: str
    operational_intent_id: str