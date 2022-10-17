from typing import Dict, Tuple
from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.monitorlib import infrastructure, fetch
from monitoring.monitorlib.clients.scd_automated_testing import (
    create_flight,
    delete_flight,
    QueryError,
    get_version,
    get_capabilities,
)
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightResult,
    DeleteFlightResult,
    InjectFlightResponse,
    DeleteFlightResponse,
)
from monitoring.uss_qualifier.resources.flight_planning.automated_test import (
    FlightInjectionAttempt,
)


class FlightPlannerConfiguration(ImplicitDict):
    participant_id: str
    """ID of the flight planner into which test data can be injected"""

    injection_base_url: str
    """Base URL for the flight planner's implementation of the interfaces/automated-testing/scd/scd.yaml API"""

    def __init__(self, *args, **kwargs):
        super().__init__(**kwargs)
        try:
            urlparse(self.injection_base_url)
        except ValueError:
            raise ValueError(
                "FlightPlannerConfiguration.injection_base_url must be a URL"
            )


class TestTarget:
    """A class managing the state and the interactions with a target"""

    def __init__(
        self,
        config: FlightPlannerConfiguration,
        auth_adapter: infrastructure.AuthAdapter,
    ):
        self.config = config
        self.client = infrastructure.UTMClientSession(
            self.config.injection_base_url, auth_adapter
        )

        # Flights injected by this target.
        # Key: flight name
        # Value: flight id
        self.created_flight_ids: Dict[str, str] = {}

    def __repr__(self):
        return "TestTarget({}, {})".format(
            self.config.participant_id, self.config.injection_base_url
        )

    @property
    def name(self) -> str:
        return self.config.participant_id

    def inject_flight(
        self, flight_request: FlightInjectionAttempt
    ) -> Tuple[InjectFlightResponse, fetch.Query, str]:
        flight_id, resp, query = create_flight(
            self.client, self.config.injection_base_url, flight_request.test_injection
        )

        if resp.result == InjectFlightResult.Planned:
            self.created_flight_ids[flight_request.name] = flight_id

        return resp, query, flight_id

    def delete_flight(
        self, flight_name: str
    ) -> Tuple[DeleteFlightResponse, fetch.Query]:
        flight_id = self.created_flight_ids[flight_name]
        resp, query = delete_flight(
            self.client, self.config.injection_base_url, flight_id
        )

        if resp.result == DeleteFlightResult.Closed:
            del self.created_flight_ids[flight_name]
        elif resp.result == DeleteFlightResult.Failed:
            raise QueryError(
                "Unable to delete flight {}. Result: {} Notes: {}".format(
                    flight_name, resp.result, resp.get("notes", None)
                ),
                query,
            )
        else:
            raise NotImplementedError(
                "Unsupported DeleteFlightResult {}".format(resp.get("result", None))
            )

        return resp, query

    def managed_flights(self):
        """Get flight names managed by this test target"""
        return list(self.created_flight_ids.keys())

    def is_managing_flight(self, flight_name: str) -> bool:
        return flight_name in self.created_flight_ids.keys()

    def get_target_information(self):
        resp, _ = get_version(self.client, self.config.injection_base_url)
        version = resp.version if resp.version is not None else "Unknown"
        resp, _ = get_capabilities(self.client, self.config.injection_base_url)

        return {"version": version, "capabilities": resp.capabilities}
