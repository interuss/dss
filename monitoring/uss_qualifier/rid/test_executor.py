import json
import os
import uuid
from pathlib import Path
from typing import List

from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.uss_qualifier.rid import display_data_evaluator, reports, aircraft_state_replayer
from monitoring.uss_qualifier.rid.aircraft_state_replayer import TestHarness, TestBuilder
from monitoring.uss_qualifier.rid.simulator import flight_state
from monitoring.uss_qualifier.rid.utils import RIDQualifierTestConfiguration, InjectedFlight, FullFlightRecord
from monitoring.uss_qualifier.utils import is_url


def load_rid_test_definitions(locale: str):
  test_definitions_path = Path(os.getcwd(), 'rid/test_definitions')
  aircraft_states_directory = Path(test_definitions_path, locale, 'aircraft_states')
  try:
    flight_records = aircraft_state_replayer.get_full_flight_records(aircraft_states_directory)
  except ValueError:
    print('[RID] No aircraft state files found in {}; generating them via simulator now'.format(aircraft_states_directory))
    flight_state.generate_aircraft_states(test_definitions_path)
    flight_records = aircraft_state_replayer.get_full_flight_records(aircraft_states_directory)
  return flight_records

def validate_configuration(test_configuration: RIDQualifierTestConfiguration):
  try:
    for injection_target in test_configuration.injection_targets:
      is_url(injection_target.injection_base_url)
  except ValueError:
    raise ValueError("A valid url for injection_target must be passed")


def run_rid_tests(test_configuration: RIDQualifierTestConfiguration,
                  auth_spec: str,
                  flight_records: List[FullFlightRecord]) -> reports.Report:
    my_test_builder = TestBuilder(test_configuration=test_configuration, flight_records=flight_records)
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
    with open('report_rid.json', 'w') as f:            
        json.dump(report, f)
    return report
