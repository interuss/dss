#!env/bin/python3

import argparse
import json
import os
import sys

from monitoring.uss_qualifier.configurations.configuration import TestConfiguration
from monitoring.uss_qualifier.reports.report import TestRunReport
from monitoring.uss_qualifier.resources.resource import create_resources
from monitoring.uss_qualifier.suites.suite import (
    TestSuite,
)


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute USS Qualifier")

    parser.add_argument(
        "--config",
        required=True,
        help="Configuration string according to monitoring/uss_qualifier/configurations/README.md",
    )

    return parser.parse_args()


def uss_test_executor(config: str):
    test_config = TestConfiguration.from_string(config)
    resources = create_resources(test_config.resources.resource_declarations)
    suite = TestSuite(test_config.test_suite, resources)
    report = suite.run()
    if report.successful:
        print("Final result: SUCCESS")
    else:
        print("Final result: FAILURE")

    return TestRunReport(configuration=test_config, report=report)


def main() -> int:
    args = parseArgs()

    reports = uss_test_executor(args.config)
    with open("report.json", "w") as f:
        json.dump(reports, f, indent=2)
    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
