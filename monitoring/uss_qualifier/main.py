#!env/bin/python3

import argparse
import json
import os
import sys

from implicitdict import ImplicitDict
from monitoring.monitorlib.versioning import get_code_version
from monitoring.uss_qualifier.configurations.configuration import (
    TestConfiguration,
    USSQualifierConfiguration,
    ArtifactsConfiguration,
)
from monitoring.uss_qualifier.reports.documents import generate_tested_requirements
from monitoring.uss_qualifier.reports.graphs import make_graph
from monitoring.uss_qualifier.reports.report import TestRunReport
from monitoring.uss_qualifier.resources.resource import create_resources
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
        default=None,
        help="(Overrides setting in artifacts configuration) File name of the report to write (if test configuration provided) or read (if test configuration not provided)",
    )

    return parser.parse_args()


def execute_test_run(config: TestConfiguration):
    codebase_version = get_code_version()
    resources = create_resources(config.resources.resource_declarations)
    suite = TestSuite(config.test_suite, resources)
    report = suite.run()
    if report.successful:
        print("Final result: SUCCESS")
    else:
        print("Final result: FAILURE")

    return TestRunReport(
        codebase_version=codebase_version, configuration=config, report=report
    )


def main() -> int:
    args = parseArgs()

    config = USSQualifierConfiguration.from_string(args.config)
    if args.report:
        if not config.artifacts:
            config.artifacts = ArtifactsConfiguration(report_path=args.report)
        else:
            config.artifacts.report_path = args.report

    if config.test_run:
        report = execute_test_run(config.test_run)
        if config.artifacts and config.artifacts.report_path:
            print(f"Writing report to {config.artifacts.report_path}")
            with open(config.artifacts.report_path, "w") as f:
                json.dump(report, f, indent=2)
    elif config.artifacts and config.artifacts.report_path:
        with open(config.artifacts.report_path, "r") as f:
            report = ImplicitDict.parse(json.load(f), TestRunReport)
    else:
        raise ValueError(
            "No input provided; test_run or artifacts.report_path must be specified in configuration"
        )

    if config.artifacts:
        if config.artifacts.graph:
            print(f"Writing GraphViz dot source to {config.artifacts.graph.gv_path}")
            with open(config.artifacts.graph.gv_path, "w") as f:
                f.write(make_graph(report).source)

        if config.artifacts.tested_roles:
            path = config.artifacts.tested_roles.report_path
            print(f"Writing tested roles summary to {path}")
            with open(path, "w") as f:
                f.write(
                    generate_tested_requirements(
                        report, config.artifacts.tested_roles.roles
                    )
                )

    return os.EX_OK


if __name__ == "__main__":
    sys.exit(main())
