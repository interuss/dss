from monitoring.rid_qualifier.aircraft_state_replayer import TestHarness, TestBuilder
import asyncio
import arrow
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration, RIDQualifierUSSConfig

def build_uss_config(injection_base_url:str, allocated_track:int = 0) -> RIDQualifierUSSConfig:
  return RIDQualifierUSSConfig(injection_base_url=injection_base_url, allocated_flight_track_number = allocated_track)
  

def build_test_configuration(locale: str, auth_spec:str, uss_config:RIDQualifierUSSConfig) -> RIDQualifierTestConfiguration: 
    now = arrow.now()
    test_start_time = now.shift(minutes=3) # Start the test three minutes from the time the test_exceutor is run. 
    
    test_config = RIDQualifierTestConfiguration(
      locale = locale,
      now = now.isoformat(),
      test_start_time = test_start_time.isoformat(),
      auth_spec = auth_spec,
      usses = [uss_config]
    )

    return test_config
    
async def main(test_configuration: RIDQualifierTestConfiguration):
    # This is the configuration for the test.
    my_test_builder = TestBuilder(test_configuration = test_configuration)
    test_payloads = my_test_builder.build_test_payloads()
    
    my_test_harness = TestHarness(auth_spec = test_configuration.auth_spec, injection_base_url = test_configuration.usses[0].injection_base_url)
    
    my_test_harness.submit_payloads_async(test_payloads=test_payloads)
    # TODO: call display data evaluator to read RID system state and compare to expectations



if __name__ == '__main__':
    uss_config = build_uss_config( injection_base_url="https://dss.unmanned.corp", allocated_track=0)
    test_configuration = build_test_configuration(locale='che', auth_spec="DummyOAuth(http://localhost:8085/token, sub=uss1)",uss_config = uss_config)
    
    asyncio.run(main(test_configuration=test_configuration))