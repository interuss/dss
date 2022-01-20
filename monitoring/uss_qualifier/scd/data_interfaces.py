from typing import Optional, List, Dict
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest
from monitoring.uss_qualifier.common_data_definitions import Severity

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

class KnownResponses(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest). This data structure holds information about the "result" response provided by the test interface of the USS under test and details of how long the executor should wait for a response. Per the SCD Testing API an acceptable result should be one of "Planned", "Rejected", "ConflictWithFlight" or "Failed". In the case where the USS under test provides a result that is not the same as expected result then the test driver can provide a message detailing why the expected result was populated the way it was. This should help the USS under test debug specific parts of their internal systems. '''
    acceptable_results: List[str] 
    ''' Acceptable strings that the USS under test can report as the result of processing the test data. '''

    incorrect_result_details: Dict[str, KnownIssueFields]
    ''' In cases where the USS provides a response that is not in the acceptable results, the test exceutor may display a message to the user detailing why the reported response was not correct '''

class InjectionTarget(ImplicitDict):
    ''' A class to hold the role of the USS under test '''
    uss_role: str
    ''' In some testing scenarios, the USS may be assigned a role e.g. Querying USS based on the actions they will perform in the test scenario '''

class FlightInjectionAttempt(ImplicitDict):
    ''' A class to hold details of the test injection, the injection target and the expected result of processing the flight request '''
    test_injection: InjectFlightRequest
    ''' Data around flight authorisation and operational intents that are submitted to the test interface of the USS under test '''

    known_responses: KnownResponses
    ''' Details about what the USS under test should report after processing the test data '''

    injection_target: InjectionTarget
    ''' Details of the USS under test as mapped to the test type '''

class AutomatedTest(ImplicitDict):
    ''' A class to hold injection attempts and test definitions '''

    injection_attempts: List[FlightInjectionAttempt]
    ''' Details of attempts of submitting test data to the interface of USS under test '''
