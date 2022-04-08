from monitoring.monitorlib.scd_automated_testing.scd_injection_api import OperationalIntentTestInjection,FlightAuthorisationData, InjectFlightRequest
from monitoring.uss_qualifier.scd.data_interfaces import FlightInjectionAttempt, InjectionTarget, KnownResponses, AutomatedTest, TestStep
from monitoring.uss_qualifier.scd.simulator.utils import TestOutputPathDetails, AutomatedTestDetails, FlightNameIncorrectField
from utils import GeneratedGeometry, GeometryGenerationRule
from monitoring.monitorlib.formats import OperatorRegistrationNumber, SerialNumber
from shapely.geometry import asShape
from shapely.geometry import LineString
from monitoring.monitorlib.scd import Time, Volume3D, Volume4D, Polygon, Altitude, LatLngPoint
from typing import List, Tuple
from pathlib import Path
import geojson
import json
from pyproj import Geod, Proj
import arrow
import random, string
from typing import List, Union
import shapely.geometry
import os
import known_issues_generator
import existing_test_parser


class ProximateOperationalIntentGenerator():
    """A class to generate operational intents. As a input the module takes in a bounding box for which to generate the volumes within. """

    def __init__(self, minx: float, miny: float, maxx: float, maxy: float, utm_zone:str) -> None:
        """Create a ProximateVolumeGenerator within a given geographic bounding box.

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
        self.first_flight_geometry: Union[LineString, Polygon] # the initial flight path or geometry against which subsequent flight paths are generated, this flag

        self.raw_geometries: List[GeneratedGeometry] # Object to hold polyons or linestrings, and the rule that generated the geometry (e.g. should this geometry intersect with the control)
        self.now = arrow.now()
        self.geod = Geod(ellps='WGS84')
        self.grid_cells : List[shapely.geometry.box] # When a bounding box is given, it is split into smaller boxes this object holds the grids
        self._input_extents_valid()
        self._generate_grid_cells()

    def _generate_grid_cells(self):
        """Compute the box where the flights will be created. For a the sample bounds given, over Bern, Switzerland, a division by 2 produces a cell_size of 0.0025212764739985793, a division of 3 is 0.0016808509826657196 and division by 4 0.0012606382369992897. As the cell size goes smaller more number of flights can be accomodated within the grid. For the study area bounds we build a 3x2 box for six flights by creating 3 column 2 row grid. """
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
        """A helper function to convert from lat / lon to UTM coordinates for buffering. tracks. This is the UTM projection (https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system), we use Zone 33T which encompasses Switzerland, this zone has to be set for each locale / city. Adapted from https://gis.stackexchange.com/questions/325926/buffering-geometry-with-points-in-wgs84-using-shapely """

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
        """This method checks if the input extents are valid i.e. small enough, if the extent is too large, we reject them, at the moment it checks for extents less than 500m x 500m square but can be changed as necessary."""

        box = shapely.geometry.box(self.minx, self.miny, self.maxx, self.maxy)
        area = abs(self.geod.geometry_area_perimeter(box)[0])

        # Have a area less than 500m x 500m square and more than 300m x 300m square to ensure enough space for tracks
        if (area) < 250000 and (area) > 90000:
            return
        else:
            raise ValueError('The extents provided are not of the correct size, please provide extents that are less than 500m x 500m and more than 300m x 300m square')

    def _generate_random_flight_path(self) -> LineString:
        """Generate a random flight path. this code uses the `generate_random` method (https://github.com/jazzband/geojson/blob/master/geojson/utils.py#L131) to generate the initial linestring. """

        random_flight_path = geojson.utils.generate_random(featureType = 'LineString', numberVertices=2, boundingBox=[self.minx, self.miny, self.maxx, self.maxy])

        return random_flight_path

    def _generate_random_flight_polygon(self) -> Polygon:
        """Generate a random polygon, if a polygon is specified then this method picks one of the grid cells to generate the flight path within that, this is to ensure that a polygon geometry does not take over the entire bounding box. """

        grid_cell = random.choice(self.grid_cells) # Pick a random grid cell
        random_flight_polygon = geojson.utils.generate_random(featureType = 'LineString', numberVertices=2, boundingBox=grid_cell.bounds)
        random_flight_polygon = asShape(random_flight_polygon).envelope # Get the envelope of the linestring and create a box
        return random_flight_polygon

    def _generate_single_flight_geometry(self, geometry_generation_rule:GeometryGenerationRule, injection_number:int) -> Union[LineString, Polygon]:
        """A method to generates flight geometry within a geographic bounds. The geometry can be a linestring or a polygon, simple rules for generation can be specificed. At the moment the method check if the geometry should intersect with the control and based on that, linestring / polygons are created """

        coin_flip = random.choice([0,0,1])
        if coin_flip:
            flight_geometry = self._generate_random_flight_polygon()
        else:
            flight_geometry = self._generate_random_flight_path()

        if injection_number == 0:
            self.first_flight_geometry = asShape(flight_geometry)

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

                geometry_intersects = self.first_flight_geometry.intersects(raw_geom) # Check this intersects with the control

        return flight_geometry

    def convert_geometry_to_volume(self, flight_geometry:LineString, altitude_of_ground_level_wgs_84:int) -> Volume3D:
        """A method to convert a GeoJSON LineString or Polygon to a ASTM outline_polygon object by buffering 15m spatially """

        flight_geometry_shp = asShape(flight_geometry)
        flight_geometry_utm = self.utm_converter(flight_geometry_shp)
        buffer_shape_utm = flight_geometry_utm.buffer(15)

        alt_upper = altitude_of_ground_level_wgs_84 + self.altitude_agl +self.altitude_envelope
        alt_lower = altitude_of_ground_level_wgs_84 + self.altitude_agl - self.altitude_envelope

        buffered_shape_geo = self.utm_converter(buffer_shape_utm, inverse=True)

        all_vertices = []

        altitude_upper = Altitude(value= alt_upper, reference = 'W84', units='M')
        altitude_lower = Altitude(value=alt_lower, reference = 'W84', units='M')
        coords = list(buffered_shape_geo.exterior.coords)
        coords.pop() # remove the last item of a Shapely coords since it is the same as the first one       
        for vertex in coords:
            coord = LatLngPoint(lat = vertex[0] , lng = vertex[1])
            all_vertices.append(coord)

        p = Polygon(vertices=all_vertices)

        volume3D = Volume3D(outline_polygon = p, altitude_lower = altitude_lower, altitude_upper = altitude_upper, outline_circle = None)

        return volume3D

    def transform_3d_volume_to_astm_4d(self, volume_3d : Volume3D) -> Volume4D:
        """This method converts a 3D Volume to 4D Volume, the flight start time is 3 mins from now  """

        three_mins_from_now = self.now.shift(minutes = 3)
        eight_mins_from_now = self.now.shift(minutes = 8)
        start_time = Time(value = three_mins_from_now.isoformat(), format = 'RFC3339')
        end_time = Time(value = eight_mins_from_now.isoformat(), format = 'RFC3339')
        volume_4D = Volume4D(volume=volume_3d, time_start= start_time, time_end=end_time)

        return volume_4D

    def generate_nominal_test_geometry(self, geometry_generation_rule: GeometryGenerationRule, injection_number: int) -> GeneratedGeometry:
        """A method to generate two Volume 4D payloads to submit to the system to be tested.  """

        flight_path_geometry = self._generate_single_flight_geometry(geometry_generation_rule = geometry_generation_rule, injection_number= injection_number)

        raw_geometry = GeneratedGeometry(geometry = flight_path_geometry, geometry_generation_rule = geometry_generation_rule)

        return raw_geometry


    def generate_astm_4d_volumes(self,raw_geometry:GeneratedGeometry, altitude_of_ground_level_wgs_84 :int) -> Volume4D:
        """A method to generate ASTM specified Volume 4D payloads to submit to the system to be tested.  """

        flight_volume_3d = self.convert_geometry_to_volume(flight_geometry = raw_geometry.geometry, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
        flight_volume_4d = self.transform_3d_volume_to_astm_4d(volume_3d = flight_volume_3d)

        return flight_volume_4d

def generate_valid_flight_authorisation_data_for_nominal_test(locale:str ='CHE') -> FlightAuthorisationData:
    """A method to generate valid flight authorisation data for the nominal test, in the nominal test we are providing valid flight authorisation data since the operational intent processing is the main intent of the test """

    serial_number = SerialNumber.generate_valid()
    operator_id = OperatorRegistrationNumber.generate_valid(prefix=locale)

    flight_authorisation_data = FlightAuthorisationData(uas_serial_number = serial_number, operation_category="Open", operation_mode = "Vlos",uas_class="C0", identification_technologies = ["ASTMNetRID"], connectivity_methods = ["cellular"], endurance_minutes = 30, emergency_procedure_url = "https://uav.com/emergency", operator_id = operator_id, uas_id= '', uas_type_certificate = '')

    return flight_authorisation_data


def generate_operational_intents_for_flight_authorisation_test(num_operational_intents:int, bbox:Tuple[float, float, float, float]=None, utm_zone:str=None)->List[OperationalIntentTestInjection]:
    """A method to generate well clear operational intents and use them in the flight authorisation data format tests """

    all_operational_intent_test_injections = []
    if bbox:
        minx, miny, maxx, maxy = bbox
    else: 
        minx, miny, maxx, maxy = (7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,)
    utm_zone = utm_zone if utm_zone else "32T"
    my_operational_intent_generator = ProximateOperationalIntentGenerator(minx=minx, miny=miny, maxx=maxx, maxy=maxy, utm_zone=utm_zone)
    altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit

    for injection_number in range(0,num_operational_intents):
        # The flight path geometry should not intersect 
        should_intersect = False                 
        geometry_generation_rule = GeometryGenerationRule(intersect_space = should_intersect)
        flight_geometry = my_operational_intent_generator.generate_nominal_test_geometry(geometry_generation_rule= geometry_generation_rule, injection_number = injection_number)        
        flight_volume = my_operational_intent_generator.generate_astm_4d_volumes(raw_geometry = flight_geometry, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
        operational_intent_test_injection = generate_operational_intent_injection(astm_4d_volume = flight_volume)
        all_operational_intent_test_injections.append(operational_intent_test_injection)

    return all_operational_intent_test_injections


def generate_nominal_test_flight_injection_attempts(all_flight_names: List[str],locale:str, bbox:Tuple[float, float, float, float], utm_zone:str=None, with_priority:bool = False) -> List[FlightInjectionAttempt]:
    """A method to generate flight injection data and associated messages in case the result of data processing is different from the expectation for the nominal test """

    nominal_test_flight_injection_attempts = []
    if bbox:
        minx, miny, maxx, maxy = bbox
    else: 
        minx, miny, maxx, maxy = (7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,)
    utm_zone = utm_zone if utm_zone else "32T"
    my_operational_intent_generator = ProximateOperationalIntentGenerator(minx=minx, miny=miny, maxx=maxx, maxy=maxy, utm_zone=utm_zone)

    altitude_of_ground_level_wgs_84 = 570 # height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit

    for injection_number, flight_name in enumerate(all_flight_names):  

        # If the test is with a nominal planning test or a nominal planning test with priority the second flight generated should always intersect the frist one 
        should_intersect = False if injection_number == 0 else True            
        if with_priority:
            expected_operational_intent_processing_result = 'Planned'
        else:
            expected_operational_intent_processing_result = 'ConflictWithFlight' if should_intersect else 'Planned'
        geometry_generation_rule = GeometryGenerationRule(intersect_space = should_intersect)
        flight_geometry = my_operational_intent_generator.generate_nominal_test_geometry(geometry_generation_rule= geometry_generation_rule, injection_number = injection_number)
        reference_time = my_operational_intent_generator.now.isoformat()
        flight_volume = my_operational_intent_generator.generate_astm_4d_volumes(raw_geometry = flight_geometry, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)

        if with_priority:
            priority = 0 if injection_number == 0 else 100
        else: 
            priority = 1

        operational_intent_test_injection = generate_operational_intent_injection(astm_4d_volume = flight_volume, priority=priority)
        valid_flight_authorisation_data = generate_valid_flight_authorisation_data_for_nominal_test(locale= locale)
        inject_flight_request = InjectFlightRequest(operational_intent= operational_intent_test_injection, flight_authorisation= valid_flight_authorisation_data)
        if with_priority:            
            nominal_test_w_priority_known_issues_acceptable_result_generator = known_issues_generator.NominalTestwPrioritiesKnownIssuesAcceptableResults(expected_flight_authorisation_processing_result ='Planned',expected_operational_intent_processing_result = expected_operational_intent_processing_result)
            all_incorrect_result_details = nominal_test_w_priority_known_issues_acceptable_result_generator.generate_nominal_test_with_priroties_known_issues_fields()            
            known_responses = KnownResponses(acceptable_results=[expected_operational_intent_processing_result], incorrect_result_details= all_incorrect_result_details)

        else:            
            nominal_test_known_issues_acceptable_result_generator = known_issues_generator.NominalTestKnownIssuesAcceptableResults(expected_flight_authorisation_processing_result = 'Planned',expected_operational_intent_processing_result = expected_operational_intent_processing_result)
            all_incorrect_result_details = nominal_test_known_issues_acceptable_result_generator.generate_nominal_test_known_issues_fields()            
            known_responses = KnownResponses(acceptable_results=[expected_operational_intent_processing_result], incorrect_result_details= all_incorrect_result_details)
            
        if with_priority:
            uss_role = "First-Mover USS" if injection_number ==0 else "Second USS"
        else:
            uss_role = "First-Mover USS" if injection_number ==0 else "Blocked USS"

        injection_target = InjectionTarget(uss_role = uss_role)
        

        flight_injection_attempt = FlightInjectionAttempt(reference_time = reference_time, test_injection = inject_flight_request, known_responses = known_responses,injection_target = injection_target, name = flight_name)

        nominal_test_flight_injection_attempts.append(flight_injection_attempt)

    return nominal_test_flight_injection_attempts

def generate_flight_authorisation_u_space_format_injection_attempt(flight_name:str, operational_intent_test_injection: OperationalIntentTestInjection,locale:str,field_to_make_incorrect:str = None) -> FlightInjectionAttempt:
    """A method to generate data for flight authorisation test and the associated injection attempt for the authorisation format test """

    serial_number = SerialNumber.generate_valid()
    operator_id = OperatorRegistrationNumber.generate_valid(prefix=locale)

    if field_to_make_incorrect == "uas_serial_number":
        serial_number = serial_number.make_invalid_by_changing_payload_length()
    elif field_to_make_incorrect == "operator_registration_number":
        operator_id = operator_id.make_invalid_by_changing_final_control_string()

    expected_flight_authorisation_processing_result = 'Rejected' if field_to_make_incorrect else 'Planned'

    flight_authorisation_data = FlightAuthorisationData(uas_serial_number = serial_number, operation_category="Open", operation_mode = "Vlos",uas_class="C0", identification_technologies = ["ASTMNetRID"], connectivity_methods = ["cellular"], endurance_minutes = 30, emergency_procedure_url = "https://uav.com/emergency", operator_id = operator_id, uas_id= '', uas_type_certificate = '')

    flight_auth_test_issues_acceptable_result_generator = known_issues_generator.FlightAuthorisationKnownIssuesAcceptableResults(expected_flight_authorisation_processing_result = expected_flight_authorisation_processing_result,expected_operational_intent_processing_result = 'Planned')

    inject_flight_request = InjectFlightRequest(operational_intent= operational_intent_test_injection, flight_authorisation= flight_authorisation_data)
    all_incorrect_result_details = flight_auth_test_issues_acceptable_result_generator.generate_flight_authorisation_test_known_issue_fields(incorrect_field= field_to_make_incorrect)

    injection_target = InjectionTarget(uss_role = "Submitting USS")

    known_responses = KnownResponses(acceptable_results=[expected_flight_authorisation_processing_result], incorrect_result_details= all_incorrect_result_details)

    reference_time = arrow.now().isoformat()
    flight_injection_attempt = FlightInjectionAttempt(reference_time = reference_time, name=flight_name, test_injection = inject_flight_request, known_responses = known_responses,injection_target = injection_target)

    return flight_injection_attempt


def generate_nominal_test_data(bbox:Tuple[float, float, float, float]=(7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,), utm_zone:str= '32T',locale:str ='CHE') -> AutomatedTestDetails:
    """A method to run the data generator to generate the nominal and flight authorisation data test and the associated steps"""

    ## Begin nominal test data generation ##
 
    all_flight_names = []
    injection_attempts = 2
    for injection_attempt in range(0,injection_attempts):
        random_flight_name = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(8))
        all_flight_names.append(random_flight_name)
    nominal_test_steps = []
    nominal_test_flight_injection_attempts = generate_nominal_test_flight_injection_attempts(all_flight_names = all_flight_names,bbox =bbox, utm_zone = utm_zone, locale=locale)

    # Build nominal test steps
    for idx, injection_attempt in enumerate(nominal_test_flight_injection_attempts):
        if idx == 0:
            nominal_test_step_1 = TestStep(name="Inject flight via First-mover USS", inject_flight = injection_attempt, delete_flight=None)
            nominal_test_steps.append(nominal_test_step_1)
        elif idx == 1:
            nominal_test_step_2 = TestStep(name="Inject flight via Blocked USS", inject_flight = injection_attempt, delete_flight=None)
            nominal_test_steps.append(nominal_test_step_2)
    
    # End build nominal test steps

    test_output_details = TestOutputPathDetails(group='astm-strategic-coordination', name ='nominal-planning-1')
    test_name = test_output_details.group +'/'+test_output_details.name
    nominal_test_details = AutomatedTest(name=test_name, steps = nominal_test_steps)
    nominal_test_and_output_details = AutomatedTestDetails(automated_test = nominal_test_details, output_path_details= test_output_details)
    
    return nominal_test_and_output_details
    

def generate_nominal_priority_test_data(bbox:Tuple[float, float, float, float]=(7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,), utm_zone:str= '32T',locale:str ='CHE') -> AutomatedTestDetails:
    """A method to run the data generator to generate the nominal and flight authorisation data test and the associated steps"""

    ## Begin nominal test (with priorities) data generation  ##
    all_flight_names = []
    injection_attempts = 2
    for injection_attempt in range(0,injection_attempts):
        random_flight_name = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(8))
        all_flight_names.append(random_flight_name)
    nominal_test_with_priority_steps = []
    nominal_test_with_priority_flight_injection_attempts = generate_nominal_test_flight_injection_attempts(all_flight_names = all_flight_names,bbox =bbox, utm_zone = utm_zone, locale=locale,with_priority=True)

    # Build nominal test (with priorities) steps
    for idx, injection_attempt in enumerate(nominal_test_with_priority_flight_injection_attempts):
        if idx == 0:
            nominal_test_with_priority_step_1 = TestStep(name="Inject flight via First-mover USS", inject_flight = injection_attempt, delete_flight=None)
            nominal_test_with_priority_steps.append(nominal_test_with_priority_step_1)
        elif idx == 1:
            nominal_test_with_priority_step_2 = TestStep(name="Inject flight via Second USS", inject_flight = injection_attempt, delete_flight=None)
            nominal_test_with_priority_steps.append(nominal_test_with_priority_step_2)
    
    # End build nominal test (with priorities) steps 

    test_with_priority_output_details = TestOutputPathDetails(group='astm-strategic-coordination', name ='nominal-planning-priority-1')
    with_priority_test_name = test_with_priority_output_details.group +'/'+test_with_priority_output_details.name
    nominal_test_with_priority_details = AutomatedTest(name=with_priority_test_name, steps = nominal_test_with_priority_steps)
    nominal_test_with_priority_and_output_details = AutomatedTestDetails(automated_test = nominal_test_with_priority_details, output_path_details= test_with_priority_output_details)
    
    return nominal_test_with_priority_and_output_details



