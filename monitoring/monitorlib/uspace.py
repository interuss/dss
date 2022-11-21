from typing import List
from urllib.parse import urlparse

from monitoring.monitorlib.scd_automated_testing import scd_injection_api
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    FlightAuthorisationData,
)


def problems_with_flight_authorisation(
    flight_auth: FlightAuthorisationData,
) -> List[str]:
    problems: List[str] = []
    if not flight_auth.uas_serial_number.valid:
        problems.append("Invalid serial number")
    if not flight_auth.operator_id.valid:
        problems.append("Invalid operator ID")
    if flight_auth.uas_class == scd_injection_api.UASClass.Other:
        problems.append("Invalid UAS class")
    if flight_auth.operation_category == scd_injection_api.OperationCategory.Unknown:
        problems.append("Invalid operation category")
    if (
        flight_auth.endurance_minutes < 1
        or flight_auth.endurance_minutes > 10 * 24 * 60
    ):
        problems.append("Invalid endurance")
    if sum(1 if len(m) > 0 else 0 for m in flight_auth.connectivity_methods) == 0:
        problems.append("Invalid connectivity methods")
    if (
        sum(1 if len(m) > 0 else 0 for m in flight_auth.identification_technologies)
        == 0
    ):
        problems.append("Invalid identification technologies")
    try:
        urlparse(flight_auth.emergency_procedure_url)
    except ValueError:
        problems.append("Invalid emergency procedure URL")
    return problems
