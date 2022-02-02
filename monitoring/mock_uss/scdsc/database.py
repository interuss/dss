import dataclasses
from typing import Dict

from monitoring.monitorlib.scd_automated_testing import scd_injection_api


@dataclasses.dataclass
class FlightRecord(object):
    """Representation of a flight in a USS"""
    op_intent: scd_injection_api.OperationalIntentTestInjection
    flight_authorisation: scd_injection_api.FlightAuthorisationData
    op_intent_id: str


class Database(object):
    """Simple in-memory pseudo-database tracking the state of the mock system"""
    flights: Dict[str, FlightRecord]

    def __init__(self):
        self.flights = {}


db = Database()
