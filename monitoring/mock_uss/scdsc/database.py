import json
from typing import Dict

from monitoring.monitorlib import scd
from monitoring.monitorlib.multiprocessing import SynchronizedValue
from monitoring.monitorlib.scd_automated_testing import scd_injection_api
from monitoring.monitorlib.typing import ImplicitDict


class FlightRecord(ImplicitDict):
    """Representation of a flight in a USS"""
    op_intent_injection: scd_injection_api.OperationalIntentTestInjection
    flight_authorisation: scd_injection_api.FlightAuthorisationData
    op_intent_reference: scd.OperationalIntentReference


class Database(ImplicitDict):
    """Simple in-memory pseudo-database tracking the state of the mock system"""
    flights: Dict[str, FlightRecord] = {}
    cached_operations: Dict[str, scd.OperationalIntent] = {}


db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode('utf-8')), Database))
