from typing import List

from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration


class SCDQualifierTestConfiguration(ImplicitDict):
    injection_targets: List[InjectionTargetConfiguration]
    """Set of USS into which data should be injected"""
