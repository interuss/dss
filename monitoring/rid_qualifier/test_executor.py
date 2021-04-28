from pathlib import Path
import json
from typing import List
from aircraft_state_replayer import TestHarness, TestBuilder
import asyncio
class TestSubmitter():
    ''' A class to submit the test data to USS end point '''

    def __init__(self, test_payloads):
        self.test_payload_valid(test_payloads)
        self.submit_payload(test_payloads)

    def test_payload_valid(self, test_payloads: List) -> None:
        ''' This method checks if the test definition is a valid JSON '''  # TODO : Have a comprehensive way to check JSON definition
        if len(test_payloads):
            pass
        else:
            raise ValueError(
                "A valid payload object with atleast one flight / USS must be submitted")

    def submit_payload(self, test_payloads: List) -> None:
        ''' This method submits the payload to indvidual USS '''
        my_test_harness = TestHarness()
        for payload in test_payloads:
            my_test_harness.submit_test(payload)



async def main():
    # This is the configuration for the test.
    test_configuration = {
        "locale": "che", # The locale here is indicating the geographical location in ISO3166 3-letter country code and also a folder within the test definitions directory. The aircraft_state_replayer reads flight track information from the locale/aircraft_states directory.  The locale directory also contains information about the query_bboxes that the rid display provider will use to query and retrieve the flight information. 
        'auth_url':'http://localhost:8085/token',
        "auth_spec": "DummyOAuth(http://localhost:8085/token, sub=fake_uss)",
        "usses":[
            {
                "name": "Unmanned Systems Corp.",
                "injection_url": "https://dss.unmanned.corp/tests/",
                "allocated_flight_track_number": 1,  # The track that will be allocated to the uss
                "start_time_from_now_secs": 180 # Tell USS to start injection in 3 minutes
            }
        ]
    }
    my_test_builder = TestBuilder(test_config=json.dumps(
        test_configuration), country_code='CHE')
    test_payloads = my_test_builder.build_test_payload()

    my_test_harness = TestHarness(auth_spec=test_configuration['auth_spec'], auth_url= test_configuration['auth_url'])
    await my_test_harness.submit_payload_async(test_payloads=test_payloads)
    
if __name__ == '__main__':
    asyncio.run(main())