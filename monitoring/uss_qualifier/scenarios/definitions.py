from typing import Dict

from implicitdict import ImplicitDict


class TestScenarioDeclaration(ImplicitDict):
    scenario_type: str
    """Type of test scenario, expressed as a Python class name qualified relative to this `scenarios` module"""

    resources: Dict[str, str] = {}
    """Mapping of resource parameter (additional argument to concrete test scenario constructor) to ID of resource to use"""
