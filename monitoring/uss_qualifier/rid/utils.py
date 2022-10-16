from typing import List, NamedTuple
from shapely.geometry import Polygon
import shapely.geometry
from datetime import datetime


class QueryBoundingBox(NamedTuple):
    """This is the object that stores details of query bounding box"""

    name: str
    shape: Polygon
    timestamp_before: datetime
    timestamp_after: datetime


class FlightPoint(NamedTuple):
    """This object holds basic information about a point on the flight track, it has latitude, longitude and altitude in WGS 1984 datum"""

    lat: float  # Degrees of latitude north of the equator, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1160
    lng: float  # Degrees of longitude east of the Prime Meridian, with reference to the WGS84 ellipsoid. For more information see: https://github.com/uastech/standards/blob/master/remoteid/canonical.yaml#L1170
    alt: float  # meters in WGS 84, normally calculated as height of ground level in WGS84 and altitude above ground level
    speed: float  # speed in m / s
    bearing: float  # forward azimuth for the this and the next point on the track


class GridCellFlight(NamedTuple):
    """A object to hold details of a grid location and the track within it"""

    bounds: shapely.geometry.polygon.Polygon
    track: List[FlightPoint]
