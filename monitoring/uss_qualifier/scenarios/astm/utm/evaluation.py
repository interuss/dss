from datetime import timedelta
from typing import Optional

from monitoring.monitorlib import scd
from monitoring.monitorlib.scd import OperationalIntent, Volume4D

NUMERIC_PRECISION = 0.001


def validate_op_intent_details(
    operational_intent: OperationalIntent, expected_extent: Volume4D
) -> Optional[str]:
    # Check that the USS is providing reasonable details
    resp_vol4s = (
        operational_intent.details.volumes
        + operational_intent.details.off_nominal_volumes
    )
    resp_alts = scd.meter_altitude_bounds_of(resp_vol4s)
    resp_start = scd.start_of(resp_vol4s)
    resp_end = scd.end_of(resp_vol4s)
    error_text = None
    if resp_alts[0] > expected_extent.volume.altitude_lower.value + NUMERIC_PRECISION:
        error_text = "Lower altitude specified by USS in operational intent details ({} m WGS84) is above the lower altitude in the injected flight ({} m WGS84)".format(
            resp_alts[0], expected_extent.volume.altitude_lower.value
        )
    elif resp_alts[1] < expected_extent.volume.altitude_upper.value - NUMERIC_PRECISION:
        error_text = "Upper altitude specified by USS in operational intent details ({} m WGS84) is below the upper altitude in the injected flight ({} m WGS84)".format(
            resp_alts[1], expected_extent.volume.altitude_upper.value
        )
    elif resp_start > expected_extent.time_start.value.datetime + timedelta(
        seconds=NUMERIC_PRECISION
    ):
        error_text = "Start time specified by USS in operational intent details ({}) is past the start time of the injected flight ({})".format(
            resp_start.isoformat(), expected_extent.time_start.value
        )
    elif resp_end < expected_extent.time_end.value.datetime - timedelta(
        seconds=NUMERIC_PRECISION
    ):
        error_text = "End time specified by USS in operational intent details ({}) is prior to the end time of the injected flight ({})".format(
            resp_end.isoformat(), expected_extent.time_end.value
        )
    return error_text
