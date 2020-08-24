#!env/bin/python3

import os
import sys
import argparse
from typing import Dict

from monitoring.monitorlib import auth, infrastructure
from .interop_test_suite import InterOpTestSuite


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Test Interoperability of DSSs")

    # When using Password OAuth flow, Username, Password, and Clients-id are
    # necessary for authentication
    parser.add_argument("--auth", help="Auth spec for obtaining authorization to DSS instances; see README.md")

    parser.add_argument(
        "DSS", help="List of URIs to DSS Servers. At least 2 DSSs", nargs="+"
    )

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
