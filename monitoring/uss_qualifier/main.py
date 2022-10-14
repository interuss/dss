#!env/bin/python3

import argparse
import json
import os
import sys

from monitoring.monitorlib.locality import Locality
from implicitdict import ImplicitDict
from monitoring.uss_qualifier.scd.executor import executor as scd_test_executor
from monitoring.uss_qualifier.utils import USSQualifierTestConfiguration


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute USS Qualifier for a locale")

    parser.add_argument(
        "--auth",
        required=True,
        help="Auth spec for obtaining authorization to DSS instances; see README.md",
    )

    parser.add_argument(
        "--config",
        required=True,
        help="Configuration of test to be conducted; either JSON describing a utils.USSQualifierTestConfig, or the name of a file with that content",
    )

    return parser.parse_args()


def uss_test_executor(config, auth_spec, scd_test_definitions_path=None):
    # TODO: Harmonize and formalize report format shared between all scenarios
    ad_hoc_report = {"scd": {}}

    resources = config.resources.create_resources()
    scenarios = [s.make_test_scenario(resources) for s in config.scenarios]
    for i, scenario in enumerate(scenarios):
        ad_hoc_report["Scenario{}".format(i + 1)] = scenario.run()

    if "scd" in config:
        print(
            f"[SCD] Configuration provided with {len(config.scd.injection_targets)} injection targets."
        )
        scd_test_executor.validate_configuration(config.scd)

        locale = Locality(config.locale.upper())
        print(
            f"[SCD] Locale: {locale.value} (is_uspace_applicable:{locale.is_uspace_applicable}, allow_same_priority_intersections:{locale.allow_same_priority_intersections})"
        )

        scd_test_report, executed_test_run_count = scd_test_executor.run_scd_tests(
            locale=locale,
            test_configuration=config.scd,
            auth_spec=auth_spec,
            scd_test_definitions_path=scd_test_definitions_path,
        )
        ad_hoc_report["scd"] = {
            "report": scd_test_report,
            "executed_test_run_count": executed_test_run_count,
        }
    else:
        print("[SCD] No configuration provided.")
    return ad_hoc_report


def main() -> int:
    args = parseArgs()

    auth_spec = args.auth

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
    reports = uss_test_executor(config, auth_spec)
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
