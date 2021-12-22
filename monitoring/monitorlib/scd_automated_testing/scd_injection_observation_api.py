import enum
from monitoring.monitorlib.typing import ImplicitDict
from typing import List, Literal
from monitoring.monitorlib.scd import Volume4D

SCOPE_SCD_QUALIFIER_INJECT = 'utm.inject_test_data'

class InjectionStatus(str, enum.Enum):
    ''' A enum to hold results of flight processing as defined by the SCD test API '''
    Planned = 'Planned'
    Rejected = 'Rejected'
    ConflictWithFlight = 'ConflictWithFlight'
    Failed = 'Failed'

class OperationalIntentTestInjection(ImplicitDict):
    ''' A class to hold data for operational intent data that will be submitted to the SCD testing interface. '''
    state: str
    priorty: int = 0
    volumes: List[Volume4D]
    off_nominal_volumes: List[Volume4D]= []

## Definitions around flight authorisation data that need to be submitted to the test injection interface
   
class FlightAuthorisationData(ImplicitDict):
    '''A class to hold information about Flight Authorisation Test, for more information see https://github.com/interuss/dss/blob/master/interfaces/automated-testing/scd/scd.yaml#L317'''
    
    uas_serial_number: str
    operation_mode: str
    operation_category: str
    uas_class: str
    identification_technologies: List[str]
    uas_type_certificate: str
    connectivity_methods: List[str]
    endurance_minutes: int
    emergency_procedure_url: str
    operator_id: str
    uas_id: str    

class InjectFlightRequest(ImplicitDict):
    ''' A class to hold the details of a test injection payload '''
    flight_id: str
    operation_intent: OperationalIntentTestInjection
    flight_authorisation: FlightAuthorisationData