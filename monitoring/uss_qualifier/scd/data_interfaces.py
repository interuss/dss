from typing import Dict
from monitoring.monitorlib.locality import Locality
from implicitdict import ImplicitDict


class AutomatedTestContext(ImplicitDict):
    test_id: str
    """ID of test"""

    test_name: str
    """Name of test"""

    locale: Locality
    """Locale of test"""

    targets_combination: Dict[str, str]
    """Mapping of target role and target name used for this test."""
