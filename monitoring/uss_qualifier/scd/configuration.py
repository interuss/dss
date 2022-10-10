from typing import List, Optional

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration


class SCDQualifierTestConfiguration(ImplicitDict):
    injection_targets: List[InjectionTargetConfiguration]
    """Set of USS into which data should be injected"""

    dss_base_url: Optional[str]
    """Base URL of DSS serving the above targets, or blank to not perform DSS
    checks"""
