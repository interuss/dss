from typing import Optional, List, Dict
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest
from monitoring.uss_qualifier.common_data_definitions import Severity

### Begin Definitions around users and test executor 

class Result(ImplicitDict):
    ''' A class to hold the expected output from data processing and the test interface of the USS under test '''
    processing_outcome: str
        
class KnownIssueFields(ImplicitDict):
    ''' A class to hold a message that the test executor can provide to the USS in cases when the USS provides a response that is not same as the expected result. '''    
    test_code: str
    '''Code corresponding to check generating this issue'''

    relevant_requirements: List[str] = []
    '''Requirements that this issue relates to'''

    severity: Severity
    '''How severe the issue is'''

    subject: Optional[str]
    ''' Identifier of the subject of this issue, if applicable. This may be a UAS serial number, or any field of other object central to the issue. '''

    summary: str
    '''Human-readable summary of the issue'''

    details: str
    '''Human-readable description of the issue'''

class RequiredResult(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    acceptable_results: List[str] # Holds information about the "result" response provided by the test interface of the USS under test. Per the SCD Testing API this should be one of "Planned", "Rejected", "ConflictWithFlight" or "Failed"
    incorrect_result_details: Optional[Dict[Result, KnownIssueFields]] # In the case where the USS under test provides a result that is not the same as expected result then the test driver can provide a message detailing why the expected result was populated the way it was. This should help the USS under test debug specific parts of their internal systems. 

### End of definitions aronud users and test executor

### Definitions around interface between scd data generation and test executor 
  
class InjectionTarget(ImplicitDict):
    ''' A class to hold details of the USS under test '''
    uss_sequence_id: int
    name: str

class SCDTestConfiguration(ImplicitDict):
    ''' A class to hold test configuration, injection target data  '''
    injection_targets: List[InjectionTarget]

class FlightInjectionAttempt(ImplicitDict):
    ''' A class to hold details of the test injection and the expected result of processing the flight request '''
    test_injection: InjectFlightRequest
    required_result: RequiredResult

class TestDetails(ImplicitDict):
    ''' A class to hold details of different test structures '''
    name: str # e.g. a Nominal test or Flight Authorisation test
    response_wait_time_secs: int # amount of time the test driver should wait before expecting a response

class AutomatedTest(ImplicitDict):
    ''' A class to hold injection attempts and test definitions '''
    description: TestDetails # Details about the type of test it is
    injection_attempts: List[FlightInjectionAttempt]

###  End of Definitions around interface between scd data generation and test executor 