def generate_flight_authorisation_test_data(bbox:Tuple[float, float, float, float]=(7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,), utm_zone:str= '32T',locale:str ='CHE') -> AutomatedTestDetails:
    """A method to run the data generator to generate the nominal and flight authorisation data test and the associated steps"""

    ## Begin flight authorisation test data generation  ##  
    fields_to_make_incorrect = ["uas_serial_number", "operator_registration_number", None]
    number_of_injections = len(fields_to_make_incorrect)

    all_flight_authorisation_test_flights = []

    for field_index in range(0,number_of_injections):
        random_flight_name = ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(8))
        flight_name_incorrect_field = FlightNameIncorrectField(flight_name = random_flight_name, incorrect_field = fields_to_make_incorrect[field_index])
        all_flight_authorisation_test_flights.append(flight_name_incorrect_field)

    
    all_operational_intents_for_flight_authorisation_test = generate_operational_intents_for_flight_authorisation_test(bbox = bbox, utm_zone= utm_zone, num_operational_intents= number_of_injections)

    flight_authorisation_test_steps = []
    # Build flight authorisation test steps
    for test_id, flight_auth_test_metadata in enumerate(all_flight_authorisation_test_flights):
        flight_authorisation_test_injection_attempt = generate_flight_authorisation_u_space_format_injection_attempt(field_to_make_incorrect=flight_auth_test_metadata.incorrect_field, flight_name= flight_auth_test_metadata.flight_name, operational_intent_test_injection = all_operational_intents_for_flight_authorisation_test[test_id], locale=locale)

        inject_test_step = TestStep(name="Inject Fight Authorisation data", inject_flight= flight_authorisation_test_injection_attempt, delete_flight=None)
        flight_authorisation_test_steps.append(inject_test_step)
        
    # End build flight authorisation test steps
    
    flight_authorisation_test_output_details = TestOutputPathDetails(group='u-space', name ='flight-authorisation-validation-1')
    flight_authorisation_test_name = flight_authorisation_test_output_details.group +'/'+ flight_authorisation_test_output_details.name
    
    flight_authorisation_test = AutomatedTest(name = flight_authorisation_test_name, steps = flight_authorisation_test_steps)
    flight_authorisation_test_details = AutomatedTestDetails(automated_test = flight_authorisation_test, output_path_details= flight_authorisation_test_output_details)  

    ## End flight authorisation test data generation ##

    return flight_authorisation_test_details
    

