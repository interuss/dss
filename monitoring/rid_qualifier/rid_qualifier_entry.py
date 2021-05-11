#!env/bin/python3

import os
import sys
import argparse
import asyncio
import json
from urllib.parse import urlparse
import monitoring.rid_qualifier.test_executor as test_executor 

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
        "--locale",
        required = True,
        help="A three letter ISO 3166 country code to run the qualifier against, this should be the same one used to simulate the flight_data in flight_data_generator.py module.")

    parser.add_argument(
        "--injection_base_url",
        required = True,
        help="A USS url where the test data is to be submitted")

    parser.add_argument(
        "--injection_suffix",
        help="A test_id and endpoint can be provided, if it not provided, qualifier will submit a test to /test/{test_id} endpoint ")

    return parser.parse_args()


def main() -> int:
    args = parseArgs()
    
    auth_spec = args.auth
    locale = args.locale    
    injection_base_url = args.injection_base_url
    injection_suffix = args.injection_suffix

    is_url(injection_base_url)
    
    test_configuration = test_executor.build_test_configuration(locale = locale, auth_spec=auth_spec, injection_base_url= injection_base_url, injection_suffix=injection_suffix)
    
    asyncio.run(test_executor.main(test_configuration=test_configuration))
    
    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
