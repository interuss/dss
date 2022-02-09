from enum import Enum
from monitoring.monitorlib.typing import ImplicitDict
from typing import List, Optional
from monitoring.monitorlib.scd import Volume4D

SCOPE_SCD_QUALIFIER_INJECT = 'utm.inject_test_data'

## Definitions around operational intent data that need to be submitted to the test injection interface
  
class OperationalIntentTestInjection(ImplicitDict):
    ''' A class to hold data for operational intent data that will be submitted to the SCD testing interface. '''
    state: str
    priority: int = 0
    volumes: List[Volume4D]
    off_nominal_volumes: List[Volume4D]= []

### End of definitions around operational intent data

## Definitions around flight authorisation data that need to be submitted to the test injection interface
   
class FlightAuthorisationData(ImplicitDict):
    '''A class to hold information about Flight Authorisation Test '''
    
    uas_serial_number: str
    operation_mode: str
    operation_category: Optional[str]
    uas_class: str
    identification_technologies: List[str]
    uas_type_certificate: Optional[str]
    connectivity_methods: List[str]
    endurance_minutes: int
    emergency_procedure_url: str
    operator_id: str
    uas_id: Optional[str]

### End of definitions around flight authorisation data

class InjectFlightRequest(ImplicitDict):
    ''' A class to hold the details of a test injection payload '''
    operational_intent: Optional[OperationalIntentTestInjection]
    flight_authorisation: Optional[FlightAuthorisationData]


class InjectFlightResult(str, Enum):
    Planned = 'Planned'
    Rejected = 'Rejected'
    ConflictWithFlight = 'ConflictWithFlight'
    Failed = 'Failed'


class InjectFlightResponse(ImplicitDict):
    ''' A class to hold test flight submission response '''
    result: InjectFlightResult
    notes: Optional[str]
    operational_intent_id: Optional[str]


class DeleteFlightResult(str, Enum):
    Closed = 'Closed'
    Failed = 'Failed'


class DeleteFlightResponse(ImplicitDict):
    result: DeleteFlightResult
    notes: Optional[str]
