from datetime import timedelta
from enum import Enum

from monitoring.monitorlib import rid as rid_v1
from monitoring.monitorlib import rid_v2

# TODO(BenjaminPelletier): Rename current `rid.py` to `rid_v1.py`, then rename this file to `rid.py`


class RIDVersion(str, Enum):
    f3411_19 = "F3411-19"
    """ASTM F3411-19 (first version, v1)"""

    f3411_22a = "F3411-22a"
    """ASTM F3411-22a (second version, v2, API version 2.1)"""

    @property
    def read_scope(self) -> str:
        if self == RIDVersion.f3411_19:
            return rid_v1.SCOPE_READ
        elif self == RIDVersion.f3411_22a:
            return rid_v2.SCOPE_DP
        else:
            raise ValueError("Unsupported RID version '{}'".format(self))

    @property
    def realtime_period(self) -> timedelta:
        if self == RIDVersion.f3411_19:
            return rid_v1.NetMaxNearRealTimeDataPeriod
        elif self == RIDVersion.f3411_22a:
            return rid_v2.NetMaxNearRealTimeDataPeriod
        else:
            raise ValueError("Unsupported RID version '{}'".format(self))

    @property
    def max_diagonal_km(self) -> float:
        if self == RIDVersion.f3411_19:
            return rid_v1.NetMaxDisplayAreaDiagonal
        elif self == RIDVersion.f3411_22a:
            return rid_v2.NetMaxDisplayAreaDiagonal
        else:
            raise ValueError("Unsupported RID version '{}'".format(self))

    @property
    def max_details_diagonal_km(self) -> float:
        if self == RIDVersion.f3411_19:
            return rid_v1.NetDetailsMaxDisplayAreaDiagonal
        elif self == RIDVersion.f3411_22a:
            return rid_v2.NetDetailsMaxDisplayAreaDiagonal
        else:
            raise ValueError("Unsupported RID version '{}'".format(self))
