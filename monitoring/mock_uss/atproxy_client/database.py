import json
from enum import Enum
from typing import Dict, Optional

from monitoring.monitorlib.multiprocessing import SynchronizedValue
from implicitdict import ImplicitDict


ATProxyWorkerID = int


class ATProxyWorkerState(str, Enum):
    NotStarted = "NotStarted"
    Starting = "Starting"
    Running = "Running"
    Stopping = "Stopping"
    Stopped = "Stopped"


class ATProxyWorker(ImplicitDict):
    pid: int
    """Process ID of this worker"""

    state: ATProxyWorkerState = ATProxyWorkerState.NotStarted
    """Execution state of this worker"""

    handling_request: Optional[str] = None
    """Request ID currently being handled by this worker"""


class Database(ImplicitDict):
    """Simple in-memory pseudo-database tracking the state of the mock system"""

    atproxy_workers: Dict[ATProxyWorkerID, ATProxyWorker]
    """Map of atproxy worker ID to information about that worker"""


db = SynchronizedValue(
    Database(atproxy_workers={}),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode("utf-8")), Database),
)
