import json
import uuid
from typing import List, Optional
from monitoring.uss_qualifier.rid.aircraft_state_replayer import TestHarness, TestBuilder
from monitoring.uss_qualifier.utils import RIDQualifierTestConfiguration, InjectedFlight
from monitoring.uss_qualifier.rid import display_data_evaluator, reports
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.auth import make_auth_adapter


def main(
        test_configuration: RIDQualifierTestConfiguration,
        auth_spec: str, aircraft_state_files: Optional[list] = None) -> reports.Report:
    my_test_builder = TestBuilder(test_configuration=test_configuration, aircraft_state_files=aircraft_state_files)
    test_payloads = my_test_builder.build_test_payloads()
    test_id = str(uuid.uuid4())
    report = reports.Report(setup=reports.Setup(configuration=test_configuration))

    # Inject flights into all USSs
    injected_flights = []
    for i, target in enumerate(test_configuration.injection_targets):
      uss_injection_harness = TestHarness(
        auth_spec=auth_spec,
        injection_base_url=target.injection_base_url)
      uss_injection_harness.submit_test(test_payloads[i], test_id, report.setup)
      for flight in test_payloads[i].requested_flights:
        injected_flights.append(InjectedFlight(uss=target, flight=flight))

    # Create observers
    observers: List[display_data_evaluator.RIDSystemObserver] = []
    for observer_config in test_configuration.observers:
        observer = display_data_evaluator.RIDSystemObserver(
            observer_config.name, DSSTestSession(
                observer_config.observation_base_url,
                make_auth_adapter(auth_spec)))
        observers.append(observer)

    # Evaluate observed RID system states
    display_data_evaluator.evaluate_system(
        injected_flights, observers, test_configuration.evaluation,
        report.findings)
    with open('../report.json', 'w') as f:
        json.dump(report, f)
    return report
