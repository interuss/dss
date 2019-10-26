#!env/bin/python3

import os
import sys
import argparse
import clients
import datetime
import uuid
from interop_test_suite import InterOpTestSuite
from typing import Dict


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Test Interoperability of DSSs")
    parser.add_argument("OAuth", help="URI to the OAuth Server.")

    # When using Password OAuth flow, Username, Password, and Clients-id are
    # necessary for authentication
    parser.add_argument("--username", help="Username used to get OAuth Token")
    parser.add_argument("--password", help="Password used to get OAuth Token")
    parser.add_argument(
        "--client-id",
        help="Client ID used to get OAuth Token, used with Username and Password",
    )

    # When using Service Account OAuth flow, only the Service Account JSON File
    # is required to request Token.
    parser.add_argument(
        "--service-account",
        "--svc",
        help="Path to Service Account Credentials file used to get OAuth Token",
    )

    parser.add_argument(
        "DSS", help="List of URIs to DSS Servers. At least 2 DSSs", nargs="+"
    )

    return parser.parse_args()


def main() -> int:
    args = parseArgs()

    if args.service_account:
        oauth_client = clients.OAuthClient(
            args.OAuth,
            clients.AuthType.SERVICE_ACCOUNT,
            service_account_json=args.service_account,
        )
    elif args.username:
        assert args.password, "Password is required when using Username"
        assert args.client_id, "Client ID is required when authenticating with Password"
        oauth_client = clients.OAuthClient(
            args.OAuth,
            clients.AuthType.PASSWORD,
            username=args.username,
            password=args.password,
            client_id=args.client_id,
        )
    else:
        oauth_client = clients.OAuthClient(args.OAuth, clients.AuthType.NONE)
        oauth_client.parameterized_url = True

    dss_clients: Dict[str, clients.DSSClient] = {}
    for dss in args.DSS:
        dss_clients[dss] = clients.DSSClient(host=dss, oauth_client=oauth_client)

    # Begin Tests
    tests = InterOpTestSuite(dss_clients)
    tests.startTest()

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
