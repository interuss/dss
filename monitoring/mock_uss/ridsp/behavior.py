from typing import Optional

from monitoring.monitorlib.rid_automated_testing.injection_api import TestFlight
from monitoring.monitorlib.rid import RIDFlight
from implicitdict import ImplicitDict


FEET_PER_METER = 1 / 0.3048


class ServiceProviderBehavior(ImplicitDict):
    switch_latitude_and_longitude_when_reporting: Optional[bool] = False
    use_agl_instead_of_wgs84_for_altitude: Optional[bool] = False
    use_feet_instead_of_meters_for_altitude: Optional[bool] = False


def adjust_reported_flight(flight: TestFlight, reported_flight: RIDFlight, behavior: ServiceProviderBehavior) -> RIDFlight:
    """Adjust how a flight is reported based on the SP behavior"""

    adjusted = ImplicitDict.parse(reported_flight, RIDFlight)
    if behavior.switch_latitude_and_longitude_when_reporting:
        if adjusted.has_field_with_value('current_state'):
            lng = adjusted.current_state.position.lat
            adjusted.current_state.position.lat = adjusted.current_state.position.lng
            adjusted.current_state.position.lng = lng
        if adjusted.has_field_with_value('recent_positions'):
            for p in adjusted.recent_positions:
                lng = p.position.lat
                p.position.lat = p.position.lng
                p.position.lng = lng

    if behavior.use_agl_instead_of_wgs84_for_altitude:
        ground_altitude = min(t.position.alt for t in flight.telemetry)
        if adjusted.has_field_with_value('current_state'):
            if adjusted.current_state.has_field_with_value('height'):
                adjusted.current_state.position.alt = adjusted.current_state.height.distance
            else:
                adjusted.current_state.position.alt -= ground_altitude

        if adjusted.has_field_with_value('recent_positions'):
            for p in adjusted.recent_positions:
                p.position.alt -= ground_altitude

    if behavior.use_feet_instead_of_meters_for_altitude:
        if adjusted.has_field_with_value('current_state'):
            adjusted.current_state.position.alt *= FEET_PER_METER
            if adjusted.current_state.has_field_with_value('height'):
                adjusted.current_state.height.distance *= FEET_PER_METER
        if adjusted.has_field_with_value('recent_positions'):
            for p in adjusted.recent_positions:
                p.position.alt *= FEET_PER_METER

    return adjusted
