#!env/bin/python3

import json
import os
import sys
import argparse
from urllib.parse import urlparse
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import RIDQualifierTestConfiguration
import monitoring.uss_qualifier.rid.test_executor as test_executor

def is_url(url_string):
    try:
        urlparse(url_string)
    except ValueError:
        raise ValueError("A valid injection_url must be passed")

def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Exceute RID_Qualifier for a locale")

    parser.add_argument(
        "--auth",
        required = True,
        help="Auth spec for obtaining authorization to DSS instances; see README.md")

    parser.add_argument(
        "--config",
        required=True,
        help="Configuration of test to be conducted; either JSON describing a utils.RIDQualifierTestConfig, or the name of a file with that content")

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
    config: RIDQualifierTestConfiguration = ImplicitDict.parse(config_json, RIDQualifierTestConfiguration)

    # Validate configuration
    for injection_target in config.injection_targets:
        is_url(injection_target.injection_base_url)

    # Run test
    test_executor.main(test_configuration=config, auth_spec=auth_spec)

    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
