#!env/bin/python3

import json
import os
import sys
import argparse
from pathlib import Path
from urllib.parse import urlparse

from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import RIDQualifierTestConfiguration
from monitoring.uss_qualifier.rid.simulator import flight_state
from monitoring.uss_qualifier.rid import test_executor, aircraft_state_replayer

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
    
    parser.add_argument(
        "--flight-records",
        required=True,
        help="Path to flight records input files")

    return parser.parse_args()


def main() -> int:
    args = parseArgs()
    auth_spec = args.auth
    input_path = args.flight_records
    flight_records = [f for f in os.listdir(input_path) if f.endswith('.json')]
    file_objs = []
    for f in flight_records:
        filepath = f'{input_path}/{f}'
        with open(filepath) as fo:
            file_objs.append(fo.read())
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

    # Load aircraft state files
    aircraft_states_directory = Path(os.getcwd(), 'rid/test_definitions', config.locale, 'aircraft_states')
    try:
      flight_records = aircraft_state_replayer.get_full_flight_records(aircraft_states_directory)
    except ValueError:
      print('No aircraft state files found; generating them via simulator now')
      flight_state.generate_aircraft_states()
      flight_records = aircraft_state_replayer.get_full_flight_records(aircraft_states_directory)

    # Run RID tests
    test_executor.run_rid_tests(test_configuration=config, auth_spec=auth_spec,
                                flight_records=flight_records)

    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
