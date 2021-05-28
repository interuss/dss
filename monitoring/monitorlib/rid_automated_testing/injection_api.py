import datetime
from typing import List, Optional, Tuple

import arrow

from monitoring.monitorlib import rid
from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime


SCOPE_RID_QUALIFIER_INJECT = 'rid.inject_test_data'

# Mirrors of types defined in remote ID automated testing injection API

class OperatorLocation(ImplicitDict):
    ''' A object to hold location of the operator when submitting flight data to USS '''
    lat: float
    lng: float


class TestFlightDetails(ImplicitDict):
    ''' A object to hold the remote ID Details,  and a date time after which the USS should submit the flight details, it matches the TestFlightDetails in the injection interface, for more details see: https://github.com/interuss/dss/blob/master/interfaces/automated-testing/rid/injection.yaml#L158 '''
    effective_after: StringBasedDateTime # ISO 8601 datetime string
    details: rid.RIDFlightDetails


class TestFlight(ImplicitDict):
    ''' Represents the data necessary to inject a single, complete test flight into a Remote ID Service Provider under test; matches TestFlight in injection interface '''

    injection_id: str
    telemetry: List[rid.RIDAircraftState]
    details_responses : List[TestFlightDetails]

    def get_span(self) -> Tuple[Optional[datetime.datetime], Optional[datetime.datetime]]:
        earliest = None
        latest = None
        times = [arrow.get(aircraft_state.timestamp).datetime
                 for aircraft_state in self.telemetry]
        times.extend(arrow.get(details.effective_after).datetime
                     for details in self.details_responses)
        for t in times:
            if earliest is None or t < earliest:
                earliest = t
            if latest is None or t > latest:
                latest = t
        return (earliest, latest)

    def get_details(self, t_now: datetime.datetime) -> Optional[TestFlightDetails]:
        latest_after: Optional[datetime.datetime] = None
        tf_details = None
        for response in self.details_responses:
            t_response = response.effective_after.datetime
            if t_now >= t_response:
                if latest_after is None or t_response > latest_after:
                    latest_after = t_response
                    tf_details = response.details
        return tf_details

    def get_id(self, t_now: datetime.datetime) -> Optional[str]:
        details = self.get_details(t_now)
        return details.id if details else None

    def order_telemetry(self):
        self.telemetry = sorted(self.telemetry,
                                key=lambda telemetry: telemetry.timestamp.datetime)


class CreateTestParameters(ImplicitDict):
    requested_flights: List[TestFlight]


class ChangeTestResponse(ImplicitDict):
  injected_flights: List[TestFlight]
  version: str
