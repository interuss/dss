import os
import sys

from monitoring.monitorlib import inspection
from monitoring.uss_qualifier import scenarios
from monitoring.uss_qualifier.scenarios.documentation.autoformat import (
    format_scenario_documentation,
)
from monitoring.uss_qualifier.scenarios.scenario import find_test_scenarios


def main() -> int:
    inspection.import_submodules(scenarios)
    test_scenarios = find_test_scenarios(scenarios)
    changes = format_scenario_documentation(list(test_scenarios))
    for filename, content in changes.items():
        with open(filename, "w") as f:
            f.write(content)
        print(f"Reformatted documentation in {filename}")
    if not changes:
        print("No scenario documentation needs to be reformatted.")
    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
