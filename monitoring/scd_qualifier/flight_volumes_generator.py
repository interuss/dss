from pathlib import Path
import arrow
from typing import List
from shapely.geometry import LineString, Polygon
from pyproj import Geod, Proj
from monitoring.monitorlib import scd
from monitoring.scd_qualifier.utils import Altitude, VolumePolygon, VolumeGenerationRules, PathGenerationRules, Volume3D, Volume4D, PathPayload, LatLngPoint, Time
import shapely.geometry
from shapely.geometry import asShape
import pathlib, os
import geojson
import json
import random

class FlightVolumeGenerator():
    ''' A class to generate flight volumes in the ASTM Volume 4D specification. As a input the module takes in a bounding box for which to generate the volumes within. Further test'''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl:float = 50.0
        self.altitude_envelope: int = 15 # the buffer in meters for converting from flight path to a volume
        self.control_flight_path: LineString # the initial flight path against which subsequent flight paths are generated
                
        
        self.raw_paths: List[PathPayload]        
        self.now = arrow.now()        
        self.geod = Geod(ellps="WGS84")
        self.input_extents_valid()

    def utm_converter(self, shapely_shape: shapely.geometry, inverse:bool=False) -> shapely.geometry.shape:
        ''' A helper function to convert from lat / lon to UTM coordinates for buffering. tracks. This is the UTM projection (https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system), we use Zone 33T which encompasses Switzerland, this zone has to be set for each locale / city. Adapted from https://gis.stackexchange.com/questions/325926/buffering-geometry-with-points-in-wgs84-using-shapely '''

        proj = Proj(proj="utm", zone=self.utm_zone, ellps="WGS84", datum="WGS84")

        geo_interface = shapely_shape.__geo_interface__
        feature_type = geo_interface['type']
        coordinates = geo_interface['coordinates']
        if feature_type == 'Polygon':
            new_coordinates = [[proj(*point, inverse=inverse) for point in linring] for linring in coordinates]
        elif feature_type == 'LineString':
            new_coordinates = [proj(*point, inverse=inverse) for point in coordinates]
        else:
            raise RuntimeError('Unexpected geo_interface type: {}'.format(feature_type))

        return shapely.geometry.shape({'type': feature_type, 'coordinates': tuple(new_coordinates)})

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
        

    def generate_single_flight_path(self, path_options:PathGenerationRules, is_control:bool= False) -> LineString: 
        ''' A method to generates flight path within a geographic bounds. This method utilzies the generate_random utiltiy provided by the geojson module to generate flight paths. '''
        if is_control:
            flight_path = self.generate_random_flight_path()
            flight_path_shp = asShape(flight_path)
            flight_path_utm = self.utm_converter(flight_path_shp)
            buffer_shape_utm = flight_path_utm.buffer(15)
            buffered_shape = self.utm_converter(buffer_shape_utm, inverse=True)
            
            self.control_flight_path = flight_path
        else: 
            if path_options.intersect_space: 
                path_intersects = False
                while (path_intersects==False):
                    # We are trying to generate a path that intersects with the control, we keep generating paths till one is found that does intersect
                    flight_path = self.generate_random_flight_path()
                    line_shape = asShape(flight_path)
                    intersects = self.flight_path.intersects(line_shape)
                    path_intersects = intersects    
            else: 
                flight_path = self.generate_random_flight_path()
                
        return flight_path

    def convert_path_to_volume(self, flight_path:LineString, volume_generation_options: VolumeGenerationRules, altitude_of_ground_level_wgs_84:int) -> Volume3D:
        ''' A method to convert a GeoJSON LineString to a ASTM outline_polygon object by buffering 15m spatially '''
        
        flight_path_shp = asShape(flight_path)
        flight_path_utm = self.utm_converter(flight_path_shp)
        buffer_shape_utm = flight_path_utm.buffer(15)
        
        if volume_generation_options.intersect_altitude == True:    
            alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_envelope  
            alt_lower = altitude_of_ground_level_wgs_84 - self.altitude_envelope
        else:
            # Raise the altitude by 50m so that they dont intersect in altitude
            alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_envelope  + 50
            alt_lower = altitude_of_ground_level_wgs_84 - self.altitude_envelope + 50
        
        buffered_shape_geo = self.utm_converter(buffer_shape_utm, inverse=True)
        
        all_vertices = []
        altitude_upper = Altitude(value= alt_upper, reference = "W84", units="M")
        altitude_lower = Altitude(value=alt_lower, reference = "W84", units="M")
        for vertex in list(buffered_shape_geo.exterior.coords):
            coord = LatLngPoint(lat = vertex[0] , lng = vertex[1])
            all_vertices.append(coord)
        p = VolumePolygon(vertices=all_vertices)
        
        volume3D = Volume3D(outline_polygon = p, altitude_lower=altitude_lower, altitude_upper=altitude_upper)
        
        return volume3D

    def transform_3d_volume_to_4d(self, volume_3d : Volume3D,volume_generation_options: VolumeGenerationRules) -> Volume4D:
        if volume_generation_options.intersect_time: 
            # Overlap with the control 
            three_mins_from_now = self.now.shift(minutes = 3)
            eight_mins_from_now = self.now.shift(minutes = 8)
            start_time = Time(value = three_mins_from_now.isoformat(), format = "RFC3339")
            end_time = Time(value = eight_mins_from_now.isoformat(), format = "RFC3339")

        else: 
            mins = [10,15,20,25,30,25,40,45,50,55,60,65,70]
            future_minutes = random.choice(mins)
            future_start = self.now.shift(minutes = future_minutes)
            future_end = self.now.shift(minutes = (future_minutes+ 4))
            start_time = Time(value = future_start.isoformat(), format = "RFC3339")
            end_time = Time(value = future_end.isoformat(), format = "RFC3339")

    
        volume_4D = Volume4D(volume=volume_3d, time_start= start_time, time_end=end_time)
        
        return volume_4D
    
    def generate_raw_paths(self, number_of_paths:int = 6) -> List[PathPayload]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        
        raw_paths: List[PathPayload] = []
        for volume_idx in range(0, number_of_paths):
            is_control = True if (volume_idx == 0) else False
            path_generation_options = PathGenerationRules()
            
            if (volume_idx == (number_of_paths -1)):
                should_intersect = False
                path_generation_options.intersect_space = should_intersect
            else:
                should_intersect = True
            
            current_path = self.generate_single_flight_path(path_options = path_generation_options, is_control= is_control)
            raw_path = PathPayload(path = current_path, path_options = path_generation_options, is_control = is_control)
            # the first path is control, the last path does not intersect the control
            raw_paths.append(raw_path)
        return raw_paths

    def generate_path_parameters(self, raw_paths:List[PathPayload]) -> List[VolumeGenerationRules]: 
        ''' A method to generate rules for generation of new paths '''
        
        all_volume_rules: List[VolumeGenerationRules] = []
        last_path_index = len(raw_paths) - 1 
        for path_index, raw_path in enumerate(raw_paths): 
            if path_index in [0,last_path_index]:
                # This the control path or the well clear path no need to have any time / atltitude interserction
                volume_generation_options = VolumeGenerationRules(intersect_altitude= False, intersect_time= False, expected_result = 'pass')
                if path_index == 0:
                    volume_generation_options.is_control = True
            else:
                # intersect in time and / or intersect in altitude 
                volume_generation_options = random.choice([VolumeGenerationRules(intersect_altitude=True, intersect_time= False, expected_result = 'pass'),VolumeGenerationRules(intersect_altitude=True, intersect_time= True, expected_result = 'fail'),VolumeGenerationRules(intersect_altitude=False, intersect_time= True, expected_result = 'pass')])
            all_volume_rules.append(volume_generation_options)
        return all_volume_rules

    def generate_4d_volumes(self,raw_paths:List[PathPayload],rules : List[PathPayload], altitude_of_ground_level_wgs_84 :int) -> List[Volume4D]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        all_volume_4d :List[Volume4D] = []
        last_path_index = len(raw_paths) - 1 
        for path_index, raw_path in enumerate(raw_paths): 
            volume_generation_options = rules[path_index]
            
            flight_volume_3d = self.convert_path_to_volume(flight_path = raw_path.path, volume_generation_options = volume_generation_options, altitude_of_ground_level_wgs_84=altitude_of_ground_level_wgs_84)
            flight_volume_4d = self.transform_3d_volume_to_4d(volume_3d= flight_volume_3d, volume_generation_options =  volume_generation_options)
            all_volume_4d.append(flight_volume_4d)
            
        return all_volume_4d


