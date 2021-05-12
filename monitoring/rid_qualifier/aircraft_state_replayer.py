from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import DSSTestSession
import asyncio
from monitoring.monitorlib import rid
import json, os
import uuid
from pathlib import Path
from typing import  Any
from monitoring.rid_qualifier.utils import OperatorLocation, RIDFlightDetails, TestFlightDetails, TestFlight, TestPayload, DeliverablePayload
import arrow
from uuid import UUID
from typing import List
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration

class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''

    def __init__(self, test_configuration: RIDQualifierTestConfiguration) -> None:
        self.test_configuration = test_configuration

        self.aircraft_states_directory = Path('test_definitions', test_configuration.locale, 'aircraft_states')
        
        self.verify_aircraft_states_directory(self.aircraft_states_directory)
        self.aircraft_state_files = os.listdir(self.aircraft_states_directory) 
                
        self.usses = self.test_configuration.usses
        
        self.disk_rid_state_data =[]
        for uss_index, uss in enumerate(self.usses):
            flight_track_path = Path(self.aircraft_states_directory, self.aircraft_state_files[uss['allocated_flight_track_number']])
            
            with open(flight_track_path) as generated_rid_state:
                disk_rid_state_file = json.load(generated_rid_state)
                self.disk_rid_state_data.append(disk_rid_state_file)
        
            
    def verify_aircraft_states_directory(self, aircraft_states_directory) -> None:

        ''' This method checks if there are tracks in the tracks directory '''
        
        files = [f for f in os.listdir(aircraft_states_directory) if os.path.isfile(os.path.join(aircraft_states_directory, f))]
        if files:
            pass
        else:
            raise ValueError("The there are no tracks in the tracks directory, create tracks first using the flight_data_generator module. ")

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

    def build_test_payload(self) ->List[DeliverablePayload]: 
        ''' This is the main method to process the test configuration and build RID payload object, maxium of one flight is allocated to each USS. '''
        
        usses = self.test_configuration.usses # Store the USS details 

        all_test_payloads = [] # This holds the data that will be deilver
        
        test_reference_time = arrow.get(self.test_configuration.now)
        test_start_time = arrow.get(self.test_configuration.test_start_time)
        test_start_offset = test_start_time.shift(minutes =1)
        test_start_offset_isoformat = test_start_offset.isoformat()

        requested_flights = [] # This objects holds the modified aircraft state data, aircraft state data is read from disk and timestamps are modified to have recent ones. 
        for state_data_index, current_disk_rid_state_data in enumerate(self.disk_rid_state_data): 
            uss = usses[state_data_index]  

            if uss.injection_suffix:
                suffix_paths = uss.injection_suffix.split('/')
                try: 
                    # get the test ID from the injection suffix, per the injection API, it should be the last element in the path 
                    test_uuid = suffix_paths[-1]
                    is_uuid = UUID(test_uuid, version=4)
                except ValueError as ve: 
                    test_id = str(uuid.uuid4())
                    test_injection_path = '/tests/{test_id}'.format(test_id=test_id)
                else: 
                    test_id = test_uuid
                    test_injection_path = uss.injection_suffix

            else: 
                test_id = str(uuid.uuid4())
                test_injection_path = '/tests/{test_id}'.format(test_id=test_id)

            
            current_disk_rid_state_data['reference_time'] = test_reference_time.isoformat()
            updated_timestamps_telemetry = []            

            timestamp = test_start_offset.shift(seconds = 1)
            
            for telemetry_id, flight_telemetry in enumerate(current_disk_rid_state_data['flight_telemetry']['states']):

                timestamp = timestamp.shift(seconds =1)
                flight_telemetry['timestamp'] = timestamp.isoformat()
                updated_timestamps_telemetry.append(flight_telemetry)
                        
            flight_details =  current_disk_rid_state_data['flight_details']
            operator_details = current_disk_rid_state_data['operator_details']
            
            operator_location = OperatorLocation(lat = operator_details['location']['latitude'], lng = operator_details['location']['longitude'])            
            rid_flight_details = RIDFlightDetails(operator_id = str(uuid.uuid4()), operator_location = operator_location, operation_description = flight_details['operation_description'] , serial_number = flight_details['serial_number'], registration_number = flight_details['registration_number'])
            test_flight_details = TestFlightDetails(effective_after= test_start_offset_isoformat,details = rid_flight_details)
            test_flight = TestFlight(injection_id = str(uuid.uuid4()), telemetry= updated_timestamps_telemetry, details_responses = test_flight_details)
            test_flight_deserialized = self.make_json_compatible(test_flight)
            requested_flights.append(test_flight_deserialized)
            test_payload = TestPayload(test_id = test_id, requested_flights= requested_flights)
            test_payload_data_metadata = DeliverablePayload(injection_path= test_injection_path, injection_payload = test_payload)
            all_test_payloads.append(test_payload_data_metadata)
        
        return all_test_payloads


class TestHarness():
    ''' A class to submit Aircraft RID State to the USS test endpoint '''

    def __init__(self, auth_spec:str, injection_base_url:str):
        
        auth_adapter = make_auth_adapter(auth_spec)
        self.uss_session = DSSTestSession(injection_base_url, auth_adapter)

    def submit_test(self,uss_session:DSSTestSession, test_payload:DeliverablePayload, test_injection_path:str) -> None:
        
        response = uss_session.put(url = test_injection_path, data=json.dumps(test_payload.injection_payload), scope = ' '.join([rid.SCOPE_RID_QUALIFIER_INJECT ]))

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
            raise RuntimeError("Test with ID %(test_id)s not submitted, the server returned the following HTTP error code: %(error_code)d" % {'test_id':test_payload['injection_payload']['test_id'], 'error_code': response.error_code})


    async def submit_payload_async(self, test_payloads):
        ''' This method submits the payload to the injection url by creating a DSSTestSession and then using that session to send the payload '''
        for test_payload in test_payloads:            
            self.submit_test(uss_session=self.uss_session, test_payload=test_payload,test_injection_path = test_payload.injection_path)