from datetime import datetime
from typing import List

import arrow
from s2sphere import LatLng, LatLngRect

from monitoring.monitorlib import geo
from monitoring.uss_qualifier.scenarios.astm.netrid.injection import InjectedFlight


class InjectedFlightCollection(object):
    _injected_flights: List[InjectedFlight]

    def __init__(self, injected_flights: List[InjectedFlight]):
        self._injected_flights = injected_flights

    def get_query_rect(
        self, t_min: datetime, t_max: datetime, min_query_diagonal_m: float
    ) -> LatLngRect:
        # Find the bounds of all relevant points
        data_exists = False
        lat_min = 90
        lng_min = 360
        lat_max = -90
        lng_max = -360
        for injected_flight in self._injected_flights:
            for telemetry in injected_flight.flight.telemetry:
                t = arrow.get(telemetry.timestamp).datetime
                if t_min <= t <= t_max:
                    data_exists = True
                    lat_min = min(lat_min, telemetry.position.lat)
                    lat_max = max(lat_max, telemetry.position.lat)
                    lng_min = min(lng_min, telemetry.position.lng)
                    lng_max = max(lng_max, telemetry.position.lng)

        # If there is no flight data yet, look at the center of where the data will be
        if not data_exists:
            lat = 0
            lng = 0
            n = 0
            for injected_flight in self._injected_flights:
                for telemetry in injected_flight.flight.telemetry:
                    lat += telemetry.position.lat
                    lng += telemetry.position.lng
                    n += 1
            lat_min = lat_max = lat / n
            lng_min = lng_max = lng / n

        # Expand view size to meet minimum, if necessary
        OVERSHOOT = 1.01
        while True:
            c1 = LatLng.from_degrees(lat_min, lng_min)
            c2 = LatLng.from_degrees(lat_max, lng_max)
            diagonal_m = (
                c1.get_distance(c2).degrees * geo.EARTH_CIRCUMFERENCE_KM * 1000 / 360
            )
            if diagonal_m >= min_query_diagonal_m:
                break
            if lat_min == lat_max and lng_min == lng_max:
                lat_min -= 1e-5
                lat_max += 1e-5
                lng_min -= 1e-5
                lng_max += 1e-5
                continue
            lat_center = 0.5 * (lat_min + lat_max)
            lat_span = (
                (lat_max - lat_min) * min_query_diagonal_m / diagonal_m * OVERSHOOT
            )
            lat_min = lat_center - 0.5 * lat_span
            lat_max = lat_center + 0.5 * lat_span
            lng_center = 0.5 * (lng_min + lng_max)
            lng_span = (
                (lng_max - lng_min) * min_query_diagonal_m / diagonal_m * OVERSHOOT
            )
            lng_min = lng_center - 0.5 * lng_span
            lng_max = lng_center + 0.5 * lng_span

        p1 = LatLng.from_degrees(lat_min, lng_min)
        p2 = LatLng.from_degrees(lat_max, lng_max)
        return LatLngRect.from_point_pair(p1, p2)

    def get_end_of_injected_data(self) -> datetime:
        t_end = arrow.utcnow()
        for injected_flight in self._injected_flights:
            for telemetry in injected_flight.flight.telemetry:
                t = arrow.get(telemetry.timestamp)
                t_end = max(t_end, t)
        return t_end
