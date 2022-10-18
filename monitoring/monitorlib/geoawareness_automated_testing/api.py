from enum import Enum
from typing import List, Optional
from implicitdict import ImplicitDict, StringBasedDateTime

SCOPE_GEOAWARENESS_TEST = "geo-awareness.test"

# Mirrors of types defined in geo-awareness automated testing API


class HarnessStatus(str, Enum):
    Starting = "Starting"
    Ready = "Ready"


class StatusResponse(ImplicitDict):
    status: HarnessStatus
    version: str


class GeozoneSourceState(str, Enum):
    Activating = "Activating"
    Ready = "Ready"
    Deactivating = "Deactivating"
    Unsupported = "Unsupported"
    Rejected = "Rejected"
    Error = "Error"


class GeozoneSourceResponse(ImplicitDict):
    result: GeozoneSourceState
    message: Optional[str]


class HttpsSourceFormat(str, Enum):
    Ed269 = "ED-269"


class GeozoneHttpsSource(ImplicitDict):
    url: str
    format: HttpsSourceFormat


class GeozoneSourceDefinition(ImplicitDict):
    https_source: GeozoneHttpsSource
