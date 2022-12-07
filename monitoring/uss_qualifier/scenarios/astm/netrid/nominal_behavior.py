import traceback
import uuid
import time
from typing import List

import arrow
from implicitdict import ImplicitDict
from requests.exceptions import RequestException

from monitoring.monitorlib.rid_automated_testing.injection_api import (
    CreateTestParameters,
)
from monitoring.monitorlib.rid_common import RIDVersion
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.netrid import (
    FlightDataResource,
    NetRIDServiceProviders,
    NetRIDObserversResource,
    EvaluationConfigurationResource,
)
from monitoring.uss_qualifier.scenarios.astm.netrid.injected_flight_collection import (
    InjectedFlightCollection,
)
from monitoring.uss_qualifier.scenarios.astm.netrid.virtual_observer import (
    VirtualObserver,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from monitoring.uss_qualifier.scenarios.astm.netrid import display_data_evaluator
from monitoring.uss_qualifier.scenarios.astm.netrid.injection import InjectedFlight
from uas_standards.interuss.automated_testing.rid.v1.injection import ChangeTestResponse


class NominalBehavior(TestScenario):
    _flights_data: FlightDataResource
    _service_providers: NetRIDServiceProviders
    _observers: NetRIDObserversResource
    _evaluation_configuration: EvaluationConfigurationResource

    _injected_flights: List[InjectedFlight]

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
        self._injected_flights = []

    def run(self):
        self.begin_test_scenario()
        self.begin_test_case("Nominal flight")

        self.begin_test_step("Injection")
        if not self._inject_flights():
            return
        self.end_test_step()

        self.begin_test_step("Polling")
        if not self._poll_during_flights():
            return
        self.end_test_step()

        self.end_test_case()
        self.end_test_scenario()

    def _inject_flights(self) -> bool:
        # Inject flights into all USSs
        test_id = str(uuid.uuid4())
        test_flights = self._flights_data.get_test_flights()
        service_providers = self._service_providers.service_providers
        if len(service_providers) > len(test_flights):
            raise ValueError(
                "{} service providers were specified, but data for only {} test flights were provided".format(
                    len(service_providers), len(test_flights)
                )
            )
        for i, target in enumerate(service_providers):
            p = CreateTestParameters(requested_flights=[test_flights[i]])
            check = self.check("Successful injection", [target.participant_id])
            try:
                query = target.submit_test(p, test_id)
            except RequestException as e:
                stacktrace = "".join(
                    traceback.format_exception(
                        etype=type(e), value=e, tb=e.__traceback__
                    )
                )
                check.record_failed(
                    summary="Error while trying to inject test flight",
                    severity=Severity.High,
                    details=f"While trying to inject a test flight into {target.participant_id}, encountered error:\n{stacktrace}",
                )
                return False
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
                check.record_passed()
            except ValueError as e:
                check.record_failed(
                    summary="Error injecting test flight",
                    severity=Severity.High,
                    details=f"Attempting to inject a test flight into {target.participant_id}, encountered status code {query.status_code}: {str(e)}",
                )
                return False
            # TODO: Validate injected flights, especially to make sure they contain the specified injection IDs
            for flight in injections:
                self._injected_flights.append(
                    InjectedFlight(
                        uss_participant_id=target.participant_id,
                        test_id=test_id,
                        flight=flight,
                        query_timestamp=query.request.timestamp,
                    )
                )

        config = self._evaluation_configuration.configuration
        # TODO: Replace hardcoded value
        rid_version = RIDVersion.f3411_19
        self._virtual_observer = VirtualObserver(
            injected_flights=InjectedFlightCollection(self._injected_flights),
            repeat_query_rect_period=config.repeat_query_rect_period,
            min_query_diagonal_m=config.min_query_diagonal,
            relevant_past_data_period=rid_version.realtime_period
            + config.max_propagation_latency.timedelta,
        )

        return True

    def _poll_during_flights(self) -> bool:
        # Evaluate observed RID system states
        evaluator = display_data_evaluator.RIDObservationEvaluator(
            self,
            self._injected_flights,
            self._evaluation_configuration.configuration,
            # TODO: Replace hardcoded value
            RIDVersion.f3411_19,
        )

        t_end = self._virtual_observer.get_last_time_of_interest()
        t_now = arrow.utcnow()
        if t_now > t_end:
            raise RuntimeError(
                f"Cannot evaluate RID system: injected test flights ended at {t_end}, which is before now ({t_now})"
            )

        t_next = arrow.utcnow()
        dt = self._evaluation_configuration.configuration.min_polling_interval.timedelta
        while arrow.utcnow() < t_end:
            # Evaluate the system at an instant in time
            rect = self._virtual_observer.get_query_rect()
            evaluator.evaluate_system_instantaneously(self._observers.observers, rect)

            # Wait until minimum polling interval elapses
            while t_next < arrow.utcnow():
                t_next += dt
            if t_next > t_end:
                break
            delay = t_next - arrow.utcnow()
            if delay.total_seconds() > 0:
                print(
                    f"Waiting {delay.total_seconds()} seconds before polling RID observers again..."
                )
                time.sleep(delay.total_seconds())

        return True

    def cleanup(self):
        self.begin_cleanup()
        while self._injected_flights:
            injected_flight = self._injected_flights.pop()
            matching_sps = [
                sp
                for sp in self._service_providers.service_providers
                if sp.participant_id == injected_flight.uss_participant_id
            ]
            if len(matching_sps) != 1:
                matching_ids = ", ".join(sp.participant_id for sp in matching_sps)
                raise RuntimeError(
                    f"Found {len(matching_sps)} service providers with participant ID {injected_flight.uss_participant_id} ({matching_ids}) when exactly 1 was expected"
                )
            sp = matching_sps[0]
            check = self.check("Successful test deletion", [sp.participant_id])
            try:
                query = sp.delete_test(injected_flight.test_id)
                self.record_query(query)
                if query.status_code != 200:
                    raise ValueError(
                        f"Received status code {query.status_code} after attempting to delete test {injected_flight.test_id} from service provider {sp.participant_id}"
                    )
                check.record_passed()
            except (RequestException, ValueError) as e:
                stacktrace = "".join(
                    traceback.format_exception(
                        etype=type(e), value=e, tb=e.__traceback__
                    )
                )
                check.record_failed(
                    summary="Error while trying to delete test flight",
                    severity=Severity.Medium,
                    details=f"While trying to delete a test flight from {sp.participant_id}, encountered error:\n{stacktrace}",
                )
        self.end_cleanup()
