import datetime
from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib.infrastructure import UTMClientSession
import json, os
import uuid
from pathlib import Path
from monitoring.monitorlib import fetch
from monitoring.uss_qualifier.rid.utils import FullFlightRecord
from monitoring.uss_qualifier.rid import reports
from monitoring.monitorlib.rid_automated_testing.injection_api import (
    TestFlightDetails,
    TestFlight,
    CreateTestParameters,
    SCOPE_RID_QUALIFIER_INJECT,
    ChangeTestResponse,
)
from monitoring.monitorlib.typing import ImplicitDict
import arrow

from typing import List
from monitoring.uss_qualifier.rid.utils import RIDQualifierTestConfiguration


def get_full_flight_records(aircraft_states_directory: Path) -> List[FullFlightRecord]:
    """Gets full flight records from the specified directory if they exist"""

    if not os.path.exists(aircraft_states_directory):
        raise ValueError(
            "The aircraft states directory does not exist: {}".format(
                aircraft_states_directory
            )
        )

    all_files = os.listdir(aircraft_states_directory)
    files = [
        os.path.join(aircraft_states_directory, f)
        for f in all_files
        if os.path.isfile(os.path.join(aircraft_states_directory, f))
    ]

    if not files:
        raise ValueError(
            "There are no states in the states directory, create states first using the simulator/flight_state module."
        )

    flight_records: List[FullFlightRecord] = []
    for file in files:
        with open(file, "r") as f:
            flight_records.append(ImplicitDict.parse(json.load(f), FullFlightRecord))

    return flight_records


class TestBuilder:
    """A class to setup the test data and create the objects ready to be submitted to the test harness"""

    def __init__(
        self,
        test_configuration: RIDQualifierTestConfiguration,
        flight_records: List[FullFlightRecord],
    ) -> None:
        self.test_configuration = test_configuration
        usses = self.test_configuration.injection_targets
        if len(usses) > len(flight_records):
            raise ValueError(
                "There are not enough flight records ({}) to test the specified USSes ({})".format(
                    len(usses), len(flight_records)
                )
            )
        self.disk_flight_records: List[FullFlightRecord] = flight_records

    def build_test_payloads(self) -> List[CreateTestParameters]:
        """This is the main method to process the test configuration and build RID payload object, maxium of one flight is allocated to each USS."""

        all_test_payloads = []  # This holds the data that will be deilver

        test_reference_time = arrow.now()
        test_start_time = (
            test_reference_time + self.test_configuration.flight_start_delay.timedelta
        )
        test_start_isoformat = test_start_time.isoformat()

        for state_data_index, flight_record in enumerate(self.disk_flight_records):
            disk_reference_time_raw = flight_record.reference_time
            disk_reference_time = arrow.get(disk_reference_time_raw)

            flight_record.reference_time = test_reference_time.isoformat()

            timestamp_offset = test_start_time - disk_reference_time

            for telemetry_id, flight_telemetry in enumerate(flight_record.states):
                timestamp = arrow.get(flight_telemetry.timestamp) + timestamp_offset
                flight_telemetry.timestamp = timestamp.isoformat()

            test_flight_details = TestFlightDetails(
                effective_after=test_start_isoformat,
                details=flight_record.flight_details.rid_details,
            )

            test_flight = TestFlight(
                injection_id=str(uuid.uuid4()),
                telemetry=flight_record.states,
                details_responses=[test_flight_details],
            )

            test_payload = CreateTestParameters(requested_flights=[test_flight])

            all_test_payloads.append(test_payload)

        return all_test_payloads


class TestHarness:
    """A class to submit Aircraft RID State to the USS test endpoint"""

    def __init__(self, auth_spec: str, injection_base_url: str):

        auth_adapter = make_auth_adapter(auth_spec)
        self._base_url = injection_base_url
        self.uss_session = UTMClientSession(injection_base_url, auth_adapter)

    def submit_test(
        self, payload: CreateTestParameters, test_id: str, setup: reports.Setup
    ) -> List[TestFlight]:
        injection_path = "/tests/{}".format(test_id)

        initiated_at = datetime.datetime.utcnow()
        response = self.uss_session.put(
            url=injection_path, json=payload, scope=SCOPE_RID_QUALIFIER_INJECT
        )
        setup.injections.append(fetch.describe_query(response, initiated_at))

        if response.status_code == 200:
            changed_test: ChangeTestResponse = ImplicitDict.parse(response.json(), ChangeTestResponse)
            print("New test with ID %s created" % test_id)
            return changed_test.injected_flights
        else:
            raise RuntimeError(
                "Error {} submitting test ID {} to {}: {}".format(
                    response.status_code,
                    test_id,
                    self._base_url,
                    response.content.decode("utf-8"),
                )
            )
