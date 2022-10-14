import json
import uuid

from monitoring.monitorlib.rid_automated_testing.injection_api import (
    CreateTestParameters,
)
from monitoring.uss_qualifier.resources import ResourceCollection
from monitoring.uss_qualifier.resources.netrid import (
    NetRIDServiceProviders,
    NetRIDObserversResource,
    FlightDataResource,
)
from monitoring.uss_qualifier.rid import (
    display_data_evaluator,
    reports,
)
from monitoring.uss_qualifier.rid.utils import (
    RIDQualifierTestConfiguration,
    InjectedFlight,
)


def run_rid_tests(
    resources: ResourceCollection,
    test_configuration: RIDQualifierTestConfiguration,
) -> reports.Report:
    test_id = str(uuid.uuid4())
    report = reports.Report(setup=reports.Setup(configuration=test_configuration))

    # Inject flights into all USSs
    # TODO: Replace magic string 'netrid_flights_data' with dependency explicitly declared by the test scenario/case/step
    flights_data: FlightDataResource = resources["netrid_flights_data"]
    test_flights = flights_data.get_test_flights()

    # TODO: Replace magic string 'netrid_service_providers' with dependency explicitly declared by the test scenario/case/step
    injection_targets: NetRIDServiceProviders = resources["netrid_service_providers"]
    service_providers = injection_targets.service_providers
    if len(service_providers) > len(test_flights):
        raise ValueError(
            "{} service providers were specified, but data for only {} test flights were provided".format(
                len(service_providers), len(test_flights)
            )
        )

    injected_flights = []
    for i, target in enumerate(service_providers):
        p = CreateTestParameters(requested_flights=[test_flights[i]])
        injections = target.submit_test(p, test_id, report.setup)
        for flight in injections:
            injected_flights.append(InjectedFlight(uss=target.config, flight=flight))

    # Create observers
    # TODO: Replace magic string 'netrid_observers' with dependency explicitly declared by the test scenario/case/step
    observers: NetRIDObserversResource = resources["netrid_observers"]

    # Evaluate observed RID system states
    evaluator = display_data_evaluator.RIDObservationEvaluator(
        report.findings,
        injected_flights,
        # TODO: Replace magic string 'netrid_observation_evaluation_configuration' with dependency explicitly declared by the test scenario/case/step
        resources["netrid_observation_evaluation_configuration"].configuration,
        test_configuration.rid_version,
    )
    evaluator.evaluate_system(observers.observers)
    with open("report_rid.json", "w") as f:
        json.dump(report, f)
    return report
