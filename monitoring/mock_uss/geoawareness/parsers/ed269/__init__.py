from datetime import time
from enum import Enum
from typing import List, Any, Optional, Dict, Union
import arrow

from implicitdict import ImplicitDict, StringBasedDateTime


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


class HorizontalProjectionType(str, Enum):
    Circle = "Circle"
    Polygon = "Polygon"


class CircleOrPolygonType(ImplicitDict):
    type: HorizontalProjectionType
    center: Optional[List[float]]  # 2 items
    radius: Optional[float]  # > 0
    coordinates: Optional[List[List[float]]]  # min 4 items  # 2 items


class UomDimensions(str, Enum):
    M = "M"
    FT = "FT"


class UASZoneAirspaceVolume(ImplicitDict):
    uomDimensions: UomDimensions
    lowerLimit: Optional[int]
    lowerVerticalReference: VerticalReferenceType
    upperLimit: Optional[int]
    upperVerticalReference: VerticalReferenceType
    horizontalProjection: CircleOrPolygonType


class WeekDateType(str, Enum):
    MON = "MON"
    TUE = "TUE"
    WED = "WED"
    THU = "THU"
    FRI = "FRI"
    SAT = "SAT"
    SUN = "SUN"
    ANY = "ANY"


class ED269TimeType(str):
    """String that allows values which describe a time in ED-269 flavour of ISO 8601 format.

    ED-269 standard specifies that a time instant type should be in the form of hh:mmS where S is
    the timezone. However, examples are using the following format: 00:00:00.00Z
    This class supports both formats as inputs and uses the long form as the output format.
    """

    time: time
    """`time` representation of the str value with timezone"""

    def __new__(cls, value: Union[str, time]):
        if isinstance(value, str):
            t = arrow.get(value, ["HH:mm:ss.SZ", "HH:mmZ"]).timetz()
        else:
            t = value
        str_value = str.__new__(
            cls, t.strftime("%H:%M:%S.%f")[:11] + t.strftime("%z").replace("+0000", "Z")
        )
        str_value.time = t
        return str_value


class DailyPeriod(ImplicitDict):
    day: List[WeekDateType]  # min items: 1, max items: 7
    startTime: ED269TimeType
    endTime: ED269TimeType


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

    @staticmethod
    def from_dict(raw_data: Dict) -> "ED269Schema":
        return ImplicitDict.parse(raw_data, ED269Schema)
