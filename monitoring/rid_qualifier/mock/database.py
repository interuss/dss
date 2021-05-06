from typing import Dict, List

from monitoring.rid_qualifier.mock import api


class TestRecord(object):
  """Representation of RID SP's record of a set of injected test flights"""
  version: str
  flights: List[api.TestFlight]

  def __init__(self, version: str, flights: List[api.TestFlight]):
    flights = [api.TestFlight(**flight) for flight in flights]
    for flight in flights:
      flight.order_telemetry()
    self.version = version
    self.flights = flights


class RIDSP(object):
  tests: Dict[str, TestRecord] = {}


class Database(object):
  """Simple in-memory pseudo-database tracking the state of the mock system"""
  sps: Dict[str, RIDSP] = {}

  def __init__(self):
    self.sps = {}


db = Database()
