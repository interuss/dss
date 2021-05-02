#!env/bin/python3

import os
import sys
import argparse
from typing import Dict

from monitoring.monitorlib import auth

def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Exceute RID_Qualifier for a locale")

    parser.add_argument(
        "--auth",
        help="Auth spec for obtaining authorization to DSS instances; see README.md")


    return parser.parse_args()


def main() -> int:
    args = parseArgs()

    adapter = auth.make_auth_adapter(args.auth)
    dss_clients: Dict[str, infrastructure.DSSTestSession] = {}
    for dss in args.DSS:
        dss_clients[dss] = infrastructure.DSSTestSession(dss, adapter)

    # Begin Tests
    tests = InterOpTestSuite(dss_clients)
    tests.startTest()

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
