from typing import List, Optional

from implicitdict import ImplicitDict


class InjectionTargetConfiguration(ImplicitDict):
    """This object defines the data required for a uss"""

    name: str
    injection_base_url: str


class SCDQualifierTestConfiguration(ImplicitDict):
    injection_targets: List[InjectionTargetConfiguration]
    """Set of USS into which data should be injected"""

    dss_base_url: Optional[str]
    """Base URL of DSS serving the above targets, or blank to not perform DSS
    checks"""
