import datetime
import time
from typing import List, Optional

import arrow
import s2sphere

from monitoring.monitorlib import fetch, geo
from monitoring.monitorlib.rid_common import RIDVersion
from monitoring.uss_qualifier.common_data_definitions import Severity
from monitoring.uss_qualifier.resources.netrid.evaluation import EvaluationConfiguration
from monitoring.uss_qualifier.resources.netrid.observers import RIDSystemObserver
from monitoring.uss_qualifier.scenarios import TestScenarioType
from monitoring.uss_qualifier.scenarios.astm.netrid.injection import InjectedFlight
from monitoring.monitorlib.rid_automated_testing import observation_api


def _rect_str(rect) -> str:
    return "({}, {})-({}, {})".format(
        rect.lo().lat().degrees,
        rect.lo().lng().degrees,
        rect.hi().lat().degrees,
        rect.hi().lng().degrees,
    )


class RIDObservationEvaluator(object):
    """Evaluates observations of an RID system over time.

    This evaluator observes a set of provided RIDSystemObservers in
    evaluate_system by repeatedly polling them according to the expected data
    provided to RIDObservationEvaluator upon construction.  During these
    evaluations, RIDObservationEvaluator mutates provided findings object to add
    additional findings.
    """

    def __init__(
        self,
        test_scenario: TestScenarioType,
        injected_flights: List[InjectedFlight],
        config: EvaluationConfiguration,
        rid_version: RIDVersion,
    ):
        self._test_scenario = test_scenario
        self._injected_flights = injected_flights
        self._config = config
        self._rid_version = rid_version

    def _get_query_rect(
        self,
        t: datetime.datetime,
    ) -> s2sphere.LatLngRect:
        data_exists = False
        lat_min = 90
        lng_min = 360
        lat_max = -90
        lng_max = -360

        # Find the bounds of all relevant points
        t_min = (
            t
            - self._rid_version.realtime_period
            - self._config.max_propagation_latency.timedelta
        )
        t_max = t
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
            c1 = s2sphere.LatLng.from_degrees(lat_min, lng_min)
            c2 = s2sphere.LatLng.from_degrees(lat_max, lng_max)
            diagonal = (
                c1.get_distance(c2).degrees * geo.EARTH_CIRCUMFERENCE_KM * 1000 / 360
            )
            if diagonal >= self._config.min_query_diagonal:
                break
            if lat_min == lat_max and lng_min == lng_max:
                lat_min -= 1e-5
                lat_max += 1e-5
                lng_min -= 1e-5
                lng_max += 1e-5
                continue
            lat_center = 0.5 * (lat_min + lat_max)
            lat_span = (
                (lat_max - lat_min)
                * self._config.min_query_diagonal
                / diagonal
                * OVERSHOOT
            )
            lat_min = lat_center - 0.5 * lat_span
            lat_max = lat_center + 0.5 * lat_span
            lng_center = 0.5 * (lng_min + lng_max)
            lng_span = (
                (lng_max - lng_min)
                * self._config.min_query_diagonal
                / diagonal
                * OVERSHOOT
            )
            lng_min = lng_center - 0.5 * lng_span
            lng_max = lng_center + 0.5 * lng_span

        p1 = s2sphere.LatLng.from_degrees(lat_min, lng_min)
        p2 = s2sphere.LatLng.from_degrees(lat_max, lng_max)
        return s2sphere.LatLngRect.from_point_pair(p1, p2)

    def evaluate_system(self, observers: List[RIDSystemObserver]) -> None:
        """Evaluate a system by polling system state and comparing to expectations.

        This routine periodically polls each of the specified observers for the system
        state and checks that each system state matches expectations based on the
        provided injected flights, updating the provided report findings.
        """

        # Compute the end of all injected data
        t_end = arrow.utcnow()
        for injected_flight in self._injected_flights:
            for telemetry in injected_flight.flight.telemetry:
                t = arrow.get(telemetry.timestamp)
                t_end = max(t_end, t)
        t_end += (
            self._rid_version.realtime_period
            + self._config.max_propagation_latency.timedelta
        )

        if arrow.utcnow() > t_end:
            raise RuntimeError(
                "Cannot evaluate system: injected test flights ended at {}, which is before now ({})".format(
                    t_end, datetime.datetime.utcnow()
                )
            )

        query_counter = 0
        last_rect = None

        t_next = arrow.utcnow()

        while arrow.utcnow() < t_end:
            # Evaluate the system at an instant in time

            t_now = arrow.utcnow().datetime
            if (
                last_rect
                and self._config.repeat_query_rect_period > 0
                and query_counter % self._config.repeat_query_rect_period == 0
            ):
                rect = last_rect
            else:
                rect = self._get_query_rect(
                    t_now,
                )
                last_rect = rect
            self._evaluate_system_instantaneously(observers, rect)
            print("Completed observation at {}".format(arrow.utcnow()))

            # Wait until minimum polling interval elapses
            while t_next < arrow.utcnow():
                t_next += self._config.min_polling_interval.timedelta
            if t_next > t_end:
                break
            delay = t_next - arrow.utcnow()
            if delay.total_seconds() > 0:
                time.sleep(delay.total_seconds())
            query_counter += 1

    def _evaluate_system_instantaneously(
        self,
        observers: List[RIDSystemObserver],
        rect: s2sphere.LatLngRect,
    ) -> None:
        for observer in observers:
            # Conduct an observation, then log and evaluate it
            (observation, query) = observer.observe_system(rect)
            self._test_scenario.record_query(query)
            self._evaluate_observation(
                observer,
                rect,
                observation,
                query,
            )

            # TODO: If bounding rect is smaller than cluster threshold, expand slightly above cluster threshold and re-observe
            # TODO: If bounding rect is smaller than area-too-large threshold, expand slightly above area-too-large threshold and re-observe

    def _evaluate_observation(
        self,
        observer: RIDSystemObserver,
        rect: s2sphere.LatLngRect,
        observation: Optional[observation_api.GetDisplayDataResponse],
        query: fetch.Query,
    ) -> None:
        diagonal_km = (
            rect.lo().get_distance(rect.hi()).degrees * geo.EARTH_CIRCUMFERENCE_KM / 360
        )
        if diagonal_km > self._rid_version.max_diagonal_km:
            self._evaluate_area_too_large_observation(observer, rect, diagonal_km, query)
        elif diagonal_km > self._rid_version.max_details_diagonal_km:
            self._evaluate_clusters_observation()
        else:
            self._evaluate_normal_observation(
                observer,
                rect,
                observation,
                query,
            )

    def _evaluate_normal_observation(
        self,
        observer: RIDSystemObserver,
        rect: s2sphere.LatLngRect,
        observation: Optional[observation_api.GetDisplayDataResponse],
        query: fetch.Query,
    ) -> None:
        if observation is None:
            self._test_scenario.record_failed_check(
                name="Successful observation",
                summary="Observation failed",
                details=f"When queried for an observation in {_rect_str(rect)}, {observer.participant_id} returned code {query.status_code}",
                severity=Severity.Critical,
                relevant_participants=[observer.participant_id],
                query_timestamps=[query.request.timestamp],
            )
            return

        for expected_flight in self._injected_flights:
            t_initiated = query.request.timestamp
            t_response = query.response.reported
            timestamps = [
                arrow.get(t.timestamp) for t in expected_flight.flight.telemetry
            ]
            t_min = min(timestamps).datetime
            t_max = max(timestamps).datetime

            flight_id = expected_flight.flight.details_responses[
                0
            ].details.id  # TODO: Choose appropriate details rather than first
            matching_flights = [
                observed_flight
                for observed_flight in observation.flights
                if observed_flight.id == flight_id
            ]
            if len(matching_flights) > 1:
                self._test_scenario.record_failed_check(
                    name="Duplicate flights",
                    summary="Duplicate flights observed",
                    details=f'When queried for an observation in {_rect_str(rect)}, {observer.participant_id} found {len(matching_flights)} flights with flight ID "{flight_id}" that was injected into {expected_flight.uss_participant_id}',
                    severity=Severity.Critical,
                    relevant_participants=[
                        expected_flight.uss_participant_id,
                        observer.participant_id,
                    ],
                    query_timestamps=[
                        query.request.timestamp,
                        expected_flight.query_timestamp,
                    ],
                )

            if t_response < t_min:
                # This flight should definitely not have been observed (it starts in the future)
                if matching_flights:
                    self._test_scenario.record_failed_check(
                        name="Premature flight",
                        summary="Flight observed before it started",
                        details=f"Flight {flight_id} injected into {expected_flight.uss_participant_id} was observed by {observer.participant_id} at {t_response.isoformat()} before that flight should have started at {t_min.isoformat()}",
                        severity=Severity.High,
                        relevant_participants=[expected_flight.uss_participant_id],
                        query_timestamps=[
                            query.request.timestamp,
                            expected_flight.query_timestamp,
                        ],
                    )
                # TODO: attempt to observe flight details
                continue
            elif (
                t_response
                > t_max
                + self._rid_version.realtime_period
                + self._config.max_propagation_latency.timedelta
            ):
                # This flight should not have been observed (it was too far in the past)
                if matching_flights:
                    self._test_scenario.record_failed_check(
                        name="Lingering flight",
                        summary="Flight still observed long after it ended",
                        details=f"Flight {flight_id} injected into {expected_flight.uss_participant_id} was observed by {observer.participant_id} at {t_response.isoformat()} after it ended at {t_max.isoformat()}",
                        severity=Severity.High,
                        relevant_participants=[
                            expected_flight.uss_participant_id,
                            observer.participant_id,
                        ],
                        query_timestamps=[
                            query.request.timestamp,
                            expected_flight.query_timestamp,
                        ],
                    )
                    continue
            elif (
                t_min + self._config.max_propagation_latency.timedelta
                < t_initiated
                < t_max + self._rid_version.realtime_period
            ):
                # This flight should definitely have been observed
                if not matching_flights:
                    self._test_scenario.record_failed_check(
                        name="Missing flight",
                        summary="Expected flight not observed",
                        details=f"Flight {flight_id} injected into {expected_flight.uss_participant_id} was not listed in the observation by {observer.participant_id} at {t_response.isoformat()} even though it should have been active from {t_min.isoformat()} to {t_max.isoformat()}",
                        severity=Severity.High,
                        relevant_participants=[
                            expected_flight.uss_participant_id,
                            observer.participant_id,
                        ],
                        query_timestamps=[
                            query.request.timestamp,
                            expected_flight.query_timestamp,
                        ],
                    )
                    continue
                # TODO: observe flight details
            elif t_initiated > t_min:
                # If this flight was not observed, there may be propagation latency
                pass  # TODO: findings propagation latency

            for matching_flight in matching_flights:
                pass  # TODO: Check position, altitude, flight details, etc

    def _evaluate_area_too_large_observation(
        self,
        observer: RIDSystemObserver,
        rect: s2sphere.LatLngRect,
        diagonal: float,
        query: fetch.Query,
    ) -> None:
        if query.status_code != 413:
            self._test_scenario.record_failed_check(
                name="Area too large",
                summary="Did not receive expected error code for too-large area request",
                details=f"{observer.participant_id} was queried for flights in {_rect_str(rect)} with a diagonal of {diagonal} which is larger than the maximum allowed diagonal of {self._rid_version.max_diagonal_km}.  The expected error code is 413, but instead code {query.status_code} was received.",
                severity=Severity.High,
                relevant_participants=[observer.participant_id],
                query_timestamps=[query.request.timestamp],
            )

    def _evaluate_clusters_observation(self) -> None:
        # TODO: Check cluster sizing, aircraft counts, etc
        pass
