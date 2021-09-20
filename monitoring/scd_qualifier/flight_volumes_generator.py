from pathlib import Path
import arrow
from typing import List
from shapely.geometry.linestring import LineString
from pyproj import Geod
from monitoring.monitorlib import scd
from .utils import Volume3D, Volume4D
import shapely.geometry
import pathlib, os
import geojson
from geojson import LineString

class FlightVolumeGenerator():
    ''' A class to generate flight volumes in the ASTM Volume 4D specification. As a input the module takes in a bounding box for which to generate the volumes within. Further test'''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl = 50.0

        self.geod = Geod(ellps="WGS84")
        self.input_extents_valid()

    def input_extents_valid(self) -> None:
        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.'''

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure a 50 m diameter tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError("The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square")


    def generate_flight_path(self) -> LineString: 
        ''' A method to generate flight path within a geographic bounds. This method utilzies the generate_random utiltiy provided by the geojson module to generate flight paths. '''
        pass

    def path_to_volume_converter(flight_path:LineString) -> Volume3D:
        volume3D = None
        return volume3D
        
    def generate_volume_payloads(number_of_volumes:int = 6) -> List[Volume4D]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        pass


if __name__ == '__main__':
    COUNTRY_CODE = 'che'
    # Generate volumes 
    my_volume_generator = FlightVolumeGenerator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')
    
    # Change directory to write test_definitions folder is created in the rid_qualifier folder.
    p = pathlib.Path(__file__).parent.absolute()
    os.chdir(p)
    pass
