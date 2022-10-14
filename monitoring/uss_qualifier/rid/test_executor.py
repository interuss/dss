import json
import os
import uuid
from pathlib import Path
from typing import List

from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import UTMClientSession
from monitoring.uss_qualifier.resources import ResourceCollection
from monitoring.uss_qualifier.resources.netrid import NetRIDServiceProviders
from monitoring.uss_qualifier.rid import (
    display_data_evaluator,
    reports,
    aircraft_state_replayer,
)
from monitoring.uss_qualifier.rid.aircraft_state_replayer import (
    TestBuilder,
)
from monitoring.uss_qualifier.rid.simulator import flight_state
from monitoring.uss_qualifier.rid.utils import (
    RIDQualifierTestConfiguration,
    InjectedFlight,
    FullFlightRecord,
)
from monitoring.uss_qualifier.utils import is_url


def load_rid_test_definitions(locale: str):
    test_definitions_path = Path(os.getcwd(), "rid/test_definitions")
    aircraft_states_directory = Path(test_definitions_path, locale, "aircraft_states")
    try:
        flight_records = aircraft_state_replayer.get_full_flight_records(
            aircraft_states_directory
        )
    except ValueError:
        print(
            "[RID] No aircraft state files found in {}; generating them via simulator now".format(
                aircraft_states_directory
            )
        )
        flight_state.generate_aircraft_states(test_definitions_path)
        flight_records = aircraft_state_replayer.get_full_flight_records(
            aircraft_states_directory
        )
    return flight_records


def run_rid_tests(
    resources: ResourceCollection,
    test_configuration: RIDQualifierTestConfiguration,
    auth_spec: str,
    flight_records: List[FullFlightRecord],
) -> reports.Report:
    my_test_builder = TestBuilder(
        test_configuration=test_configuration, flight_records=flight_records
    )
    test_payloads = my_test_builder.build_test_payloads()
    test_id = str(uuid.uuid4())
    report = reports.Report(setup=reports.Setup(configuration=test_configuration))

    # Inject flights into all USSs
    # TODO: Replace magic string 'netrid_service_providers' with dependency explicitly declared by the test scenario/case/step
    injection_targets: NetRIDServiceProviders = resources["netrid_service_providers"]
    injected_flights = []
    for i, target in enumerate(injection_targets.service_providers):
        injections = target.submit_test(test_payloads[i], test_id, report.setup)
        for flight in injections:
            injected_flights.append(InjectedFlight(uss=target.config, flight=flight))

    # Create observers
    observers: List[display_data_evaluator.RIDSystemObserver] = []
    for observer_config in test_configuration.observers:
        observer = display_data_evaluator.RIDSystemObserver(
            observer_config.name,
            UTMClientSession(
                observer_config.observation_base_url, make_auth_adapter(auth_spec)
            ),
            test_configuration.rid_version,
        )
        observers.append(observer)

    # Evaluate observed RID system states
    evaluator = display_data_evaluator.RIDObservationEvaluator(
        report.findings,
        injected_flights,
        test_configuration.evaluation,
        test_configuration.rid_version,
    )
    evaluator.evaluate_system(observers)
    with open("report_rid.json", "w") as f:
        json.dump(report, f)
    return report
