from datetime import timedelta, datetime
from typing import Optional

import arrow
from s2sphere import LatLngRect

from monitoring.uss_qualifier.scenarios.astm.netrid.injected_flight_collection import (
    InjectedFlightCollection,
)


class VirtualObserver(object):
    """Defines the behavior of a virtual human-like observer.

    The observer wants to look at the specified collection of flights, and this
    class indicates their behavior by computing the query rectangle at each
    polling instance.
    """

    _injected_flights: InjectedFlightCollection
    """Set of flights this virtual observer is going to observe"""

    _repeat_query_rect_period: int
    """If set to a value above zero, reuse the most recent query rectangle/view every this many queries."""

    _min_query_diagonal_m: float
    """Do not make queries with diagonals smaller than this many meters."""

    _relevant_past_data_period: timedelta
    """Length of time prior to a query that may contain relevant observable data"""

    _repeat_query_counter: int = 0
    """Number of repeated queries to the same rectangle; related to _repeat_query_rect_period"""

    _last_rect: Optional[LatLngRect] = None
    """The most recent query rectangle"""

    def __init__(
        self,
        injected_flights: InjectedFlightCollection,
        repeat_query_rect_period: int,
        min_query_diagonal_m: float,
        relevant_past_data_period: timedelta,
    ):
        self._injected_flights = injected_flights
        self._repeat_query_rect_period = repeat_query_rect_period
        self._min_query_diagonal_m = min_query_diagonal_m
        self._relevant_past_data_period = relevant_past_data_period

    def get_query_rect(self) -> LatLngRect:
        t_now = arrow.utcnow().datetime
        if (
            self._last_rect
            and self._repeat_query_rect_period > 0
            and self._repeat_query_counter >= self._repeat_query_rect_period == 0
        ):
            rect = self._last_rect
        else:
            t_min = t_now - self._relevant_past_data_period
            rect = self._injected_flights.get_query_rect(
                t_min, t_now, self._min_query_diagonal_m
            )
            self._last_rect = rect
        return rect

    def get_last_time_of_interest(self) -> datetime:
        """Return the time after which there will be no data of interest to this observer."""
        return (
            self._injected_flights.get_end_of_injected_data()
            + self._relevant_past_data_period
        )
