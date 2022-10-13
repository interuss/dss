import json
from typing import Dict

from .behavior import DisplayProviderBehavior
from implicitdict import ImplicitDict
from monitoring.monitorlib.multiprocessing import SynchronizedValue


class FlightInfo(ImplicitDict):
    flights_url: str


class Database(ImplicitDict):
    """Simple pseudo-database structure tracking the state of the mock system"""

    flights: Dict[str, FlightInfo] = {}
    behavior: DisplayProviderBehavior = DisplayProviderBehavior()


db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode("utf-8")), Database),
)
