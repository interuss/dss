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
        help="Auth spec for obtaining authorization to DSS instances; see README.md")

    parser.add_argument(
        "--locale",
        help="A three letter ISO 3166 country code to run the qualifier against, this should be the same one used to simulate the flight_data in flight_data_generator.py module.")

    parser.add_argument(
        "--injection_url",
        help="A USS url where the test data is to be submitted")

    return parser.parse_args()



def main() -> int:
    args = parseArgs()
    
    auth_spec = args.auth
    locale = args.locale    
    injection_url = args.injection_url
    
    try:
        assert auth_spec is not None    
    except AssertionError as ae:
        print("No auth_spec provided, please provide a auth_spec as a string")
        sys.exit(0)

    try: 
        assert locale is not None
    except AssertionError as ae: 
        print("No locale provided, please provide a 3-letter ISO 3166 country code")
        sys.exit(0)
        
    try: 
        assert injection_url is not None
    except AssertionError as ae: 
        print("No injection_url provided, please provide a USS url where tests can be injected")
        sys.exit(0)
    
    is_url(injection_url)
    
    test_configuration = test_executor.build_test_configuration(locale = locale, auth_spec=auth_spec, injection_url= injection_url)
    
    asyncio.run(test_executor.main(test_configuration=test_configuration))
    
    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