def generate_operational_intent_injection(astm_4d_volume : Volume4D, priority:int = 0) -> OperationalIntentTestInjection:
    """A method to generate Operational Intent injection by taking in Volume 4D object and a priority """
    current_operational_intent_reference = OperationalIntentTestInjection(volumes = [astm_4d_volume], state = 'Accepted', off_nominal_volumes = [], priority =priority)
    return current_operational_intent_reference

def update_existing_test_definition(output_path:os.path, new_test_data: AutomatedTestDetails, locale ="che") -> None:
    """This method updates the default test defnititions in the repository for the nominal test
    """
    # read file     
    automated_test_file_directory_name = new_test_data.output_path_details.group
    automated_test_file_directory = Path(output_path,locale, automated_test_file_directory_name)
    automated_test_file_directory.mkdir(parents=True, exist_ok=True)
    automated_test_file_name = new_test_data.output_path_details.name + '.json'
    automated_test_file = Path(automated_test_file_directory, automated_test_file_name)

    with open(automated_test_file, "r") as existing_test:
        current_test_definition = json.load(existing_test)  
    
    existing_test:AutomatedTest = existing_test_parser.parse_and_load_existing_test(current_test_definition=current_test_definition)

    # over write the steps with the new steps    
    updated_current_test = AutomatedTest(name = existing_test.name, uss_capabilities = existing_test.uss_capabilities, steps = new_test_data.automated_test.steps, )
    
    with open(automated_test_file, "w") as test_to_overwrite:
        test_to_overwrite.write(json.dumps(updated_current_test, indent=4))


if __name__ == '__main__':

    # Set the bounding box for operational intent generation, default is set to Bern, Switzerland
    minx, miny, maxx, maxy = (7.4735784530639648,46.9746744128218410,7.4786210060119620,46.9776318195799121,)
    # Set the UTM Zone for coversion / reprojection to spatial coordinates
    utm_zone="32T"
    # Set the locale. 
    locale = 'CHE'
    nominal_test_data = generate_nominal_test_data(bbox=(minx, miny, maxx, maxy,), utm_zone= utm_zone, locale = locale)
    flight_authorisation_data = generate_flight_authorisation_test_data(bbox=(minx, miny, maxx, maxy,), utm_zone= utm_zone, locale = locale)
    nominal_test_with_priorities_data = generate_nominal_priority_test_data(bbox=(minx, miny, maxx, maxy,), utm_zone= utm_zone, locale = locale)
    output_path = os.path.join(Path(__file__).parent.absolute(), "../test_definitions")
    update_existing_test_definition(output_path=output_path, new_test_data = nominal_test_data)
    update_existing_test_definition(output_path=output_path, new_test_data = flight_authorisation_data)
    update_existing_test_definition(output_path=output_path, new_test_data = nominal_test_with_priorities_data)
