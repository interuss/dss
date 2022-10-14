import json
from typing import List, Optional
import uuid

import arrow
from implicitdict import ImplicitDict, StringBasedDateTime, StringBasedTimeDelta
import requests

from monitoring.monitorlib.rid import RIDAircraftState, RIDFlightDetails
from monitoring.monitorlib.rid_automated_testing.injection_api import (
    TestFlightDetails,
    TestFlight,
)
from monitoring.uss_qualifier.resources import Resource


class FullFlightRecord(ImplicitDict):
    reference_time: StringBasedDateTime
    """The reference time of this flight (usually the time of first telemetry)"""

    states: List[RIDAircraftState]
    """All telemetry that will be/was received for this flight"""

    flight_details: RIDFlightDetails
    """Details of this flight, as would be reported at the ASTM /details endpoint"""

    aircraft_type: str
    """Type of aircraft, as per RIDFlight.aircraft_type"""


class FlightRecordCollection(ImplicitDict):
    flights: List[FullFlightRecord]


class FlightDataJSONFileConfiguration(ImplicitDict):
    path: str
    """Path to a file containing a JSON representation of an instance of FlightRecordCollection.  This may be a local file or a web URL."""


class FlightDataSpecification(ImplicitDict):
    flight_start_delay: StringBasedTimeDelta = StringBasedTimeDelta("15s")
    """Amount of time between starting the test and commencement of flights"""

    json_file_source: Optional[FlightDataJSONFileConfiguration] = None
    """When this field is populated, flight data will be loaded from a JSON file"""


class FlightDataResource(Resource[FlightDataSpecification]):
    _flight_collection: FlightRecordCollection

    def __init__(self, specification: FlightDataSpecification):
        if specification.json_file_source is not None:
            if specification.json_file_source.path.startswith("http"):
                resp = requests.get(specification.json_file_source.path)
                resp.raise_for_status()
                self._flight_collection = ImplicitDict.parse(
                    json.loads(resp.content.decode("utf-8")), FlightRecordCollection
                )
            else:
                with open(specification.json_file_source.path, "r") as f:
                    self._flight_collection = ImplicitDict.parse(
                        json.load(f), FlightRecordCollection
                    )
            self._flight_start_delay = specification.flight_start_delay
        else:
            raise ValueError(
                "A source of flight data was not identified in the specification for a NetRIDFlightDataResource"
            )

    def get_test_flights(self) -> List[TestFlight]:
        t0 = arrow.utcnow() + self._flight_start_delay.timedelta

        test_flights: List[TestFlight] = []

        for flight in self._flight_collection.flights:
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
