from typing import Dict, List

from monitoring.monitorlib.rid_automated_testing import injection_api


class TestRecord(object):
  """Representation of RID SP's record of a set of injected test flights"""
  version: str
  flights: List[injection_api.TestFlight]

  def __init__(self, version: str, flights: List[injection_api.TestFlight]):
    flights = [injection_api.TestFlight(**flight) for flight in flights]
    for flight in flights:
      flight.order_telemetry()
    self.version = version
    self.flights = flights


class Database(object):
  """Simple in-memory pseudo-database tracking the state of the mock system"""
  tests: Dict[str, TestRecord]

  def __init__(self):
    self.tests = {}


db = Database()
