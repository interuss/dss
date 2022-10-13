import json
from typing import Dict, List, Optional

from monitoring.monitorlib.multiprocessing import SynchronizedValue
from monitoring.monitorlib.rid_automated_testing import injection_api
from implicitdict import ImplicitDict
from .behavior import ServiceProviderBehavior


class TestRecord(ImplicitDict):
    """Representation of RID SP's record of a set of injected test flights"""

    version: str
    flights: List[injection_api.TestFlight]
    isa_version: Optional[str] = None

    def __init__(self, **kwargs):
        kwargs["flights"] = [
            injection_api.TestFlight(**flight) for flight in kwargs["flights"]
        ]
        for flight in kwargs["flights"]:
            flight.order_telemetry()

        super(TestRecord, self).__init__(**kwargs)


class Database(ImplicitDict):
    """Simple pseudo-database structure tracking the state of the mock system"""

    tests: Dict[str, TestRecord] = {}
    behavior: ServiceProviderBehavior = ServiceProviderBehavior()


db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode("utf-8")), Database),
)
