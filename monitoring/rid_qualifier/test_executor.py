from pathlib import Path
import json
from typing import List
from monitorting.aircraft_state_replayer import TestHarness, TestBuilder
import asyncio


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