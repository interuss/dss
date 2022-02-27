#!env/bin/python3

import argparse
import json
import os
import sys

from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid import test_executor as rid_test_executor
from monitoring.uss_qualifier.scd.executor import executor as scd_test_executor
from monitoring.uss_qualifier.utils import USSQualifierTestConfiguration


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute USS Qualifier for a locale")

    parser.add_argument(
        "--auth",
        required = True,
        help="Auth spec for obtaining authorization to DSS instances; see README.md")

    parser.add_argument(
        "--config",
        required=True,
        help="Configuration of test to be conducted; either JSON describing a utils.USSQualifierTestConfig, or the name of a file with that content")

    return parser.parse_args()


def main() -> int:
    args = parseArgs()

    auth_spec = args.auth

    # Load/parse configuration
    config_input = args.config
    if config_input.lower().endswith('.json'):
        with open(config_input, 'r') as f:
          config_json = json.load(f)
    else:
        config_json = json.loads(config_input)
    config: USSQualifierTestConfiguration = ImplicitDict.parse(config_json, USSQualifierTestConfiguration)

    if "rid" in config:
        print(f"[RID] Configuration provided with {len(config.rid.injection_targets)} injection targets.")
        rid_test_executor.validate_configuration(config.rid)
        rid_flight_records = rid_test_executor.load_rid_test_definitions(config.locale)
        rid_test_executor.run_rid_tests(test_configuration=config.rid, auth_spec=auth_spec,
                                        flight_records=rid_flight_records)
    else:
        print("[RID] No configuration provided.")

    if "scd" in config:
        print(f"[SCD] Configuration provided with {len(config.scd.injection_targets)} injection targets.")
        scd_test_executor.validate_configuration(config.scd)

        locale = Locality(config.locale.upper())
        print(f"[SCD] Locale: {locale.value} (is_uspace_applicable:{locale.is_uspace_applicable}, allow_same_priority_intersections:{locale.allow_same_priority_intersections})")

        if not scd_test_executor.run_scd_tests(locale=locale, test_configuration=config.scd, auth_spec=auth_spec):
            return os.EX_SOFTWARE
    else:
        print("[SCD] No configuration provided.")

    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
