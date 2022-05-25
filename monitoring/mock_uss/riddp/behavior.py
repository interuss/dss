from typing import List, Optional

from monitoring.monitorlib.typing import ImplicitDict


ServiceProviderID = str


class DisplayProviderBehavior(ImplicitDict):
  always_omit_recent_paths: Optional[bool] = False
  do_not_display_flights_from: Optional[List[ServiceProviderID]] = []
