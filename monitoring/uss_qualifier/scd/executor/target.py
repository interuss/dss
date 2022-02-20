
from typing import Dict, Tuple

from monitoring.monitorlib import infrastructure, auth, fetch
from monitoring.monitorlib.clients.scd_automated_testing import create_flight, delete_flight, QueryError
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightResult, \
    DeleteFlightResult, InjectFlightResponse, DeleteFlightResponse
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import FlightInjectionAttempt

class TestTarget():
    """A class managing the state and the interactions with a target"""

    def __init__(self, name: str, config: InjectionTargetConfiguration, auth_spec: str):
        self.name = name
        self.config = config
        self.client = infrastructure.DSSTestSession(
            self.config.injection_base_url,
            auth.make_auth_adapter(auth_spec))

        # Flights injected by this target.
        # Key: flight name
        # Value: flight id
        self.created_flight_ids: Dict[str, str] = {}


    def __repr__(self):
        return "TestTarget({}, {})".format(self.name, self.config.injection_base_url)


    def inject_flight(self, flight_request: FlightInjectionAttempt) -> Tuple[InjectFlightResponse, fetch.Query]:
        flight_id, resp, query = create_flight(self.client, self.config.injection_base_url, flight_request.test_injection)

        if resp.result == InjectFlightResult.Planned:
            self.created_flight_ids[flight_request.name] = flight_id

        return resp, query


    def delete_flight(self, flight_name: str) -> Tuple[DeleteFlightResponse, fetch.Query]:
        flight_id = self.created_flight_ids[flight_name]
        resp, query = delete_flight(self.client, self.config.injection_base_url, flight_id)

        if resp.result == DeleteFlightResult.Closed:
            del self.created_flight_ids[flight_name]
        elif resp.result == DeleteFlightResult.Failed:
            raise QueryError("Unable to delete flight {}. Result: {} Notes: {}".format(flight_name, resp.result, resp.get("notes", None)), query)
        else:
            raise NotImplementedError("Unsupported DeleteFlightResult {}".format(resp.get("result", None)))

        return resp, query


    def managed_flights(self):
        """Get flight names managed by this test target"""
        return list(self.created_flight_ids.keys())


    def is_managing_flight(self, flight_name: str) -> bool:
        return flight_name in self.created_flight_ids.keys()
