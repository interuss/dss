from pathlib import Path
import json
from typing import Dict
from monitoring.rid_qualifier.aircraft_state_replayer import TestHarness, TestBuilder
import asyncio
import arrow
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration

def build_test_configuration(locale: str, auth_spec:str, injection_url:str, allocated_track = 1) -> Dict: 
    now = arrow.now()
    test_start_time = now.shift(minutes=3) # Start the test three minutes from the time the test_exceutor is run. 
    test_configuration = {
        "locale": locale, # The locale here is indicating the geographical location in ISO3166 3-letter country code and also a folder within the test definitions directory. The aircraft_state_replayer reads flight track information from the locale/aircraft_states directory.  The locale directory also contains information about the query_bboxes that the rid display provider will use to query and retrieve the flight information. 
        "now": now.isoformat(),
        "test_start_time": test_start_time.isoformat(),
        "auth_spec": auth_spec,
        "usses":[
            {
                "injection_url": injection_url,
                "allocated_flight_track_number": allocated_track,  # The track that will be allocated to the uss
            }
        ]
    }

    test_config = ImplicitDict.parse(test_configuration, RIDQualifierTestConfiguration)
    print(test_config)
    return test_configuration
    
async def main(test_configuration:dict):
    # This is the configuration for the test.
    my_test_builder = TestBuilder(test_config=json.dumps(test_configuration), country_code=test_configuration['locale'])
    test_payloads = my_test_builder.build_test_payload()

    my_test_harness = TestHarness(auth_spec=test_configuration['auth_spec'], injection_url = test_configuration['usses'][0]['injection_url'])
    await my_test_harness.submit_payload_async(test_payloads=test_payloads)
    # TODO: call display data evaluator to read RID system state and compare to expectations
    
if __name__ == '__main__':
    
    test_configuration = build_test_configuration(locale='che', injection_url="https://dss.unmanned.corp", auth_spec="DummyOAuth(http://localhost:8085/token, sub=uss1)")
    asyncio.run(main(test_configuration=test_configuration))