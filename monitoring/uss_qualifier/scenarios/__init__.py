import inspect
from typing import Optional, Set

from .scenario import TestScenario, TestScenarioType


def find_test_scenarios(
    module, already_checked: Optional[Set[str]] = None
) -> Set[TestScenarioType]:
    if already_checked is None:
        already_checked = set()
    already_checked.add(module.__name__)
    test_scenarios = set()
    for name, member in inspect.getmembers(module):
        if (
            inspect.ismodule(member)
            and member.__name__ not in already_checked
            and member.__name__.startswith("monitoring.uss_qualifier.scenarios")
        ):
            descendants = find_test_scenarios(member, already_checked)
            for descendant in descendants:
                if descendant not in test_scenarios:
                    test_scenarios.add(descendant)
        elif inspect.isclass(member) and member is not TestScenario:
            if issubclass(member, TestScenario):
                if member not in test_scenarios:
                    test_scenarios.add(member)
    return test_scenarios
