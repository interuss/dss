from monitoring.rid_qualifier.aircraft_state_replayer import TestHarness, TestBuilder
import asyncio
import arrow
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration, RIDQualifierUSSConfig

import json

def build_test_configuration(locale: str, auth_spec:str, injection_base_url:str, injection_suffix:str=None, allocated_track = 0) -> RIDQualifierTestConfiguration: 
    now = arrow.now()
    test_start_time = now.shift(minutes=3) # Start the test three minutes from the time the test_exceutor is run. 
    
    test_config = RIDQualifierTestConfiguration(
      locale = locale,
      now = now.isoformat(),
      test_start_time = test_start_time.isoformat(),
      auth_spec = auth_spec,
      usses = [RIDQualifierUSSConfig(injection_base_url=injection_base_url, injection_suffix= injection_suffix, allocated_flight_track_number = allocated_track)]
    )

    return test_config
    
async def main(test_configuration: RIDQualifierTestConfiguration):
    # This is the configuration for the test.
    my_test_builder = TestBuilder(test_configuration = test_configuration)
    test_payloads = my_test_builder.build_test_payload()
    
    my_test_harness = TestHarness(auth_spec = test_configuration.auth_spec, injection_base_url = test_configuration.usses[0].injection_base_url)
    
    await my_test_harness.submit_payload_async(test_payloads=test_payloads)
    # TODO: call display data evaluator to read RID system state and compare to expectations

