from monitoring.monitorlib.scd_automated_testing.scd_injection_api import OperationalIntentTestInjection,FlightAuthorisationData, InjectFlightRequest
from .utils import GeneratedGeometry, VolumeGenerationRule
from shapely.geometry import LineString
from monitoring.monitorlib import Volume3D, Volume4D
from typing import List
import random

class ProximateOperationalIntentGenerator():
    ''' A class to generate operational intents. As a input the module takes in a bounding box for which to generate the volumes within. Further test'''

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
        pass

    def generate_raw_geometries(self, number_of_geometries:int = 6) -> List[GeneratedGeometry]:
        ''' A method to generate Volume 4D data '''
        
        raise NotImplementedError("")

    def convert_geometry_to_volume_3D(self, flight_geometry:LineString, volume_generation_rule: VolumeGenerationRule, altitude_of_ground_level_wgs_84:int) -> Volume3D:
        ''' A method to convert a GeoJSON LineString or Polygon to a ASTM outline_polygon object by buffering 15m spatially '''

        raise NotImplementedError("")
        
    def generate_astm_4d_volumes(self,raw_geometries:List[GeneratedGeometry],rules : List[GeneratedGeometry], altitude_of_ground_level_wgs_84 :int) -> List[Volume4D]:
        ''' A method to generate ASTM specified Volume 4D payloads to submit to the system to be tested.  '''
        
        raise NotImplementedError("")

    def generate_operational_intent_test_injection(self, astm_4d_volumes:List[Volume4D]) -> List[OperationalIntentTestInjection]:
        ''' A method to generate Operational Intent references given a list of Volume 4Ds '''

        raise NotImplementedError("")

    def generate_volume_altitude_time_intersect_rules(self, raw_geometries:List[GeneratedGeometry]) -> List[VolumeGenerationRule]: 
        ''' A method to generate rules for generation of new paths '''

        raise NotImplementedError("")



class FlightAuthorisationDataGenerator():
    ''' A class to generate data for flight authorisation per the ANNEX IV of COMMISSION IMPLEMENTING REGULATION (EU) 2021/664 for an UAS flight authorisation request. Reference: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e32-178-1 
    ''' 

    def __init__(self):
        '''
        This class generates a Flight Authorisation dataset, the dataset contains 11 fields at any time one of the authorisation data parameter would be incorrect this class generates a Flight Authorisation dataset 
        '''
        pass

    def generate_incorrect_serial_number(self, valid_serial_number:str) ->str:
        ''' 
        A method to modify a valid UAV serial number per ANSI/CTA-2063-A standard to one that does not conform to the standard.         
        '''
        raise NotImplementedError("Incorrect Serial Number generation not implemented")

    def generate_serial_number(self) -> str:
        ''' 
        A method to generate a random UAV serial number per ANSI/CTA-2063-A standard        
        '''
        
        raise NotImplementedError("Correct Serial Number generation not implemented")


if __name__ == '__main__':
    ''' This module generates a JSON that can be used to submit to the test interface '''
    
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
            serial_number = my_flight_authorisation_data_generator.generate_incorrect_serial_number(serial_number = serial_number)

    
    flight_authorisation_data = FlightAuthorisationData(uas_serial_number = serial_number, operation_category='Open', operation_mode = 'Vlos',uas_class='C0', identification_technologies = ['ASTMNetRID'], connectivity_methods = ['cellular'], endurance_minutes = 30 , emergency_procedure_url = "https://uav.com/emergency", operator_id = 'SUSz8k1ukxjfv463-brq', uas_id= '')

    flight_geometries = my_operational_intent_generator.generate_raw_geometries(number_of_geometries=1)
    all_rules = my_operational_intent_generator.generate_volume_altitude_time_intersect_rules(raw_geometries=flight_geometries)

    flight_volumes = my_operational_intent_generator.generate_astm_4d_volumes(raw_geometries = flight_geometries, rules = all_rules, altitude_of_ground_level_wgs_84 = altitude_of_ground_level_wgs_84)
    
    operational_intent_test_injection = my_operational_intent_generator.generate_operational_intent_test_injection(astm_4d_volumes = flight_volumes)
   
    inject_flight_request = InjectFlightRequest(operational_intent= operational_intent_test_injection, flight_authorisation= flight_authorisation_data)