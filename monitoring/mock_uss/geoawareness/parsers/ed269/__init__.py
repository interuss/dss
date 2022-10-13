import json
from enum import Enum
from typing import List, Any, Optional, Union, Dict, Tuple
from implicitdict import ImplicitDict, StringBasedDateTime, StringBasedTimeDelta


class Restriction(str, Enum):
    PROHIBITED = "PROHIBITED"
    REQ_AUTHORISATION = "REQ_AUTHORISATION"
    CONDITIONAL = "CONDITIONAL"
    NO_RESTRICTION = "NO_RESTRICTION"


class Reason(str, Enum):
    AIR_TRAFFIC = "AIR_TRAFFIC"
    SENSITIVE = "SENSITIVE"
    PRIVACY = "PRIVACY"
    POPULATION = "POPULATION"
    NATURE = "NATURE"
    NOISE = "NOISE"
    FOREIGN_TERRITORY = "FOREIGN_TERRITORY"
    EMERGENCY = "EMERGENCY"
    OTHER = "OTHER"


class YESNO(str, Enum):
    YES = "YES"
    NO = "NO"


class Purpose(str, Enum):
    AUTHORIZATION = "AUTHORIZATION"
    NOTIFICATION = "NOTIFICATION"
    INFORMATION = "INFORMATION"


class UASZoneAuthority(ImplicitDict):
    name: Optional[str]  # max length: 200
    service: Optional[str]  # max length: 200
    email: Optional[str]
    contactName: Optional[str]  # max length: 200
    siteURL: Optional[str]
    phone: Optional[str]  # max length: 200
    purpose: Optional[Purpose]
    intervalBefore: Optional[str]


class VerticalReferenceType(str, Enum):
    AGL = "AGL"
    AMSL = "AMSL"


# TODO: Discuss with @ben about Union[PolygonType, CircleType] which throws a not implemented exception
#  in implicitdict.
#
# class PolygonType(ImplicitDict):
#     type: str = "Polygon"
#     coordinates: List[ # min 4 items
#         List[ # 2 items
#             float
#         ]
#     ]
#
#
# class CircleType(ImplicitDict):
#     type: str = "Circle"
#     center: List[float] # 2 items
#     radius: float # > 0

# Start workaround
class UniversalHorizontalProjectionType(str, Enum):
    Circle = "Circle"
    Polygon = "Polygon"


class UniversalHorizontalProjection(ImplicitDict):
    type: str
    center: Optional[List[float]]  # 2 items
    radius: Optional[float]  # > 0
    coordinates: Optional[List[List[float]]]  # min 4 items  # 2 items


# End workaround


class UomDimensions(str, Enum):
    M = "M"
    FT = "FT"


class UASZoneAirspaceVolume(ImplicitDict):
    uomDimensions: UomDimensions
    lowerLimit: Optional[int]
    lowerVerticalReference: VerticalReferenceType
    upperLimit: Optional[int]
    upperVerticalReference: VerticalReferenceType
    horizontalProjection: UniversalHorizontalProjection  # TODO: Implement Union[PolygonType, CircleType] in implicitdict


class WeekDateType(str, Enum):
    MON = "MON"
    TUE = "TUE"
    WED = "WED"
    THU = "THU"
    FRI = "FRI"
    SAT = "SAT"
    SUN = "SUN"
    ANY = "ANY"


class DailyPeriod(ImplicitDict):
    day: List[WeekDateType]  # min items: 1, max items: 7
    startTime: str  # TODO: Convert to a Time representation
    endTime: str  # TODO: Convert to a Time representation


class ApplicableTimePeriod(ImplicitDict):
    permanent: YESNO
    startDateTime: Optional[StringBasedDateTime]
    endDateTime: Optional[StringBasedDateTime]
    schedule: Optional[List[DailyPeriod]]  # min items: 1


class UASZoneVersion(ImplicitDict):
    title: Optional[str]
    identifier: str  # max length: 7
    country: str  # length: 3
    name: Optional[str]  # max length: 200
    type: str
    restriction: Restriction
    restrictionConditions: Optional[List[str]]
    region: Optional[int]
    reason: Optional[List[Reason]]  # max length: 9
    otherReasonInfo: Optional[str]  # max length: 30
    regulationExemption: Optional[YESNO]
    uSpaceClass: Optional[str]  # max length: 100
    message: Optional[str]  # max length: 200
    applicability: List[ApplicableTimePeriod]
    zoneAuthority: List[UASZoneAuthority]
    geometry: List[UASZoneAirspaceVolume]  # min items: 1
    extendedProperties: Optional[Any]


class ED269Schema(ImplicitDict):
    title: Optional[str]
    description: Optional[str]
    features: List[UASZoneVersion]


def parse(raw_data: Dict) -> ED269Schema:
    return ImplicitDict.parse(raw_data, ED269Schema)
