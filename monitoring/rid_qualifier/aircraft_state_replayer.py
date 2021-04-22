import requests
from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import DSSTestSession

import json, os
import uuid
from pathlib import Path
from typing import List, NamedTuple, Any
from utils import RIDSP, OperatorLocation, Operator, RIDFlightDetails, TestFlightDetails, TestFlight, RIDFlight
from urllib.parse import urlparse

class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''

    def __init__(self, test_config: str, country_code='che') -> None:
        
        self.test_config_valid(test_config)
        self.test_config = json.loads(test_config)
        self.tracks_directory = Path('test_definitions', country_code, 'aircraft_states')
        self.verify_tracks_directory(self.tracks_directory)
        self.flight_tracks = self.load_flight_tracks(self.tracks_directory)
            
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

    def make_json_compatible(self, struct: Any) -> Any:
        if isinstance(struct, tuple) and hasattr(struct, '_asdict'):
            return {k: self.make_json_compatible(v) for k, v in struct._asdict().items()}
        elif isinstance(struct, dict):
            return {k: self.make_json_compatible(v) for k, v in struct.items()}
        elif isinstance(struct, str):
            return struct
        try:
            return [self.make_json_compatible(v) for v in struct]
        except TypeError:
            return struct

    def build_test_payload(self): 
        ''' This is the main method to process the test configuration and build RID payload object, maxium of one flight is allocated to each USS. '''
        
        usses = self.test_config['usses']

        all_test_payloads = []
        
        for uss_index, uss in enumerate(usses):
            requested_flights = []
            flight_track_path = Path(self.tracks_directory, self.flight_tracks[uss_index])
            with open(flight_track_path) as generated_rid_state:
                rid_state_data = json.load(generated_rid_state)

            
            effective_after = rid_state_data['reference_time']
            flight_details =  rid_state_data['flight_details']
            operator_details = rid_state_data['operator_details']
            
            operator_location = OperatorLocation(lat = operator_details['location']['latitude'], lng = operator_details['location']['longitude'])
            operator_id = str(uuid.uuid4())

            rid_flight_details = RIDFlightDetails(operator_id = operator_id, operator_location = operator_location, operation_description = flight_details['operation_description'] , serial_number = flight_details['serial_number'], registration_number = flight_details['registration_number'])

            test_flight_details = TestFlightDetails(effective_after= effective_after,details = rid_flight_details)
            test_flight = TestFlight(injection_id = str(uuid.uuid4()), telemetry= rid_state_data['flight_telemetry'], details_respones= test_flight_details)            
            test_flight_deserialized = self.make_json_compatible(test_flight)
            requested_flights.append(test_flight_deserialized)
            test_payload = {'test_id': str(uuid.uuid4()), "requested_flights": requested_flights}        

            all_test_payloads.append({'injection_url':uss['injection_url'], 'injection_payload': test_payload})        
        
        return all_test_payloads


class TestHarness():
    ''' A class to submit Aircraft RID State to the USS test endpoint '''

    def __init__(self, auth_url, auth_spec, auth_sub):
        self.auth_url = auth_url
        
    
    def get_dss_session(self, auth_spec):
        ''' This method gets a DSS session using the monitoring tools that are provided in the DSS monitoring repository'''

        auth_adapter = make_auth_adapter(auth_spec)
        s = DSSTestSession(self.auth_url, auth_adapter)

        return s


    def submit_test(self, test_payload):        
        ''' This method submits the payload to the injection url by creating a DSSTestSession and then using that session to send the payload '''

        injection_url = test_payload['injection_url']        
        auth_spec = test_payload['auth_spec']
        auth_sub = urlparse(injection_url).netloc
        auth_spec_with_sub = auth_spec # TODO: Modify auth_spec with sub information
        dss_session = self.get_dss_session(auth_spec= auth_spec, auth_url= self.auth_url, auth_sub= auth_sub)

        response = dss_session.put(injection_url, data=test_payload['injection_payload'])

        if response.status_code == 200:
            print("New test with ID %s created" % test_payload['injection_payload']['test_id'])
        elif response.status_code ==409:
            print("Test already with ID %s already exists" % test_payload['injection_payload']['test_id'])
        else: 
            print(response.json())

