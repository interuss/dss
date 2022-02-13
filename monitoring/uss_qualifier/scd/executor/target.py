import itertools
import json
import os
import typing
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List

from monitoring.monitorlib import infrastructure, auth
from monitoring.monitorlib.clients.scd import OperationError
from monitoring.monitorlib.clients.scd_automated_testing import create_flight, delete_flight
from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResult, \
    DeleteFlightResult, InjectFlightResponse
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest, TestStep, FlightInjectionAttempt
from monitoring.uss_qualifier.utils import is_url

FlightName = str

class TestTarget():
    """A class managing the state and the interactions with a target"""

    def __init__(self, name: str, config: InjectionTargetConfiguration, auth_spec: str):
        self.name = name
        self.config = config
        self.client = infrastructure.DSSTestSession(
            self.config.injection_base_url,
            auth.make_auth_adapter(auth_spec))

        # Flights injected by this target
        self.created_flight_ids: Dict[str, str] = {}

    def __repr__(self):
        return "TestTarget({}, {})".format(self.name, self.config.injection_base_url)

    def inject_flight(self, flight_request: FlightInjectionAttempt, dry=False):
        flight_id, resp = create_flight(self.client, self.config.injection_base_url, flight_request.test_injection, dry=dry)
        if resp.result in [InjectFlightResult.Planned, InjectFlightResult.DryRun]:
            self.created_flight_ids[flight_request.name] = flight_id
        return resp
        # TODO: Handle errors

    def delete_flight(self, flight_name: str, dry=False):
        flight_id = self.created_flight_ids[flight_name]
        resp = delete_flight(self.client, self.config.injection_base_url, flight_id, dry=dry)
        if resp.result in [DeleteFlightResult.Closed, InjectFlightResult.DryRun]:
            del self.created_flight_ids[flight_name]
        # TODO: Handle errors

    def delete_all_flights(self, dry=False) -> int:
        flights_count = len(self.created_flight_ids.keys())
        print("[SCD]    - Deleting {} flights for target {}.".format(flights_count, self.name))
        for flight_name, flight_id in list(self.created_flight_ids.items()):
            self.delete_flight(flight_name, dry=dry)
        return flights_count

    def is_managing_flight(self, flight_name: str):
        return flight_name in self.created_flight_ids.keys()