class SCDFlightPathVolumeWriter():
    """ A class to write raw Flight Paths and volumes to disk so that they can be examined / used in other software """

    def __init__(self, raw_paths:List[PathPayload],  serialized_flight_volumes: List[Volume4D], all_rules: List[VolumeGenerationRules],country_code='che')-> None:
        
        self.country_code = country_code
        self.output_directory = Path('test_definitions', self.country_code)        
        # Create test_definition directory if it does not exist        
        self.output_directory.mkdir(parents=True, exist_ok=True)
        self.output_subdirectories = (Path(self.output_directory, 'flight_volumes'), Path(self.output_directory, 'path_geojson'),Path(self.output_directory, 'scd_rules'),)
        for output_subdirectory in self.output_subdirectories:
            output_subdirectory.mkdir(parents=True, exist_ok=True)
        self.raw_paths = raw_paths
        self.serialized_flight_volumes = serialized_flight_volumes
        self.all_rules = all_rules

    def write_flight_paths(self) -> None:
        ''' A method to write flight paths to disk as GeoJSON features '''
        
        for path_id, path_payload in enumerate(self.raw_paths):
            feature_collection = {"type": "FeatureCollection", "features": []}                       
            line_feature = {'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(path_payload.path)}
            feature_collection['features'].append(line_feature)
            path_file_name = 'path_%s.geojson' % str(path_id + 1)  # Avoid Zero based numbering
            tracks_file_path = self.output_subdirectories[1] / path_file_name
            with open(tracks_file_path, 'w') as f:
                f.write(json.dumps(feature_collection))

    
    def write_flight_volumes(self) -> None:
        ''' A method to write volume 4D objects to disk '''

        for volume_id, volume in enumerate(self.serialized_flight_volumes):                     
            volume_file_name = 'volume_%s.json' % str(volume_id + 1)  # Avoid Zero based numbering           
            volume_file_path = self.output_subdirectories[0] / volume_file_name
            with open(volume_file_path, 'w') as f:
                f.write(json.dumps(volume))

    def write_test_rules(self) -> None:
        ''' A method to write test parameters to see expected outputs of a test'''
        
        volume_file_name = 'volume_rules.json'
        rule_file_path = self.output_subdirectories[2] / volume_file_name
        with open(rule_file_path, 'w') as f:
            f.write(json.dumps(self.all_rules))
    
if __name__ == '__main__':
    COUNTRY_CODE = 'che'
    # Generate volumes 
    my_volume_generator = FlightVolumeGenerator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')
    altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit
    # Change directory to write test_definitions folder is created in the rid_qualifier folder.
    p = pathlib.Path(__file__).parent.absolute()
    os.chdir(p)
    flight_paths = my_volume_generator.generate_raw_paths(number_of_paths=6)
    all_rules = my_volume_generator.generate_path_parameters(raw_paths=flight_paths)
    
    flight_volumes = my_volume_generator.generate_4d_volumes(raw_paths = flight_paths, rules = all_rules, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
   
    my_flight_path_writer = SCDFlightPathVolumeWriter(raw_paths= flight_paths, serialized_flight_volumes=flight_volumes, all_rules = all_rules)

    my_flight_path_writer.write_flight_paths()    
    my_flight_path_writer.write_flight_volumes()
    my_flight_path_writer.write_test_rules()
