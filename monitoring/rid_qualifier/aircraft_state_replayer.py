from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import DSSTestSession
import asyncio
from monitoring.monitorlib import rid
import json, os
import uuid
from pathlib import Path
from typing import  Any
from monitoring.rid_qualifier.utils import OperatorLocation, RIDFlightDetails, TestFlightDetails, TestFlight
import arrow
import time
from typing import List

class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''

    def __init__(self, test_config: str, country_code='che') -> None:
        
        self.test_config_valid(test_config)
        self.test_config = json.loads(test_config)
        self.tracks_directory = Path('test_definitions', country_code, 'aircraft_states')
        self.verify_tracks_directory(self.tracks_directory)
        self.flight_tracks = self.load_flight_tracks(self.tracks_directory)
            
    def load_flight_tracks(self, tracks_directory) -> List[str]:
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
        
        test_reference_time = arrow.get(self.test_config['now'])
        test_start_time = arrow.get(self.test_config['test_start_time'])
        test_start_offset = test_start_time.shift(minutes =1)
        test_start_offset_isoformat = test_start_offset.isoformat()

        for uss_index, uss in enumerate(usses):
            requested_flights = []
            flight_track_path = Path(self.tracks_directory, self.flight_tracks[uss['allocated_flight_track_number']])
            with open(flight_track_path) as generated_rid_state:
                disk_rid_state_data = json.load(generated_rid_state)
            
            disk_rid_state_data['reference_time'] = test_reference_time.isoformat()

            updated_timestamps_telemetry = []
            timestamp = test_start_offset.shift(seconds = 1)
            for telemetry_id, flight_telemetry in enumerate(disk_rid_state_data['flight_telemetry']['states']):
                
                test_start_offset.shift(seconds =1) 
                
                flight_telemetry['timestamp'] = timestamp.isoformat()
                updated_timestamps_telemetry.append(flight_telemetry)
            
            
            flight_details =  disk_rid_state_data['flight_details']
            operator_details = disk_rid_state_data['operator_details']
            
            operator_location = OperatorLocation(lat = operator_details['location']['latitude'], lng = operator_details['location']['longitude'])
            operator_id = str(uuid.uuid4())

            rid_flight_details = RIDFlightDetails(operator_id = operator_id, operator_location = operator_location, operation_description = flight_details['operation_description'] , serial_number = flight_details['serial_number'], registration_number = flight_details['registration_number'])

            test_flight_details = TestFlightDetails(effective_after= test_start_offset_isoformat,details = rid_flight_details)
            test_flight = TestFlight(injection_id = str(uuid.uuid4()), telemetry= updated_timestamps_telemetry, details_responses = test_flight_details)            
            test_flight_deserialized = self.make_json_compatible(test_flight)
            requested_flights.append(test_flight_deserialized)
            test_payload = {'test_id': str(uuid.uuid4()), "requested_flights": requested_flights}        

            all_test_payloads.append({'injection_url':uss['injection_url'], 'injection_payload': test_payload})        
        
        return all_test_payloads


class TestHarness():
    ''' A class to submit Aircraft RID State to the USS test endpoint '''

    def __init__(self, auth_spec:str, injection_url:str):
        self.auth_spec = auth_spec
        self.injection_url= injection_url
        
    def get_uss_session(self) -> DSSTestSession:
        ''' This method gets a DSS session using the monitoring tools that are provided in the DSS monitoring repository '''

        auth_adapter = make_auth_adapter(self.auth_spec)
        s = DSSTestSession(self.injection_url, auth_adapter)
    
        return s

    def submit_test(self,uss_session, test_payload, test_injection_url) -> None:
        
        response = uss_session.put(url = test_injection_url, data=test_payload['injection_payload'],scope = ' '.join([rid.SCOPE_RID_QUALIFIER_INJECT ]))

        if response.status_code == 200:
            print("New test with ID %s created" % test_payload['injection_payload']['test_id'])
        elif response.status_code ==409:
            raise RuntimeError("Test with ID %s already exists" % test_payload['injection_payload']['test_id'])
        elif response.status_code == 404:
            raise RuntimeError("Test with ID %s not submitted, the requested endpoint was not found on the server" % test_payload['injection_payload']['test_id'])
        elif response.status_code == 401:
            raise RuntimeError("Test with ID %s not submitted, the access token was not provided in the Authorization header, or the token could not be decoded, or token was invalid" % test_payload['injection_payload']['test_id'])
        elif response.status_code == 403:
            raise RuntimeError("Test with ID %s not submitted, the access token was decoded successfully but did not include the appropriate scope" % test_payload['injection_payload']['test_id'])
        elif response.status_code == 413:
            raise RuntimeError("Test with ID %s not submitted, the injection payload was too large" % test_payload['injection_payload']['test_id'])
        else:
            raise RuntimeError("Test with ID %(test_id)s not submitted, the server returned a HTTP error code %(error_code)d" % {'test_id':test_payload['injection_payload']['test_id'], 'error_code': response.error_code})


    async def submit_payload_async(self, test_payloads):        
        ''' This method submits the payload to the injection url by creating a DSSTestSession and then using that session to send the payload '''
        for test_payload in test_payloads:
            test_injection_url = self.injection_url  + '/tests/{test_id}'.format(test_id=test_payload['injection_payload']['test_id'])
                        
            
            uss_session = self.get_uss_session()
            
            
            
            self.submit_test(uss_session=uss_session, test_payload=test_payload,test_injection_url = test_injection_url)
