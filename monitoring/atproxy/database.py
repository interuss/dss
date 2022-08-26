from __future__ import annotations
import enum
import json
from typing import Dict, Optional

from monitoring.monitorlib.multiprocessing import SynchronizedValue
from monitoring.monitorlib.typing import ImplicitDict


# --- All queries ---
class QueryState(str, enum.Enum):
    Queued = 'Queued'
    BeingHandled = 'BeingHandled'
    Complete = 'Complete'


class Query(ImplicitDict):
    type: str
    request: dict

    state: QueryState = QueryState.Queued

    return_code: Optional[int] = None
    response: Optional[dict] = None


# --- Cross-process synchronization ---
class Database(ImplicitDict):
    queries: Dict[str, Query] = {}


db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode('utf-8')), Database))
