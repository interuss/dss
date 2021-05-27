import json
import uuid
from monitoring.rid_qualifier.aircraft_state_replayer import TestHarness, TestBuilder
import arrow
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration, RIDQualifierUSSConfig, InjectedFlight
from monitoring.rid_qualifier import display_data_evaluator, reports
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.auth import make_auth_adapter


def build_uss_config(injection_base_url:str) -> RIDQualifierUSSConfig:
  return RIDQualifierUSSConfig(injection_base_url=injection_base_url, name='uss1')


def build_test_configuration(locale: str, auth_spec:str, uss_config:RIDQualifierUSSConfig) -> RIDQualifierTestConfiguration:
    now = arrow.now()
    test_start_time = now.shift(seconds=15)

    test_config = RIDQualifierTestConfiguration(
      locale = locale,
      now = now.isoformat(),
      test_start_time = test_start_time.isoformat(),
      auth_spec = auth_spec,
      usses = [uss_config]
    )

    return test_config

def main(test_configuration: RIDQualifierTestConfiguration, observation_base_url: str):
    # This is the configuration for the test.
    my_test_builder = TestBuilder(test_configuration = test_configuration)
    test_payloads = my_test_builder.build_test_payloads()
    test_id = str(uuid.uuid4())
    report = reports.Report()

    # Inject flights into all USSs
    injected_flights = []
    for i, uss in enumerate(test_configuration.usses):
      uss_injection_harness = TestHarness(
        auth_spec=test_configuration.auth_spec,
        injection_base_url=uss.injection_base_url)
      uss_injection_harness.submit_test(test_payloads[i], test_id, report.setup)
      for flight in test_payloads[i].requested_flights:
        injected_flights.append(InjectedFlight(uss=uss, flight=flight))

    # Evaluate observed RID system states
    config = display_data_evaluator.EvaluationConfiguration()
    #TODO: Accept user input describing desired observers
    observer = display_data_evaluator.RIDSystemObserver(
        'uss1', DSSTestSession(
        observation_base_url,
        make_auth_adapter('NoAuth()')))
    display_data_evaluator.evaluate_system(injected_flights, [observer], config, report.findings)
    with open('report.json', 'w') as f:
      json.dump(report, f)
