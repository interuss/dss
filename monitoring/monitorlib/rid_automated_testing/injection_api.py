import datetime
from typing import List, Optional, Tuple

import arrow
import s2sphere

from monitoring.monitorlib import geo, rid

from uas_standards.interuss.automated_testing.rid.v1 import injection


SCOPE_RID_QUALIFIER_INJECT = "rid.inject_test_data"


class TestFlight(injection.TestFlight):
    def get_span(
        self,
    ) -> Tuple[Optional[datetime.datetime], Optional[datetime.datetime]]:
        earliest = None
        latest = None
        times = [
            arrow.get(aircraft_state.timestamp).datetime
            for aircraft_state in self.telemetry
        ]
        times.extend(
            arrow.get(details.effective_after).datetime
            for details in self.details_responses
        )
        for t in times:
            if earliest is None or t < earliest:
                earliest = t
            if latest is None or t > latest:
                latest = t
        return (earliest, latest)

    def get_details(self, t_now: datetime.datetime) -> Optional[rid.RIDFlightDetails]:
        latest_after: Optional[datetime.datetime] = None
        tf_details = None
        for response in self.details_responses:
            t_response = arrow.get(response.effective_after).datetime
            if t_now >= t_response:
                if latest_after is None or t_response > latest_after:
                    latest_after = t_response
                    tf_details = response.details
        return tf_details

    def get_id(self, t_now: datetime.datetime) -> Optional[str]:
        details = self.get_details(t_now)
        return details.id if details else None

    def order_telemetry(self):
        self.telemetry = sorted(
            self.telemetry, key=lambda telemetry: telemetry.timestamp.datetime
        )

    def select_relevant_states(
        self, view: s2sphere.LatLngRect, t0: datetime.datetime, t1: datetime.datetime
    ) -> List[rid.RIDAircraftState]:
        recent_states: List[rid.RIDAircraftState] = []
        previously_outside = False
        previously_inside = False
        previous_telemetry = None
        for telemetry in self.telemetry:
            if telemetry.timestamp.datetime < t0 or telemetry.timestamp.datetime > t1:
                # Telemetry not relevant based on time
                continue
            pt = s2sphere.LatLng.from_degrees(
                telemetry.position.lat, telemetry.position.lng
            )
            inside_now = view.contains(pt)
            if inside_now:
                if previously_outside:
                    recent_states.append(previous_telemetry)
                recent_states.append(telemetry)
                previously_inside = True
                previously_outside = False
            else:
                if previously_inside:
                    recent_states.append(telemetry)
                previously_outside = True
                previously_inside = False
            previous_telemetry = telemetry
        return recent_states

    def get_rect(self) -> Optional[s2sphere.LatLngRect]:
        return geo.bounding_rect(
            [(t.position.lat, t.position.lng) for t in self.telemetry]
        )


class CreateTestParameters(injection.CreateTestParameters):
    def get_span(
        self,
    ) -> Tuple[Optional[datetime.datetime], Optional[datetime.datetime]]:
        if not self.requested_flights:
            return (None, None)
        (earliest, latest) = (None, None)
        for flight in self.requested_flights:
            flight = TestFlight(flight)
            (t0, t1) = flight.get_span()
            if earliest is None or t0 < earliest:
                earliest = t0
            if latest is None or t1 > latest:
                latest = t1
        return (earliest, latest)

    def get_rect(self) -> Optional[s2sphere.LatLngRect]:
        result = None
        for flight in self.requested_flights:
            flight = TestFlight(flight)
            if result is None:
                result = flight.get_rect()
            else:
                result = result.union(flight.get_rect())
        return result
