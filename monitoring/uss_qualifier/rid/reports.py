import datetime
from typing import List, Optional

import s2sphere

from monitoring.monitorlib import fetch
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import (
    InjectedFlight,
    RIDQualifierTestConfiguration,
)
from monitoring.uss_qualifier.common_data_definitions import Severity


class Issue(ImplicitDict):
    timestamp: Optional[str]
    """Time the issue was discovered"""

    test_code: str
    """Code corresponding to check generating this issue"""

    relevant_requirements: List[str] = []
    """Requirements that this issue relates to"""

    severity: Severity
    """How severe the issue is"""

    injection_target: str
    """Issue is related to data injected into this target.

  This is generally a Service Provider for RID.
  """

    observation_source: str
    """Issue is related to the system state observed using this target.

  This is a Display Application/Provider for RID.
  """

    subject: Optional[str]
    """Identifier of the subject of this issue, if applicable.

  This may be a flight ID, or ID of other object central to the issue.
  """

    summary: str
    """Human-readable summary of the issue"""

    details: str
    """Human-readable description of the issue"""

    queries: List[fetch.Query]
    """Description of HTTP requests relevant to this issue"""

    def __init__(self, **kwargs):
        super(Issue, self).__init__(**kwargs)
        if "timestamp" not in kwargs:
            self.timestamp = datetime.datetime.utcnow().isoformat()


class Setup(ImplicitDict):
    configuration: RIDQualifierTestConfiguration
    injections: List[fetch.Query] = []


class Findings(ImplicitDict):
    issues: List[Issue] = []
    observation_queries: List[fetch.Query] = []

    def add_observation_query(self, query: fetch.Query):
        self.observation_queries.append(query)

    def add_area_too_large_not_indicated(
        self, observer_name: str, diagonal: float, query: fetch.Query
    ) -> None:
        self.issues.append(
            Issue(
                test_code="AREA_TOO_LARGE_NOT_INDICATED",
                relevant_requirements=["NET0430"],
                severity=Severity.High,
                injection_target="N/A",
                observation_source=observer_name,
                summary='Expected "Area too large" response not received',
                details="An area with {} km diagonal was queried and {} responded with {} rather than 413".format(
                    diagonal, observer_name, query.status_code
                ),
                queries=[query],
            )
        )

    def add_duplicate_flights(
        self,
        observer_name: str,
        flight_id: str,
        flight_count: int,
        injection_target: str,
        query: fetch.Query,
    ) -> None:
        self.issues.append(
            Issue(
                test_code="DUPLICATE_FLIGHTS",
                severity=Severity.Critical,
                injection_target=injection_target,
                observation_source=observer_name,
                subject=flight_id,
                summary="Found multiple flights with same ID",
                details="Found {} flights with ID {} when {} was queried".format(
                    flight_count, flight_id, observer_name
                ),
                queries=[query],
            )
        )

    def add_lingering_flight(
        self,
        observer_name: str,
        flight_id: str,
        t_max: datetime.datetime,
        t_initiated: datetime.datetime,
        injection_target: str,
        query: fetch.Query,
    ) -> None:
        self.issues.append(
            Issue(
                test_code="LINGERING_FLIGHT",
                severity=Severity.High,
                relevant_requirements=["NET0260", "NET0270"],
                injection_target=injection_target,
                observation_source=observer_name,
                subject=flight_id,
                summary="Lingering flight still observed after completion",
                details="Flight {} ended at {} but it was still observed when queried at {} by {}".format(
                    flight_id, t_max, t_initiated, observer_name
                ),
                queries=[query],
            )
        )

    def add_missing_flight(
        self,
        observer_name: str,
        injected_flight: InjectedFlight,
        rect: s2sphere.LatLngRect,
        injection_target: str,
        query: fetch.Query,
    ) -> None:
        flight_id = injected_flight.flight.details_responses[0].details.id
        timestamp = datetime.datetime.utcnow()  # TODO: Use query timestamp instead
        span = injected_flight.flight.get_span()
        self.issues.append(
            Issue(
                test_code="MISSING_FLIGHT",
                relevant_requirements=["NET0610"],
                severity=Severity.Critical,
                injection_target=injection_target,
                observation_source=observer_name,
                subject=flight_id,
                summary="Expected flight not found",
                details="Flight {} (from {} to {}) was expected inside ({}, {})-({}, {}) when {} was queried at {}".format(
                    flight_id,
                    span[0],
                    span[1],
                    rect.lo().lat().degrees,
                    rect.lo().lng().degrees,
                    rect.hi().lat().degrees,
                    rect.hi().lng().degrees,
                    observer_name,
                    timestamp,
                ),
                queries=[query],
            )
        )

    def add_observation_failure(
        self, observer_name: str, rect: s2sphere.LatLngRect, query: fetch.Query
    ) -> None:
        self.issues.append(
            Issue(
                test_code="OBSERVATION_FAILED",
                severity=Severity.Critical,
                injection_target="N/A",
                observation_source=observer_name,
                summary="Observation attempt failed",
                details="When queried for flights in ({}, {})-({}, {}), observer returned an invalid response with code {}".format(
                    rect.lo().lat().degrees,
                    rect.lo().lng().degrees,
                    rect.hi().lat().degrees,
                    rect.hi().lng().degrees,
                    query.status_code,
                ),
                queries=[query],
            )
        )

    def add_premature_flight(
        self,
        observer_name: str,
        flight_id: str,
        t_min: datetime.datetime,
        t_response: datetime.datetime,
        injection_target: str,
        query: fetch.Query,
    ) -> None:
        self.issues.append(
            Issue(
                test_code="PREMATURE_FLIGHT",
                severity=Severity.High,
                injection_target=injection_target,
                observation_source=observer_name,
                subject=flight_id,
                summary="Future flight visible before start time",
                details="Flight {} has first telemetry at {}, but it was already visible by {} at {}".format(
                    flight_id, t_min, t_response, observer_name
                ),
                queries=[query],
            )
        )

    def __repr__(self):
        return "[{} issues in {} observations]".format(
            len(self.issues), len(self.observation_queries)
        )


class Report(ImplicitDict):
    setup: Setup
    findings: Findings = Findings()
