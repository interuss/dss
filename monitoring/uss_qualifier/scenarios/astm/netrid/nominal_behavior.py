import json
import uuid

from monitoring.monitorlib.rid_automated_testing.injection_api import (
    CreateTestParameters,
)
from monitoring.monitorlib.rid_common import RIDVersion
from monitoring.uss_qualifier.resources.netrid import (
    FlightDataResource,
    NetRIDServiceProviders,
    NetRIDObserversResource,
    EvaluationConfigurationResource,
)
from monitoring.uss_qualifier.rid import reports
from monitoring.uss_qualifier.rid.utils import InjectedFlight
from monitoring.uss_qualifier.scenarios import TestScenario
from monitoring.uss_qualifier.scenarios.astm.netrid import display_data_evaluator


class NominalBehavior(TestScenario):
    _flights_data: FlightDataResource
    _service_providers: NetRIDServiceProviders
    _observers: NetRIDObserversResource
    _evaluation_configuration: EvaluationConfigurationResource

    def __init__(
        self,
        flights_data: FlightDataResource,
        service_providers: NetRIDServiceProviders,
        observers: NetRIDObserversResource,
        evaluation_configuration: EvaluationConfigurationResource,
    ):
        self._flights_data = flights_data
        self._service_providers = service_providers
        self._observers = observers
        self._evaluation_configuration = evaluation_configuration

    def run(self):
        test_id = str(uuid.uuid4())
        report = reports.Report()

        # Inject flights into all USSs
        test_flights = self._flights_data.get_test_flights()
        service_providers = self._service_providers.service_providers
        if len(service_providers) > len(test_flights):
            raise ValueError(
                "{} service providers were specified, but data for only {} test flights were provided".format(
                    len(service_providers), len(test_flights)
                )
            )
        injected_flights = []
        for i, target in enumerate(service_providers):
            p = CreateTestParameters(requested_flights=[test_flights[i]])
            injections = target.submit_test(p, test_id)
            for flight in injections:
                injected_flights.append(
                    InjectedFlight(uss=target.config, flight=flight)
                )

        # Evaluate observed RID system states
        evaluator = display_data_evaluator.RIDObservationEvaluator(
            report.findings,
            injected_flights,
            self._evaluation_configuration.configuration,
            # TODO: Replace hardcoded value
            RIDVersion.f3411_19,
        )
        evaluator.evaluate_system(self._observers.observers)

        # TODO: remove this write
        with open("report_rid.json", "w") as f:
            json.dump(report, f)

        return report
