from monitoring.monitorlib.scd_automated_testing.scd_injection_api import OperationalIntentTestInjection,FlightAuthorisationData, InjectFlightRequest
from utils import GeneratedGeometry, GeometryGenerationRule, RequiredResults,TestInjectionRequiredResult
from shapely.geometry import asShape
from shapely.geometry import LineString
from monitoring.monitorlib.scd import Time, Volume3D, Volume4D, Polygon, Altitude, LatLngPoint
from typing import List
import geojson
from pyproj import Geod, Proj
import arrow
import random
from typing import List, Union
import shapely.geometry

class ProximateOperationalIntentGenerator():
    ''' A class to generate operational intents. As a input the module takes in a bounding box for which to generate the volumes within. '''

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        ''' Create a ProximateVolumeGenerator within a given geographic bounding box. 

        Once these extents are specified, a grid will be created with two rows. A combination of LineStrings and Polygons will be generated withing these bounds. While linestrings can extend to the full boundaries of the box, polygon areas are generated within the grid. 

        Args:
        minx: Western edge of bounding box (degrees longitude)
        maxx: Eastern edge of bounding box (degrees longitude)
        miny: Southern edge of bounding box (degrees latitude)
        maxy: Northern edge of bounding box (degrees latitude)
        utm_zone: UTM Zone string for the location, see https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system to identify the zone for the location.


        Raises:
        ValueError : If the bounding box is more than 500m x 500m square
        
        
        '''
        self.minx = minx
        self.miny = miny
        self.maxx = maxx
        self.maxy = maxy
        self.utm_zone = utm_zone

        self.altitude_agl:float = 50.0
        self.altitude_envelope: int = 15 # the buffer in meters for flight when a path is converted into a volume
        self.control_flight_geometry: Union[LineString, Polygon] # the initial flight path or geometry against which subsequent flight paths are generated, this flag
        
        self.raw_geometries: List[GeneratedGeometry] # Object to hold polyons or linestrings, and the rule that generated the geometry (e.g. should this geometry intersect with the control)
        self.now = arrow.now()        
        self.geod = Geod(ellps='WGS84')
        self.grid_cells : List[shapely.geometry.box] # When a bounding box is given, it is split into smaller boxes this object holds the grids
        self._input_extents_valid()
        self._generate_grid_cells()
        
    def _generate_grid_cells(self):
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

        proj = Proj(proj='utm', zone=self.utm_zone, ellps='WGS84', datum='WGS84')

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

    def _input_extents_valid(self) -> None:
        ''' This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary.'''

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure enough space for tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError('The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square')
        
    def _generate_random_flight_path(self) -> LineString:
        '''Generate a random flight path. this code uses the `generate_random` method (https://github.com/jazzband/geojson/blob/master/geojson/utils.py#L131) to generate the initial linestring.  '''
        
        random_flight_path = geojson.utils.generate_random(featureType = 'LineString', numberVertices=2, boundingBox=[self.minx, self.miny, self.maxx, self.maxy])

        return random_flight_path
               
    def _generate_random_flight_polygon(self) -> Polygon:
        '''Generate a random polygon, if a polygon is specified then this method picks one of the grid cells to generate the flight path within that, this is to ensure that a polygon geometry does not take over the entire bounding box. '''        
        
        grid_cell = random.choice(self.grid_cells) # Pick a random grid cell
        random_flight_polygon = geojson.utils.generate_random(featureType = 'LineString', numberVertices=2, boundingBox=grid_cell.bounds)
        random_flight_polygon = asShape(random_flight_polygon).envelope # Get the envelope of the linestring and create a box
        return random_flight_polygon
        
    def _generate_single_flight_geometry(self, geometry_generation_rule:GeometryGenerationRule, is_control:bool= False) -> Union[LineString, Polygon]:
        ''' A method to generates flight geometry within a geographic bounds. The geometry can be a linestring or a polygon, simple rules for generation can be specificed. At the moment the method check if the geometry should intersect with the control and based on that, linestring / polygons are created '''
        
        coin_flip = random.choice([0,0,1])         
        if coin_flip:
            flight_geometry = self._generate_random_flight_polygon()
        else:
            flight_geometry = self._generate_random_flight_path()

        if is_control:
            self.control_flight_geometry = asShape(flight_geometry) # Assign the control since this is the first time that the flight geometry is generated
        else: 
            if geometry_generation_rule.intersect_space: # This is not the first geometry, check if it should intersect with the control
                geometry_intersects = False
                while (geometry_intersects == False):
                    coin_flip = random.choice([0,0,1]) 
                    # We are trying to generate a path that intersects with the control, we keep generating linestrings or polygons till one is found that does intersect
                    if coin_flip:
                        flight_geometry = self._generate_random_flight_polygon()
                    else:
                        flight_geometry = self._generate_random_flight_path()
                        
                    raw_geom = asShape(flight_geometry) # Generate a shape from the geometry
                    geometry_intersects = self.control_flight_geometry.intersects(raw_geom) # Check this intersects with the control                    
                
        return flight_geometry

    def convert_geometry_to_volume(self, flight_geometry:LineString, altitude_of_ground_level_wgs_84:int) -> Volume3D:
        ''' A method to convert a GeoJSON LineString or Polygon to a ASTM outline_polygon object by buffering 15m spatially '''
        
        flight_geometry_shp = asShape(flight_geometry)
        flight_geometry_utm = self.utm_converter(flight_geometry_shp)
        buffer_shape_utm = flight_geometry_utm.buffer(15)        

        alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_agl +self.altitude_envelope  
        alt_lower = altitude_of_ground_level_wgs_84 + self.altitude_agl - self.altitude_envelope

        buffered_shape_geo = self.utm_converter(buffer_shape_utm, inverse=True)
        
        all_vertices = []

        altitude_upper = Altitude(value= alt_upper, reference = 'W84', units='M')
        altitude_lower = Altitude(value=alt_lower, reference = 'W84', units='M')
        for vertex in list(buffered_shape_geo.exterior.coords):
            coord = LatLngPoint(lat = vertex[0] , lng = vertex[1])
            all_vertices.append(coord)
        
        p = Polygon(vertices=all_vertices)
        
        volume3D = Volume3D(outline_polygon = p, altitude_lower = altitude_lower, altitude_upper = altitude_upper, outline_circle = {})
        
        return volume3D

    def transform_3d_volume_to_astm_4d(self, volume_3d : Volume3D) -> Volume4D:
        ''' This method converts a 3D Volume to 4D Volume, the flight start time is 3 mins from now  '''
    
        three_mins_from_now = self.now.shift(minutes = 3)
        eight_mins_from_now = self.now.shift(minutes = 8)
        start_time = Time(value = three_mins_from_now.isoformat(), format = 'RFC3339')
        end_time = Time(value = eight_mins_from_now.isoformat(), format = 'RFC3339')    
        volume_4D = Volume4D(volume=volume_3d, time_start= start_time, time_end=end_time)
        
        return volume_4D
    
    def generate_raw_geometries(self, number_of_geometries:int = 6) -> List[GeneratedGeometry]:
        ''' A method to generate Volume 4D payloads to submit to the system to be tested.  '''
        
        raw_geometries: List[GeneratedGeometry] = []
        for volume_idx in range(0, number_of_geometries):
            is_control = 1 if (volume_idx == 0) else 0
            
            
            if (volume_idx == (number_of_geometries -1)): # The first geometry is called 'control' and it should not intersect with 
                should_intersect = False
            else:   
                coin_flip = random.choice([0,0,1]) # It can or cannot intersect
                should_intersect = coin_flip

            geometry_generation_rule = GeometryGenerationRule(intersect_space = should_intersect)
            # the first path is control, the last path does not intersect the control
            current_path = self._generate_single_flight_geometry(geometry_generation_rule = geometry_generation_rule, is_control= is_control)
            raw_path = GeneratedGeometry(geometry = current_path, geometry_generation_rule = geometry_generation_rule, is_control = is_control)
            raw_geometries.append(raw_path)
        return raw_geometries


    def generate_astm_4d_volumes(self,raw_geometries:List[GeneratedGeometry], altitude_of_ground_level_wgs_84 :int) -> List[Volume4D]:
        ''' A method to generate ASTM specified Volume 4D payloads to submit to the system to be tested.  '''

        all_volume_4d :List[Volume4D] = []
        for path_index, raw_geometry in enumerate(raw_geometries):            
            
            flight_volume_3d = self.convert_geometry_to_volume(flight_geometry = raw_geometry.geometry, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
            flight_volume_4d = self.transform_3d_volume_to_astm_4d(volume_3d = flight_volume_3d)
            all_volume_4d.append(flight_volume_4d)            
        return all_volume_4d
    
    def generate_injection_operational_intents(self, astm_4d_volumes:List[Volume4D]) -> List[OperationalIntentTestInjection ]:
        ''' A method to generate Operational Intent references given a list of Volume 4Ds '''

        all_operational_intent_references= []
        for current_volume in astm_4d_volumes: 
            current_operational_intent_reference = OperationalIntentTestInjection(volumes = [current_volume], key = [], state = 'Accepted', off_nominal_volumes = [], priority =0)
            all_operational_intent_references.append(current_operational_intent_reference)            
        return all_operational_intent_references

class FlightAuthorisationDataGenerator():
    ''' A class to generate data for flight authorisation per the ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664 for an UAS flight authorisation request. Reference: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e32-178-1 
    ''' 

    def __init__(self):
        '''
        This class generates a Flight Authorisation dataset, the dataset contains 11 fields at any time one of the authorisation data parameter would be incorrect this class generates a Flight Authorisation dataset 
        '''
        
        self.serial_number_length_code_points = {'1':1,'2':2,'3':3,'4':4,'5':5,'6':6,'7':7,'8':8,'9':9,'A':10,'B':11,'C':12,'D':13,'E':14,'F':15}
        self.serial_number_code_points = ['0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F','G','H','J','K','L','M','N','P','Q','R','S','T','U','V','W','X','Y','Z']
        self.registration_number_code_points = ['0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f','g','h','i','j','k','l','m','n','o','p','q','r','s','t','u','v','w','x','y','z']


    def generate_incorrect_serial_number(self, valid_serial_number:str) ->str:
        ''' 
        A method to modify a valid UAV serial number per ANSI/CTA-2063-A standard to one that does not conform to the standard.         
        '''
        _serial_number_length_code_points = self.serial_number_length_code_points # make a copy of the the code points
        manufacturer_code = valid_serial_number[0:4] # take out the manufacturer serial code out of the valid one
        length_code = valid_serial_number[4:5] # take out the length code out of the valid serial number         
        _serial_number_length_code_points.pop(length_code, None) # take out the length code so that we dont pick that one again (and make the serial number valid)
        dict_key, new_length_code = random.choice(list(_serial_number_length_code_points.items())) # pick a new length code
        random_serial_number = ''.join(random.choices(self.serial_number_code_points, k = new_length_code)) # generate anew 

        incorrect_serial_number =  manufacturer_code + length_code + random_serial_number

        return incorrect_serial_number

    def generate_serial_number(self) -> str:
        ''' 
        A method to generate a random UAV serial number per ANSI/CTA-2063-A standard        
        '''
        
        random.shuffle(self.serial_number_code_points)
        manufacturer_code = ''.join(self.serial_number_code_points[:4])
        dict_key, length_code = random.choice(list(self.serial_number_length_code_points.items()))
        random_serial_number = ''.join(random.choices(self.serial_number_code_points, k=length_code))

        serial_number = manufacturer_code + dict_key + random_serial_number
        return serial_number

class TestInjectionRequiredResultsGenerator(): 
    '''A class to generate TestInjection and the associated results '''
    def __init__(self, num_injections:int):
        self.num_injections = num_injections

    def generate_injections_results(self) -> List[TestInjectionRequiredResult]:
        all_injections_results = []

        for injection_number in range(0,self.num_injections):
            my_flight_authorisation_data_generator = FlightAuthorisationDataGenerator()

            my_operational_intent_generator = ProximateOperationalIntentGenerator(minx=7.4735784530639648, miny=46.9746744128218410, maxx=7.4786210060119620, maxy=46.9776318195799121, utm_zone='32T')
            altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit
            altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit
            
            serial_number = my_flight_authorisation_data_generator.generate_serial_number()
            # TODO: Code to generate additional fields 

            make_incorrect = random.choice([0,1]) # a flag specify if one of the parameters of the flight_authorisation should be incorrect
            if make_incorrect: # if the flag is on, make the serial number incorrect        
                field_to_make_incorrect = random.choice(['uas_serial_number']) # Pick a field to make incorrect, TODO: Additional fields to be added as the generation code is impl 
                if field_to_make_incorrect == 'uas_serial_number':
                    serial_number = my_flight_authorisation_data_generator.generate_incorrect_serial_number(valid_serial_number = serial_number)
            
            flight_authorisation_data = FlightAuthorisationData(uas_serial_number = serial_number, operation_category='Open', operation_mode = 'Vlos',uas_class='C0', identification_technologies = ['ASTMNetRID'], connectivity_methods = ['cellular'], endurance_minutes = 30 , emergency_procedure_url = 'https://uav.com/emergency', operator_id = 'SUSz8k1ukxjfv463-brq', uas_id= '', uas_type_certificate = '')

            flight_geometries = my_operational_intent_generator.generate_raw_geometries(number_of_geometries=1)            

            flight_volumes = my_operational_intent_generator.generate_astm_4d_volumes(raw_geometries = flight_geometries, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
            
            operational_intent_test_injection = my_operational_intent_generator.generate_injection_operational_intents(astm_4d_volumes = flight_volumes)
        
            inject_flight_request = InjectFlightRequest(operational_intent= operational_intent_test_injection, flight_authorisation= flight_authorisation_data)
            authorisation_data_fields_to_check = []
            operational_intent_validation_checks = []

            if make_incorrect: 
                expected_injection_result = 'Rejected'    
                authorisation_data_fields_to_check = ['uas_serial_number']
            else:
                expected_injection_result = 'Planned'
            
            required_result = RequiredResults(expected_response=expected_injection_result,authorisation_data_fields_to_check = authorisation_data_fields_to_check, operational_intent_validation_checks=operational_intent_validation_checks)
            
            all_injections_results.append(TestInjectionRequiredResult(test_injection=inject_flight_request,required_result=required_result))

        return all_injections_results
           

if __name__ == '__main__':    
    my_test_injection_results_generator = TestInjectionRequiredResultsGenerator(num_injections =2)
    injections_results = my_test_injection_results_generator.generate_injections_results()    
    print(injections_results)
