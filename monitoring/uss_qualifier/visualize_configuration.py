#!env/bin/python3

import argparse
import json
import os
import sys

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.reports.graphs import make_graph
from monitoring.uss_qualifier.reports.report import TestRunReport


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Visualize a complete test run")

    parser.add_argument(
        "--report",
        required=True,
        help="Path to file containing a JSON representation of a TestRunReport",
    )

    return parser.parse_args()


def main() -> int:
    args = parseArgs()

    with open(args.report, "r") as f:
        report = ImplicitDict.parse(json.load(f), TestRunReport)
    print(make_graph(report).source)

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
