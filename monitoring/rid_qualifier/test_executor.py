from pathlib import Path
import json
from typing import List
from test_executor import TestHarness, TestBuilder


class TestSubmitter():
    ''' A class to submit the test data to USS end point '''

    def __init__(self, test_payloads):
        self.test_payload_valid(test_payloads)
        self.submit_payload(test_payloads)


    def test_payload_valid(self, test_payloads: List) -> None:
        ''' This method checks if the test definition is a valid JSON ''' #TODO : Have a comprehensive way to check JSON definition
        if len(test_payloads):
            pass
        else:
            raise ValueError("A valid payload object with atleast one flight / USS must be submitted")

    def submit_payload(self, test_payloads: List) -> None:
        ''' This method submits the payload to indvidual USS '''
        my_test_harness = TestHarness()
        for payload in test_payloads:   
            my_test_harness.submit_test(payload)
        

if __name__ == '__main__':
        
    # This is the configuration for the test.
    test_configuration = {
        "locale": "che",
        "auth_spec":'DummyOAuth("https://dss.com/', sub='+ auth_sub +')'
    }
    my_test_builder = TestBuilder(test_config = json.dumps(test_configuration), country_code='CHE')    
    test_payloads = my_test_builder.build_test_payload()

    my_test_harness = TestHarness(auth = test_configuration['auth_url'])
    my_test_submitter = my_test_harness.submit_test(test_payloads= test_payloads)
