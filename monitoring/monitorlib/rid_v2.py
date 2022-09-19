import datetime
from typing import Literal

from monitoring.monitorlib.typing import ImplicitDict, StringBasedDateTime
from . import rid as rid_v1


ISA_PATH = '/dss/identification_service_areas'
SUBSCRIPTION_PATH = '/dss/subscriptions'
SCOPE_DP = 'rid.display_provider'
SCOPE_SP = 'rid.service_provider'


class Time(ImplicitDict):
    value: StringBasedDateTime
    format: Literal['RFC3339']

    @classmethod
    def make(cls, t: datetime.datetime):
        return Time(format='RFC3339', value=t.strftime(DATE_FORMAT))


class Altitude(ImplicitDict):
    reference: Literal['WGS84']
    units: Literal['M']
    value: float

    @classmethod
    def make(cls, altitude_meters: float):
        return Altitude(reference='WGS84', units='M', value=altitude_meters)


MAX_SUB_PER_AREA = rid_v1.MAX_SUB_PER_AREA
MAX_SUB_TIME_HRS = rid_v1.MAX_SUB_TIME_HRS
DATE_FORMAT = rid_v1.DATE_FORMAT
NetMaxNearRealTimeDataPeriod = rid_v1.NetMaxNearRealTimeDataPeriod
NetMaxDisplayAreaDiagonal = 7  # km
NetDetailsMaxDisplayAreaDiagonal = 2  # km
geo_polygon_string = rid_v1.geo_polygon_string
