import argparse
import sys
from typing import List

import prober.auth


def parse_args(argv: List[str]):
  parser = argparse.ArgumentParser(description='Retrieve an access token')
  parser.add_argument(
    '--spec', action='store', dest='spec', type=str,
    help='The auth spec for which to retrieve the token.  See README.md in this folder for examples.')
  parser.add_argument(
    '--scopes', action='store', dest='scopes', type=str,
    help='The scope or scopes to request.  Multiple scopes should be space-separated (so, included in quotes on the command line).')
  parser.add_argument(
    '--audience', action='store', dest='audience', type=str,
    help='The audience to request.')
  return parser.parse_args(argv)


def get_access_token(spec: str, scopes: str, audience: str):
  adapter = prober.auth.make_auth_adapter(spec)
  return adapter.issue_token(audience, scopes.split(' '))


if __name__ == '__main__':
  args = parse_args(sys.argv[1:])
  print(get_access_token(args.spec, args.scopes, args.audience))

