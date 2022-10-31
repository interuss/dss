from datetime import timedelta
from typing import List, Optional
import uuid

import arrow
from implicitdict import ImplicitDict, StringBasedDateTime
from uas_standards.interuss.automated_testing.rid.v1.injection import TestFlightDetails

from monitoring.monitorlib.rid import RIDAircraftState
from monitoring.monitorlib.rid_automated_testing.injection_api import (
    TestFlight,
)
from monitoring.uss_qualifier.fileio import load_dict_with_references, load_content
from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.netrid.flight_data import (
    FlightDataSpecification,
    FlightRecordCollection,
)
from monitoring.uss_qualifier.resources.netrid.simulation.adjacent_circular_flights_simulator import (
    generate_aircraft_states,
)
from monitoring.uss_qualifier.resources.netrid.simulation.kml_flights import (
    get_flight_records,
)


class FlightDataResource(Resource[FlightDataSpecification]):
    _flight_start_delay: timedelta
    flight_collection: FlightRecordCollection

    def __init__(self, specification: FlightDataSpecification):
        if "record_source" in specification:
            self.flight_collection = ImplicitDict.parse(
                load_dict_with_references(specification.record_source),
                FlightRecordCollection,
            )
        elif "adjacent_circular_flights_simulation_source" in specification:
            self.flight_collection = generate_aircraft_states(
                specification.adjacent_circular_flights_simulation_source
            )
        elif "kml_source" in specification:
            kml_content = load_content(specification.kml_source.kml_location)
            self.flight_collection = get_flight_records(
                kml_content,
                specification.kml_source.reference_time.datetime,
                specification.kml_source.random_seed,
            )
        else:
            raise ValueError(
                "A source of flight data was not identified in the specification for a FlightDataSpecification"
            )
        self._flight_start_delay = specification.flight_start_delay.timedelta

    def get_test_flights(self) -> List[TestFlight]:
        t0 = arrow.utcnow() + self._flight_start_delay

        test_flights: List[TestFlight] = []

        for flight in self.flight_collection.flights:
            dt = t0 - flight.reference_time.datetime

            telemetry: List[RIDAircraftState] = []
            for state in flight.states:
                shifted_state = RIDAircraftState(state)
                shifted_state.timestamp = StringBasedDateTime(
                    state.timestamp.datetime + dt
                )
                telemetry.append(shifted_state)

            details = TestFlightDetails(
                effective_after=StringBasedDateTime(t0),
                details=flight.flight_details,
                aircraft_type=flight.aircraft_type,
            )

            test_flights.append(
                TestFlight(
                    injection_id=str(uuid.uuid4()),
                    telemetry=telemetry,
                    details_responses=[details],
                )
            )

        return test_flights


class FlightDataStorageSpecification(ImplicitDict):
    flight_record_collection_path: Optional[str]
    """Path, usually ending with .json, at which to store the FlightRecordCollection"""

    geojson_tracks_path: Optional[str]
    """Path (folder) in which to store track_XX.geojson files, 1 for each flight"""


class FlightDataStorageResource(Resource[FlightDataStorageSpecification]):
    storage_configuration: FlightDataStorageSpecification

    def __init__(self, specification: FlightDataStorageSpecification):
        self.storage_configuration = specification
