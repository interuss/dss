import os
import sys

from monitoring.monitorlib import inspection
from monitoring.uss_qualifier import scenarios
from monitoring.uss_qualifier.scenarios.documentation import validation
from monitoring.uss_qualifier.scenarios.scenario import find_test_scenarios


def main() -> int:
    inspection.import_submodules(scenarios)
    test_scenarios = find_test_scenarios(scenarios)
    validation.validate(list(test_scenarios))
    print("Test documentation is valid.")
    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
