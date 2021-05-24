from typing import Dict, List

from monitoring.rid_qualifier.mock import api, behavior


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
  tests: Dict[str, TestRecord]
  behavior: behavior.ServiceProviderBehavior

  def __init__(self):
    self.tests = {}
    self.behavior = behavior.ServiceProviderBehavior()


class RIDDP(object):
  behavior: behavior.DisplayProviderBehavior

  def __init__(self):
    self.behavior = behavior.DisplayProviderBehavior()


class Database(object):
  """Simple in-memory pseudo-database tracking the state of the mock system"""
  sps: Dict[str, RIDSP]
  dps: Dict[str, RIDDP]

  def __init__(self):
    self.sps = {}
    self.dps = {}

  def get_sp(self, sp_id: str) -> RIDSP:
    if sp_id not in self.sps:
      self.sps[sp_id] = RIDSP()
    return self.sps[sp_id]

  def get_dp(self, dp_id: str) -> RIDDP:
    if dp_id not in self.dps:
      self.dps[dp_id] = RIDDP()
    return self.dps[dp_id]


db = Database()
