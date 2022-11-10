"""
This script examines the repository content for good hygiene practices.

It will complete without error if no hygiene issues are found, or otherwise
exit with an error.
"""
import argparse
import os
import sys

from md_files.md_files import check_md_files


def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Check repo hygiene")

    parser.add_argument("repo_location", metavar="REPO", type=str, help="path to repository")

    return parser.parse_args()


def main() -> int:
    args = _parse_args()

    check_md_files(args.repo_location, args.repo_location)

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
