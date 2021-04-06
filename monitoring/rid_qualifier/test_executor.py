import aircraft_state_replayer
from urllib.parse import urlparse
import uuid
from typing import NamedTuple
import aircraft_state_replayer

test_config = {'locale': 'che', 'utm_sp_num': 6,
               'utm_sp_details': [{'name':"Unmanned Systems", "injection_base_url": 'https://unmanned.systems/tests/'}]}

class UTMSP(NamedTuple):

    ''' This is the object that stores details of a UTMSP, mainly it will hold the injection endpoint and details of the flights allocated to the UTMSP and their submissiion status '''

    test_id: str
    injection_id : str
    name: str
    flight_id: int
    rid_state_injection_url: str
    rid_state_submission_status: bool


class OperatorLocation(NamedTuple):
    ''' A object to hold location of the operator when submitting flight data to UTMSP ''' 
    lat : float 
    lng : float 

class Operator(NamedTuple):
    ''' A object to hold details of a operator while querying Remote ID for testing purposes '''
    id: str
    location: OperatorLocation
    operation_description: str
    serial_number: str
    registration_number: str


class TestBuilder():
    ''' A class to setup the test data and create the objects ready to be submitted to the test harness '''
    pass

class TestSubmitter():
    ''' A class to submit the test data to the UTMSP end point '''

    pass

class RIDDataValidator():
    ''' A class to check the output from UTMSP in Flight Blender and produce a text report and write on disk '''
    pass


if __name__ == '__main__':
    pass
