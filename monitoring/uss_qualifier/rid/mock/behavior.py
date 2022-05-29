from typing import List, Optional

from monitoring.monitorlib.typing import ImplicitDict


ServiceProviderID = str
DisplayProviderID = str


class ServiceProviderBehavior(ImplicitDict):
    switch_latitude_and_longitude_when_reporting: Optional[bool] = False
    use_agl_instead_of_wgs84_for_altitude: Optional[bool] = False
    use_feet_instead_of_meters_for_altitude: Optional[bool] = False


class DisplayProviderBehavior(ImplicitDict):
    always_omit_recent_paths: Optional[bool] = False
    do_not_display_flights_from: Optional[List[ServiceProviderID]] = []
