import aircraft_state_replayer
from urllib.parse import urlparse
import uuid
import json
from typing import List, NamedTuple, Any
import arrow
import datetime
from datetime import datetime, timedelta

# This is the configuration for the test.
test_configuration = {
    "locale": "che",
    "utm_sps": [
        {
            "name": "Unmanned Systems Corp.",
            "injection_base_url": "https://unmanned.systems/tests/",
            "flight_details": [
                {
                    "serial_number": "C1A10C76-22D9-44E7",
                    "registration_number": "CHE87astrdge12k8",
                    "operation_description": "Electricity Grid Inspection"
                }
            ],
            "operator_details": [
                {
                    "name": "Electricty Inspection Company",
                    "location": {
                        "latitude": 46.974432835242055,
                        "longitude": 7.472983002662658
                    }
                }
            ]
        }
    ]
}


class UTMSP(NamedTuple):

    ''' This is the object that stores details of a UTMSP, mainly it will hold the injection endpoint and details of the flights allocated to the UTMSP and their submissiion status '''

    test_id: str
    name: str
    flight_id: int
    rid_state_injection_url: str
    rid_state_submission_status: bool


class OperatorLocation(NamedTuple):
    ''' A object to hold location of the operator when submitting flight data to UTMSP '''
    lat: float
    lng: float


class Operator(NamedTuple):
    ''' A object to hold details of a operator while querying Remote ID for testing purposes '''
    id: str
    location: OperatorLocation
    operation_description: str
    serial_number: str
    registration_number: str


class AircraftState(NamedTuple):
    ''' A object to hold Aircraft state details for remote ID purposes. For more information see the published standard API specification at https://github.com/uastech/standards/blob/36e7ea23a010ff91053f82ac4f6a9bfc698503f9/remoteid/canonical.yaml#L1604 '''
    
    timestamp: datetime 
    operational_status: str 
    position: AircraftPosition # See the definition above 
    height: AircraftHeight # See the definition above 
    track: float 
    speed: float 
    speed_accuracy: str 
    vertical_speed: float 

class RIDFlightDetails(NamedTuple):
    ''' A object to hold RID details of a flight ''' 
    pass

class TestFlightDetails(NamedTuple):
    ''' A object to hold details of a test flight ''' 
    effective_after: datetime
    details: 


class TestFlight(NamedTuple):
    ''' A object to hold TestFlight object ''' 

    injection_id: str    
    telemetry: List[AircraftState]
    details_respones : List[TestFlightDetails]    


class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''

    def __init__(self, test_config: Any) -> None:

        self.test_config = test_config
        self.test_config_valid()
    
    def test_config_valid(self) -> None:

        ''' This method checks if the test definition is a valid JSON ''' #TODO : Have a comprehensive way to check JSON definition
        if json.loads(self.test_config):
            pass
        else:
            raise ValueError("A valid JSON object must be submitted ")


    def build_test_payload(self) -> None: 
        ''' This is the main method to process the test configuration and build RID payload object '''
        utm_sps = self.test_config['utm_sps']
        test_id = str(uuid.uuid4())
        requested_flights = []
        for utm_sp in utm_sps:


        pass


class TestSubmitter():
    ''' A class to submit the test data to the UTMSP end point '''

    pass


class RIDDataValidator():
    ''' A class to check the output from UTMSP in Flight Blender and produce a text report and write on disk '''
    pass


if __name__ == '__main__':
    pass
