import aircraft_state_replayer
from urllib.parse import urlparse
import uuid, os
from pathlib import Path
import json
from typing import List, NamedTuple, Any
import arrow
import datetime
from datetime import datetime, timedelta
import utils
from utils import UTMSP, OperatorLocation, Operator, RIDFlightDetails, TestFlightDetails, TestFlight
from test_executor import TestHarness


class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''

    def __init__(self, test_config: str, country_code='che') -> None:
        
        self.test_config_valid(test_config)
        self.test_config = json.loads(test_config)
        self.tracks_directory = Path('test_definitions', country_code, 'aircraft_states')
        self.verify_tracks_directory(self.tracks_directory)
        self.flight_tracks = self.load_flight_tracks(self.tracks_directory)

        self.rid_serializer = utils.RIDSerializer()
    
    def load_flight_tracks(self, tracks_directory) -> None:
        track_files = os.listdir(tracks_directory) 
        return track_files


    def verify_tracks_directory(self, tracks_directory) -> None:

        ''' This method checks if there are tracks in the tracks directory '''        
        
        files = [f for f in os.listdir(tracks_directory) if os.path.isfile(os.path.join(tracks_directory, f))]
        if files:
            pass
        else:
            raise ValueError("The there are no tracks in the tracks directory, create tracks first using the flight_data_generator module. ")

    def test_config_valid(self, test_config: str) -> None:
        ''' This method checks if the test definition is a valid JSON ''' #TODO : Have a comprehensive way to check JSON definition
        if json.loads(test_config):
            pass
        else:
            raise ValueError("A valid JSON object must be submitted ")


    def build_test_payload(self): 
        ''' This is the main method to process the test configuration and build RID payload object, maxium of one flight is allocated to each UTMSP. '''
        
        utm_sps = self.test_config['utmsps']

        all_test_payloads = []
        
        for utmsp_index, utm_sp in enumerate(utm_sps):
            requested_flights = []
            flight_track_path = Path(self.tracks_directory, self.flight_tracks[utmsp_index])
            with open(flight_track_path) as generated_rid_state:
                rid_state_data = json.load(generated_rid_state)

            
            effective_after = rid_state_data['reference_time']
            
            operator_location = OperatorLocation(lat = utm_sp['operator_details'][utmsp_index]['location']['latitude'], lng = utm_sp['operator_details'][utmsp_index]['location']['longitude'])
            operator_id = str(uuid.uuid4())
            operator = Operator(id = operator_id, location = operator_location, operation_description=  utm_sp['flight_details'][utmsp_index]['operation_description'] , serial_number = utm_sp['flight_details'][utmsp_index]['serial_number'],  registration_number =  utm_sp['flight_details'][utmsp_index]['registration_number'])

            rid_flight_details = RIDFlightDetails(operator_id = operator.id, operation_description = operator.operation_description, serial_number = operator.serial_number, registration_number = operator.registration_number)

            test_flight_details = TestFlightDetails(effective_after= effective_after,details = rid_flight_details)
            test_flight = TestFlight(injection_id = str(uuid.uuid4()), telemetry= rid_state_data['flight_telemetry'], details_respones= test_flight_details)            
            test_flight_deserialized = self.rid_serializer.make_json_compatible(test_flight)
            requested_flights.append(test_flight_deserialized)
            test_payload = {'test_id': str(uuid.uuid4()), "requested_flights": requested_flights}        

            all_test_payloads.append({'injection_url':utm_sp['injection_url'], 'injection_payload': test_payload})        
        
        return all_test_payloads


class TestSubmitter():
    ''' A class to submit the test data to the UTMSP end point '''

    def __init__(self, test_payloads):
        self.test_payload_valid(test_payloads)

        self.submit_payload(test_payloads)


    def test_payload_valid(self, test_payloads: List) -> None:
        ''' This method checks if the test definition is a valid JSON ''' #TODO : Have a comprehensive way to check JSON definition
        if len(test_payloads):
            pass
        else:
            raise ValueError("A valid payload object with atleast one flight / UTMSP must be submitted")

    def submit_payload(self, test_payloads: List) -> None:
        ''' This method submits the payload to indvidual UTMSP '''
        my_test_harness = TestHarness()
        for payload in test_payloads: 
            my_test_harness.submit_test(payload)
        


class RIDDataValidator():
    ''' A class to check the output from UTMSP in Flight Blender and produce a text report and write on disk '''
    pass


if __name__ == '__main__':
        
    # This is the configuration for the test.
    test_configuration = {
        "locale": "che",
        "utmsps": [
            {
                "name": "Unmanned Systems Corp.",
                "dss_audience":"dss.unmanned.corp",
                "injection_url": "https://dss.unmanned.corp/tests/",
                "flight_details": [
                    {
                        "serial_number": "C1A10C76-22D9-44E7",
                        "registration_number": "CHE87astrdge12k8",
                        "operation_description": "Electricity Grid Inspection"
                    }
                ],
                "operator_details": [
                    {
                        "name": "Electricity Inspection Company",
                        "location": {
                            "latitude": 46.974432835242055,
                            "longitude": 7.472983002662658
                        }
                    }
                ]
            }
        ]
    }
    my_test_builder = TestBuilder(test_config = json.dumps(test_configuration), country_code='CHE')    
    test_payloads = my_test_builder.build_test_payload()

    my_test_submitter = TestSubmitter(test_payloads= test_payloads)
