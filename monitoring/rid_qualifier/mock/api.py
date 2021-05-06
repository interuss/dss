import datetime
from typing import List, Optional

import iso8601

from monitoring.monitorlib import rid
from monitoring.monitorlib.typing import ImplicitDict


# === Mirrors of types defined in remote ID automated testing injection API ===


class TestFlightDetails(ImplicitDict):
  effective_after: str
  details: rid.RIDFlightDetails

  def __init__(self, **kwargs):
    super(TestFlightDetails, self).__init__(**kwargs)
    try:
      iso8601.parse_date(self.effective_after)
    except iso8601.ParseError:
      raise ValueError('TestFlightDetails.effective_after value "{}" could not be parsed as date-time'.format(self.effective_after))


class TestFlight(ImplicitDict):
  injection_id: str
  telemetry: List[rid.RIDAircraftState]
  details_responses: List[TestFlightDetails]

  def get_details(self, t_now: datetime.datetime) -> Optional[TestFlightDetails]:
    latest_after: Optional[datetime.datetime] = None
    tf_details = None
    for response in self.details_responses:
      t_response = iso8601.parse_date(response.effective_after)
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
                            key=lambda telemetry: iso8601.parse_date(telemetry.timestamp))


class CreateTestParameters(ImplicitDict):
  requested_flights: List[TestFlight]


class ChangeTestResponse(ImplicitDict):
  injected_flights: List[TestFlight]
  version: str


class Position(ImplicitDict):
  lat: float
  lng: float
  alt: Optional[float]


class Path(ImplicitDict):
  positions: List[Position]


class Flight(ImplicitDict):
  id: str
  most_recent_position: Optional[Position]
  recent_paths: Optional[List[Path]]


class Cluster(ImplicitDict):
  corners: List[Position]
  area_sqm: float
  number_of_flights: int


class GetDisplayDataResponse(ImplicitDict):
  flights: Optional[List[Flight]]
  clusters: Optional[List[Cluster]]


class GetDetailsResponse(ImplicitDict):
  pass
