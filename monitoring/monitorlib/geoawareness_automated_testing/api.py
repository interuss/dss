from enum import Enum
from typing import List, Optional
from implicitdict import ImplicitDict, StringBasedDateTime
from uas_standards.eurocae_ed269 import (
    UomDimensions,
    VerticalReferenceType,
    Restriction,
)

SCOPE_GEOAWARENESS_TEST = "geo-awareness.test"

# Mirrors of types defined in geo-awareness automated testing API


class Position(ImplicitDict):
    uomDimensions: UomDimensions
    verticalReferenceType: VerticalReferenceType
    height: int
    longitude: float
    latitude: float


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


class ED269Filters(ImplicitDict):
    uSpaceClass: Optional[str]
    acceptableRestrictions: Optional[List[Restriction]]


class GeozonesFilterSet(ImplicitDict):
    position: Optional[Position]
    after: Optional[StringBasedDateTime]
    before: Optional[StringBasedDateTime]
    ed269: Optional[ED269Filters]


class GeozonesCheck(ImplicitDict):
    filterSets: List[GeozonesFilterSet]


class GeozonesCheckRequest(ImplicitDict):
    checks: List[GeozonesCheck]


class GeozonesCheckResultName(str, Enum):
    Present = "Present"
    Absent = "Absent"
    UnsupportedFilter = "UnsupportedFilter"
    Error = "Error"


class GeozonesCheckResult(ImplicitDict):
    geozone: GeozonesCheckResultName
    message: Optional[str]


class GeozonesCheckResponse(ImplicitDict):
    applicableGeozone: List[GeozonesCheckResult]
