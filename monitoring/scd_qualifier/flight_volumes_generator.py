from pathlib import Path
import arrow
from typing import List
from shapely.geometry import LineString 
from pyproj import Geod
from monitoring.monitorlib import scd
from monitoring.scd_qualifier.utils import TreatmentVolumeOptions, TreatmentPathOptions, Volume3D, Volume4D
import shapely.geometry
from shapely.geometry import asShape
import pathlib, os
import geojson
import random


class FlightVolumeGenerator():
    ''' A class to generate flight volumes in the ASTM Volume 4D specification. As a input the module takes in a bounding box for which to generate the volumes within. Further test'''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl = 50.0
        self.control_flight_path: LineString
        self.geod = Geod(ellps="WGS84")
        self.input_extents_valid()

    def input_extents_valid(self) -> None:
        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.'''

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure enough space for tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError("The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square")
    def generate_random_flight_path(self) -> LineString:
        '''Generate a random flight path '''
        
        random_flight_path = geojson.utils.generate_random(featureType = "LineString", numberVertices=2,
                    boundingBox=[self.minx, self.miny, self.maxx, self.maxy])
        
        return random_flight_path
        

    def generate_flight_path(self, path_options:TreatmentPathOptions, is_control:bool= False) -> LineString: 
        ''' A method to generates flight path within a geographic bounds. This method utilzies the generate_random utiltiy provided by the geojson module to generate flight paths. '''
        if is_control:
            flight_path = self.generate_random_flight_path()
            self.control_flight_path = flight_path
        else: 
            if path_options.intersect_space: 
                path_intersects = False
                while (path_intersects==False):
                    # We are trying to generate a path that intersects with the control, we keep generating paths till one is found that does intersect
                    flight_path = self.generate_random_flight_path()
                    line_shape = asShape(flight_path)
                    control_line_shape = asShape(self.control_flight_path)
                    intersects = control_line_shape.intersects(line_shape)
                    path_intersects = intersects    
            else: 
                flight_path = self.generate_random_flight_path()
                
        return flight_path

    def path_to_outline_polygon_converter(self, flight_path:LineString) -> Volume3D:
        ''' A method to convert a GeoJSON LineString to a ASTM outline_polygon object by buffering 10m spatially '''
        volume3D = None
        return volume3D
        
    def generate_treatment_volume(self, treatment_options: TreatmentVolumeOptions) -> Volume4D:
        ''' Generate a volume given the options to the control volume '''
        pass
    
    def generate_test_payload(self, number_of_volumes:int = 6) -> List[Volume4D]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        all_payloads = []
        
        for volume_idx in range(0, number_of_volumes):
            is_control = 1 if (volume_idx == 0) else 0
            treatment_path_options = TreatmentPathOptions()
            if not is_control:
                should_intersect = random.randint(0, 1)
                treatment_path_options.intersect_space = should_intersect
            
            current_path = self.generate_flight_path(path_options = treatment_path_options, is_control= is_control)
            print(current_path) 
            
    
        return all_payloads


if __name__ == '__main__':
    COUNTRY_CODE = 'che'
    # Generate volumes 
    my_volume_generator = FlightVolumeGenerator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')
    
    # Change directory to write test_definitions folder is created in the rid_qualifier folder.
    p = pathlib.Path(__file__).parent.absolute()
    os.chdir(p)
    test_payload = my_volume_generator.generate_test_payload(number_of_volumes=3)