from pathlib import Path
import arrow
import random
from typing import List, Union
from shapely.geometry import LineString, Point
from pyproj import Geod, Proj
from shapely.geometry.polygon import Polygon
from monitoring.monitorlib import scd
from monitoring.scd_qualifier.utils import Altitude, VolumePolygon, VolumeGenerationRule, GeometryGenerationRule, Volume3D, Volume4D, GeometryPayload, LatLngPoint, Time, SCDVolume4D
import shapely.geometry
from shapely.geometry import asShape
import pathlib, os
import geojson
import json
import random

class ProximateFlightVolumeGenerator():
    ''' A class to generate flight volumes in the ASTM Volume 4D specification. As a input the module takes in a bounding box for which to generate the volumes within. Further test'''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        """ Create a ProximateVolumeGenerator within a given geographic bounding box. 

        Once these extents are specified, a grid will be created with two rows. A combination of LineStrings and Polygons will be generated withing these bounds. While linestrings can extend to the full boundaries of the box, polygon areas are generated within the grid. 

        Args:
        minx: Western edge of bounding box (degrees longitude)
        maxx: Eastern edge of bounding box (degrees longitude)
        miny: Southern edge of bounding box (degrees latitude)
        maxy: Northern edge of bounding box (degrees latitude)
        utm_zone: UTM Zone string for the location, see https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system to identify the zone for the location.


        Raises:
        ValueError : If the bounding box is more than 500m x 500m square
        
        
        """
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl:float = 50.0
        self.altitude_envelope: int = 15 # the buffer in meters for flight when a path is converted into a volume
        self.control_flight_geometry: Union[LineString, Polygon] # the initial flight path or geometry against which subsequent flight paths are generated, this flag
        
        self.raw_geometries: List[GeometryPayload] # Object to hold polyons or linestrings, and the rule that generated the geometry (e.g. should this geometry intersect with the control)
        self.now = arrow.now()        
        self.geod = Geod(ellps="WGS84")
        self.grid_cells : List[shapely.geometry.box] # When a bounding box is given, it is split into smaller boxes this object holds the grids
        self.input_extents_valid()
        self.generate_grid_cells()
        
    def generate_grid_cells(self):
        # Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid.
        N_COLS = 3
        N_ROWS = 2
        cell_size_x = (self.maxx - self.minx)/(N_COLS)  # create three columns
        cell_size_y = (self.maxy - self.miny)/(N_ROWS)  # create two rows
        grid_cells = []
        for u0 in range(0, N_COLS):  # 3 columns
            x0 = self.minx + (u0 * cell_size_x)
            for v0 in range(0, N_ROWS):  # 2 rows
                y0 = self.miny + (v0 * cell_size_y)
                x1 = x0 + cell_size_x
                y1 = y0 + cell_size_y
                grid_cells.append(shapely.geometry.box(x0, y0, x1, y1))
        self.grid_cells = grid_cells

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
        
    def generate_random_flight_path_polygon(self, generate_polygon:bool) -> Union[LineString, Polygon]:
        '''Generate a random flight path or polygon, if a polygon is specified then this method picks one of the grid cells to generate the flight path within that, this is to ensure that a polygon geometry does not take over the entire bounding box. This code uses the `generate_random` method (https://github.com/jazzband/geojson/blob/master/geojson/utils.py#L131) to generate the initial linestring.  '''
        
        if generate_polygon:
            grid_cell = random.choice(self.grid_cells) # Pick a random grid cell
            random_flight_path_polygon = geojson.utils.generate_random(featureType = "LineString", numberVertices=2, boundingBox=grid_cell.bounds)
            random_flight_path_polygon = asShape(random_flight_path_polygon).envelope # Get the envelope of the linestring and create a box
            
        else: 
            random_flight_path_polygon = geojson.utils.generate_random(featureType = "LineString", numberVertices=2, boundingBox=[self.minx, self.miny, self.maxx, self.maxy])
            
        return random_flight_path_polygon
               
    
        
    def generate_single_flight_geometry(self, geometry_generation_rule:GeometryGenerationRule, is_control:bool= False) -> Union[LineString, Polygon]:
        ''' A method to generates flight geometry within a geographic bounds. The geometry can be a linestring or a polygon, simple rules for generation can be specificed. At the moment the method check if the geometry should intersect with the control and based on that, linestring / polygons are created '''
        
        coin_flip = random.choice([0,0,1])         
        flight_geometry = self.generate_random_flight_path_polygon(generate_polygon = coin_flip) 

        if is_control:
            self.control_flight_geometry = flight_geometry # Assign the control since this is the first time that the flight geometry is generated
        else: 
            if geometry_generation_rule.intersect_space: # This is not the first geometry, check if it should intersect with the control
                geometry_intersects = False
                while (geometry_intersects == False):
                    coin_flip = random.choice([0,0,1]) 
                    # We are trying to generate a path that intersects with the control, we keep generating linestrings or polygons till one is found that does intersect
                    flight_geometry = self.generate_random_flight_path_polygon(generate_polygon = coin_flip)
                    raw_geom = asShape(flight_geometry) # Generate a shape from the geometry
                    geometry_intersects = self.control_flight_geometry.intersects(raw_geom) # Check this intersects with the control                    
                
        return flight_geometry

    def convert_geometry_to_volume(self, flight_geometry:LineString, volume_generation_rule: VolumeGenerationRule, altitude_of_ground_level_wgs_84:int) -> Volume3D:
        ''' A method to convert a GeoJSON LineString or Polygon to a ASTM outline_polygon object by buffering 15m spatially '''
        
        flight_geometry_shp = asShape(flight_geometry)
        flight_geometry_utm = self.utm_converter(flight_geometry_shp)
        buffer_shape_utm = flight_geometry_utm.buffer(15)
        
        if volume_generation_rule.intersect_altitude: # If the flight should interect in altitude (altitude is kept same)
            alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_agl +self.altitude_envelope  
            alt_lower = altitude_of_ground_level_wgs_84 + self.altitude_agl - self.altitude_envelope
        else:

            raised_altitude_meters = random.choice([50,80])
            # Raise the altitude by 50m or 80m so that the flights do not intersect in altitude
            alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_agl  + raised_altitude_meters + self.altitude_envelope 
            alt_lower = altitude_of_ground_level_wgs_84 + self.altitude_agl + raised_altitude_meters - self.altitude_envelope 
        
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

    def transform_3d_volume_to_astm_4d(self, volume_3d : Volume3D,volume_generation_rule: VolumeGenerationRule) -> Volume4D:
        ''' This method converts a 3D Volume to 4D Volume and checks if the volumes should intersect in time, if the time interesection flag is turned off it will shift the volume start and end time to a arbirtray number in the next 70 mins. '''
        
        if volume_generation_rule.intersect_time: 
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
    
    def generate_raw_geometries(self, number_of_geometries:int = 6) -> List[GeometryPayload]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        
        raw_geometries: List[GeometryPayload] = []
        for volume_idx in range(0, number_of_geometries):
            is_control = 1 if (volume_idx == 0) else 0
            geometry_generation_rule = GeometryGenerationRule()
            
            if (volume_idx == (number_of_geometries -1)): # The first geometry is called "control" and it should not intersect with 
                should_intersect = False
            else:   
                coin_flip = random.choice([0,0,1]) # It can or cannot intersect
                should_intersect = coin_flip 

            geometry_generation_rule.intersect_space = should_intersect
            
            # the first path is control, the last path does not intersect the control
            current_path = self.generate_single_flight_geometry(geometry_generation_rule = geometry_generation_rule, is_control= is_control)
            raw_path = GeometryPayload(geometry = current_path, geometry_generation_rule = geometry_generation_rule, is_control = is_control)
            raw_geometries.append(raw_path)
        return raw_geometries

    def generate_volume_altitude_time_intersect_rules(self, raw_geometries:List[GeometryPayload]) -> List[VolumeGenerationRule]: 
        ''' A method to generate rules for generation of new paths '''
        
        all_volume_rules: List[VolumeGenerationRule] = []
        last_path_index = len(raw_geometries) - 1 
        for path_index, raw_path in enumerate(raw_geometries): 
            if path_index in [0,last_path_index]:
                # This the control path or the well clear path no need to have any time / atltitude interserction
                volume_generation_rule = VolumeGenerationRule(intersect_altitude= 0, intersect_time=0, expected_result = 1)
                if path_index == 0:
                    volume_generation_rule.is_control = 1
            else:
                # intersect in time and / or intersect in altitude 
                volume_generation_rule = random.choice([VolumeGenerationRule(intersect_altitude=1, intersect_time= 0, expected_result = 1),VolumeGenerationRule(intersect_altitude=1, intersect_time= 1, expected_result = 0),VolumeGenerationRule(intersect_altitude=0, intersect_time= 1, expected_result = 1)])
            all_volume_rules.append(volume_generation_rule)
        return all_volume_rules

    def generate_astm_4d_volumes(self,raw_geometries:List[GeometryPayload],rules : List[GeometryPayload], altitude_of_ground_level_wgs_84 :int) -> List[Volume4D]:
        ''' A method to generate ASTM specified Volume 4D payloads to submit to the system to be tested.  '''
        all_volume_4d :List[Volume4D] = []
        for path_index, raw_geometry in enumerate(raw_geometries): 
            volume_generation_rule = rules[path_index]
            
            flight_volume_3d = self.convert_geometry_to_volume(flight_geometry = raw_geometry.geometry, volume_generation_rule = volume_generation_rule, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
            flight_volume_4d = self.transform_3d_volume_to_astm_4d(volume_3d = flight_volume_3d, volume_generation_rule = volume_generation_rule)
            all_volume_4d.append(flight_volume_4d)
            
        return all_volume_4d


class SCDFlightPathVolumeWriter():
    """ A class to write raw Flight Paths and volumes to disk so that they can be examined / used in other software """

    def __init__(self, raw_geometries:List[GeometryPayload],  flight_volumes: List[Volume4D], all_rules: List[VolumeGenerationRule],country_code='che')-> None:
        '''
        A method to write flight paths, volumes and volume generation rules to disk for review / submission to the test harness. All data is written in the `test_definitions` directory which is created if it does not exist.
        
        '''
        self.country_code = country_code
        self.output_directory = Path('test_definitions', self.country_code)        
        # Create test_definition directory if it does not exist        
        self.output_directory.mkdir(parents=True, exist_ok=True)
        self.output_subdirectories = (Path(self.output_directory, 'astm_4d_volumes'), Path(self.output_directory, '4d_flight_geometries'),Path(self.output_directory, 'scd_rules'),)
        for output_subdirectory in self.output_subdirectories:
            output_subdirectory.mkdir(parents=True, exist_ok=True)
        self.raw_geometries = raw_geometries
        self.flight_volumes = flight_volumes
        self.all_rules = all_rules

    def write_4d_flight_geometries(self) -> None:
        ''' A method to write flight paths to disk as GeoJSON features '''
        
        for path_id, raw_geometry in enumerate(self.raw_geometries):    
            line_polygon_feature = {'type': 'Feature', 'properties': {}, 'geometry': shapely.geometry.mapping(raw_geometry.geometry)}
            scd_volume_4d = SCDVolume4D(flight_plan= line_polygon_feature,time_start = self.flight_volumes[path_id].time_start, time_end=self.flight_volumes[path_id].time_end, altitude_lower= self.flight_volumes[path_id]['volume']['altitude_lower'], altitude_upper= self.flight_volumes[path_id]['volume']['altitude_upper'] )
            
            path_file_name = '4d_path_%s.json' % str(path_id + 1)  # Avoid Zero based numbering
            tracks_file_path = self.output_subdirectories[1] / path_file_name
            with open(tracks_file_path, 'w') as f:
                f.write(json.dumps(scd_volume_4d))

    
    def write_astm_volume_4d(self) -> None:
        ''' A method to write volume 4D objects to disk '''

        for volume_id, volume in enumerate(self.flight_volumes):                     
            volume_file_name = 'volume_%s.json' % str(volume_id + 1)  # Avoid Zero based numbering           
            volume_file_path = self.output_subdirectories[0] / volume_file_name
            with open(volume_file_path, 'w') as f:
                f.write(json.dumps(volume))

    def write_test_rules(self) -> None:
        ''' A method to write the expected outcomes of sequential input '''
        
        volume_file_name = 'volume_rules.json'
        rule_file_path = self.output_subdirectories[2] / volume_file_name
        with open(rule_file_path, 'w') as f:
            f.write(json.dumps(self.all_rules))
    
if __name__ == '__main__':
    COUNTRY_CODE = 'che'
    # Generate volumes 
    my_volume_generator = ProximateFlightVolumeGenerator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')
    altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit
    # Change directory to write test_definitions folder is created in the rid_qualifier folder.
    p = pathlib.Path(__file__).parent.absolute()
    os.chdir(p)
    flight_geometries = my_volume_generator.generate_raw_geometries(number_of_geometries=12)
    all_rules = my_volume_generator.generate_volume_altitude_time_intersect_rules(raw_geometries=flight_geometries)
    
    flight_volumes = my_volume_generator.generate_astm_4d_volumes(raw_geometries = flight_geometries, rules = all_rules, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
   
    my_flight_geometry_writer = SCDFlightPathVolumeWriter(raw_geometries= flight_geometries, flight_volumes=flight_volumes, all_rules = all_rules)

    my_flight_geometry_writer.write_4d_flight_geometries()    
    my_flight_geometry_writer.write_astm_volume_4d()
    my_flight_geometry_writer.write_test_rules()
