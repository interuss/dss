import uuid

from implicitdict import ImplicitDict

from monitoring.monitorlib.rid_automated_testing.injection_api import (
    CreateTestParameters,
    ChangeTestResponse,
)
from monitoring.monitorlib.rid_common import RIDVersion
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.netrid import (
    FlightDataResource,
    NetRIDServiceProviders,
    NetRIDObserversResource,
    EvaluationConfigurationResource,
)
from monitoring.uss_qualifier.scenarios import TestScenario
from monitoring.uss_qualifier.scenarios.astm.netrid import display_data_evaluator
from monitoring.uss_qualifier.scenarios.astm.netrid.injection import InjectedFlight


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
        super().__init__()
        self._flights_data = flights_data
        self._service_providers = service_providers
        self._observers = observers
        self._evaluation_configuration = evaluation_configuration

    def run(self):
        self.begin_test_scenario()
        self.begin_test_case("Nominal flight")

        # Inject flights into all USSs
        self.begin_test_step("Injection")
        test_id = str(uuid.uuid4())
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
            query = target.submit_test(p, test_id)
            self.record_query(query)
            try:
                if query.status_code != 200:
                    raise ValueError(
                        f"Expected response code 200 but received {query.status_code} instead"
                    )
                if "json" not in query.response:
                    raise ValueError(f"Response did not contain a JSON body")
                changed_test: ChangeTestResponse = ImplicitDict.parse(
                    query.response["json"], ChangeTestResponse
                )
                injections = changed_test.injected_flights
            except ValueError as e:
                self.record_failed_check(
                    name="Successful injection",
                    summary="Error while trying to inject test flight",
                    severity=Severity.Critical,
                    relevant_participants=[target.participant_id],
                    details=f"While trying to inject a test flight into {target.participant_id}, encountered status code {query.status_code}: {str(e)}",
                )
                return
            # TODO: Validate injected flights, especially to make sure they contain the specified injection IDs
            for flight in injections:
                injected_flights.append(
                    InjectedFlight(
                        uss_participant_id=target.participant_id,
                        flight=flight,
                        query_timestamp=query.request.timestamp,
                    )
                )
        self.end_test_step()  # Injection

        # Evaluate observed RID system states
        self.begin_test_step("Polling")
        evaluator = display_data_evaluator.RIDObservationEvaluator(
            self,
            injected_flights,
            self._evaluation_configuration.configuration,
            # TODO: Replace hardcoded value
            RIDVersion.f3411_19,
        )
        evaluator.evaluate_system(self._observers.observers)
        self.end_test_step()  # Polling

        self.end_test_case()  # Nominal flight
        self.end_test_scenario()
