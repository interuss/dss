#!env/bin/python3

import argparse
import json
import os
import sys
from typing import Dict

from monitoring.monitorlib.locality import Locality
from implicitdict import ImplicitDict
from monitoring.uss_qualifier.rid import test_executor as rid_test_executor
from monitoring.uss_qualifier.rid import reports
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


def uss_test_executor(
    config, auth_spec, scd_test_definitions_path=None
) -> Dict[str, Dict[str, reports.Report]]:
    resources = config.resources.create_resources()

    test_executor = {"rid": {}, "scd": {}}
    if "rid" in config:
        test_executor["rid"].update(
            {
                "report": rid_test_executor.run_rid_tests(
                    resources=resources,
                    test_configuration=config.rid,
                )
            }
        )
    else:
        print("[RID] No configuration provided.")

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
        test_executor["scd"].update(
            {
                "report": scd_test_report,
                "executed_test_run_count": executed_test_run_count,
            }
        )
    else:
        print("[SCD] No configuration provided.")
    return test_executor


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
