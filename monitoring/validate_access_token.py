import argparse
import sys
from typing import List

import jwt

from monitoring.monitorlib import auth_validation


def parse_args(argv: List[str]):
    parser = argparse.ArgumentParser(description='Validate an access token')
    parser.add_argument(
        '--token', action='store', dest='token', type=str,
        help='JWT access token to be validated')
    parser.add_argument(
        '--key', action='store', dest='key', type=str,
        help='Public key to validate against.  May be PEM-format text or a URL to a JWKS or plaintext PEM file.')
    return parser.parse_args(argv)


def get_access_token_payload(token: str, public_key: str):
    key = auth_validation.fix_key(public_key)
    return jwt.decode(token, key, algorithms='RS256', options={'verify_aud': False})


if __name__ == '__main__':
    args = parse_args(sys.argv[1:])
    print(get_access_token_payload(args.token, args.key))
