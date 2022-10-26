#!env/bin/python3

import argparse
import json
import os
import sys

from implicitdict import ImplicitDict
from monitoring.monitorlib.versioning import get_code_version
from monitoring.uss_qualifier.configurations.configuration import TestConfiguration
from monitoring.uss_qualifier.reports.documents import render_requirement_table
from monitoring.uss_qualifier.reports.graphs import make_graph
from monitoring.uss_qualifier.reports.report import TestRunReport
from monitoring.uss_qualifier.resources.resource import create_resources
from monitoring.uss_qualifier.scenarios.documentation.requirements import (
    evaluate_requirements,
)
from monitoring.uss_qualifier.suites.suite import (
    TestSuite,
)


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute USS Qualifier")

    parser.add_argument(
        "--config",
        help="Configuration string according to monitoring/uss_qualifier/configurations/README.md",
    )

    parser.add_argument(
        "--report",
        help="File name of the report to write (if --config provided) or read (if --config not provided)",
    )

    parser.add_argument(
        "--dot",
        help="File name to create for a GraphViz dot text file summarizing the test run",
    )

    parser.add_argument(
        "--tested_requirements",
        help="File name to create for a tested requirements HTML summary",
    )

    return parser.parse_args()


def uss_test_executor(config: str):
    codebase_version = get_code_version()
    test_config = TestConfiguration.from_string(config)
    resources = create_resources(test_config.resources.resource_declarations)
    suite = TestSuite(test_config.test_suite, resources)
    report = suite.run()
    if report.successful:
        print("Final result: SUCCESS")
    else:
        print("Final result: FAILURE")

    return TestRunReport(
        codebase_version=codebase_version, configuration=test_config, report=report
    )


def main() -> int:
    args = parseArgs()

    if args.config is not None:
        report = uss_test_executor(args.config)
        if args.report is not None:
            print(f"Writing report to {args.report}")
            with open(args.report, "w") as f:
                json.dump(report, f, indent=2)
    elif args.report is not None:
        with open(args.report, "r") as f:
            report = ImplicitDict.parse(json.load(f), TestRunReport)
    else:
        raise ValueError("No input provided; --config or --report must be specified")

    if args.dot is not None:
        print(f"Writing GraphViz dot source to {args.dot}")
        with open(args.dot, "w") as f:
            f.write(make_graph(report).source)

    if args.tested_requirements is not None:
        print(f"Writing tested requirements summary to {args.tested_requirements}")
        requirements = evaluate_requirements(report)
        with open(args.tested_requirements, "w") as f:
            f.write(render_requirement_table(requirements))

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
