import inspect
import os
import sys
from typing import Optional, Set, Type

from monitoring.monitorlib import inspection
from monitoring.uss_qualifier import scenarios
from monitoring.uss_qualifier.scenarios.documentation import validation


def main() -> int:
    inspection.import_submodules(scenarios)
    test_scenarios = scenarios.find_test_scenarios(scenarios)
    validation.validate(list(test_scenarios))
    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
