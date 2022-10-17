#!env/bin/python3

import argparse
import json
import os
import sys

import yaml

from monitoring.monitorlib.locality import Locality
from implicitdict import ImplicitDict
from monitoring.uss_qualifier.configurations.configuration import TestConfiguration
from monitoring.uss_qualifier.reports import FailedCheck
from monitoring.uss_qualifier.scd.executor import executor as scd_test_executor
from monitoring.uss_qualifier.suites.suite import (
    TestSuiteAction,
    TestSuiteActionDeclaration,
    ReactionToFailure,
)
from monitoring.uss_qualifier.utils import USSQualifierTestConfiguration


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute USS Qualifier for a locale")

    parser.add_argument(
        "--config",
        required=True,
        help="Configuration of test to be conducted; either JSON describing a utils.USSQualifierTestConfig, or the name of a file with that content",
    )

    return parser.parse_args()


def _print_failed_check(failed_check: FailedCheck) -> None:
    print("New failed check:")
    yaml_lines = yaml.dump(json.loads(json.dumps(failed_check))).split("\n")
    print("\n".join("  " + line for line in yaml_lines))


def uss_test_executor(config: USSQualifierTestConfiguration):
    if config.config:
        test_config = TestConfiguration.from_string(config.config)
        resources = test_config.resources.create_resources()
        suite_action = TestSuiteAction(
            TestSuiteActionDeclaration(
                test_suite=test_config.test_suite, on_failure=ReactionToFailure.Continue
            ),
            resources,
        )
        action_report = suite_action.run()

        legacy_reports = {"suite": action_report.test_suite}
    else:
        legacy_reports = {}

    # TODO: Convert SCD tests into new architecture
    if "resources" in config:
        scd_test_report, executed_test_run_count = scd_test_executor.run_scd_tests(
            test_configuration=config
        )
        legacy_reports["scd"] = {
            "report": scd_test_report,
            "executed_test_run_count": executed_test_run_count,
        }
    else:
        legacy_reports["scd"] = {}
        print("[SCD] No configuration provided.")
    return legacy_reports


def main() -> int:
    args = parseArgs()

    # Load/parse configuration
    config_input = args.config
    if config_input.lower().endswith(".json"):
        with open(config_input, "r") as f:
            config_json = json.load(f)
    else:
        config_json = json.loads(config_input)
    config: USSQualifierTestConfiguration = ImplicitDict.parse(
        config_json, USSQualifierTestConfiguration
    )
    reports = uss_test_executor(config)
    with open("report.json", "w") as f:
        json.dump(reports, f, indent=2)
    scd_report = reports["scd"].get("report")
    executed_test_run_count = reports["scd"].get("executed_test_run_count")
    if (
        scd_report
        and executed_test_run_count
        and (
            not scd_test_executor.check_scd_test_run_issues(
                scd_report, executed_test_run_count
            )
        )
    ):
        return os.EX_SOFTWARE
    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
