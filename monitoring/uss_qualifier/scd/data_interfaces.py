from typing import Optional
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest
import datetime

class Issue(ImplicitDict):
    ''' A class to hold a message that the test executor can provide to the USS in cases when the USS provides a response that is not same as the expected result. '''
    timestamp: Optional[str]

    subject: Optional[str]
    ''' Identifier of the subject of this issue, if applicable. This may be a UAS serial number, or any field of other object central to the issue. '''

    summary: str
    '''Human-readable summary of the issue'''

    details: str
    '''Human-readable description of the issue'''

    def __init__(self, **kwargs):
        super(Issue, self).__init__(**kwargs)
        if 'timestamp' not in kwargs:
            self.timestamp = datetime.datetime.utcnow().isoformat()

class RequiredResult(ImplicitDict):
    ''' A class to evaluate results / response to an injection of test flight data (TestFlightRequest) '''
    expected_result: str
    issue_details: Optional[Issue]
           

class TestInjectionRequiredResult(ImplicitDict):
    test_injection: InjectFlightRequest
    required_result: RequiredResult