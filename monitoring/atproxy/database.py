from __future__ import annotations
import enum
import json
from typing import Dict, Optional

from monitoring.monitorlib.multiprocessing import SynchronizedValue
from implicitdict import ImplicitDict


# --- All queries ---
class QueryState(str, enum.Enum):
    """Whether a query is being handled, or has already been handled."""
    Queued = 'Queued'
    BeingHandled = 'BeingHandled'
    Complete = 'Complete'


class Query(ImplicitDict):
    """An incoming call to an automated testing endpoint."""

    type: str
    """The endpoint being queried.
    
    Uses the form <category>.<api>.<operation>, and is populated using the
    appropriate request's request_type_name() method.
    """

    request: dict
    """Information about the query request.
    
    Schema corresponds to the *Request object in requests.py whose
    request_type_name() matches `type`.
    """

    state: QueryState = QueryState.Queued

    return_code: Optional[int] = None
    """For a Query in the Complete `state`, the HTTP return code to return."""

    response: Optional[dict] = None
    """For a Query in the Complete `state`, the JSON response body to return."""


# --- Cross-process synchronization ---
class Database(ImplicitDict):
    """The body of data shared and synchronized across handler processes."""

    queries: Dict[str, Query] = {}
    """Mapping between query ID and query content."""


db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode('utf-8')), Database))
