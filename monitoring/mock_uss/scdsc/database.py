import dataclasses
from multiprocessing import Lock
from typing import Dict

from monitoring.monitorlib import scd
from monitoring.monitorlib.scd_automated_testing import scd_injection_api


@dataclasses.dataclass
class FlightRecord(object):
    """Representation of a flight in a USS"""
    op_intent_injection: scd_injection_api.OperationalIntentTestInjection
    flight_authorisation: scd_injection_api.FlightAuthorisationData
    op_intent_reference: scd.OperationalIntentReference


class Database(object):
    """Simple in-memory pseudo-database tracking the state of the mock system"""
    lock: Lock
    flights: Dict[str, FlightRecord]
    cached_operations: Dict[str, scd.OperationalIntent]

    def __init__(self):
        self.lock = Lock()
        self.flights = {}
        self.cached_operations = {}


db = Database()
