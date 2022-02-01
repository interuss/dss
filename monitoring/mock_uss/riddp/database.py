from typing import Dict

from monitoring.monitorlib.typing import ImplicitDict


class FlightInfo(ImplicitDict):
  flights_url: str


class Database(object):
  """Simple in-memory pseudo-database tracking the state of the mock system"""
  flights: Dict[str, FlightInfo]

  def __init__(self):
    self.flights = {}


db = Database()
