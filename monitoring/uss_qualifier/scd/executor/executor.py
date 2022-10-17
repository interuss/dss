import itertools
import json
import os
import typing
from typing import Dict, List

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.scd.configuration import (
    SCDQualifierTestConfiguration,
    InjectionTargetConfiguration,
)
from monitoring.uss_qualifier.resources.flight_planning.automated_test import (
    AutomatedTest,
    TestStep,
)
from monitoring.uss_qualifier.scd.data_interfaces import (
    AutomatedTestContext,
)
from monitoring.uss_qualifier.scd.executor.errors import TestRunnerError
from monitoring.uss_qualifier.scd.executor.runner import TestRunner
from monitoring.uss_qualifier.scd.executor.target import TestTarget
from monitoring.uss_qualifier.scd.reports import Report
from monitoring.uss_qualifier.utils import is_url


def load_scd_test_definitions() -> Dict[str, AutomatedTest]:
    """Gets automated tests"""

    # TODO: Get test definitions via Resource rather than hardcoding here
    tests = {
        "u-space/flight-authorisation-validation-1": "test_data/che/flight_planning/flight-authorisation-validation-1.json",
        "astm-strategic-coordination/nominal-planning-1": "test_data/che/flight_planning/nominal-planning-1.json",
        "astm-strategic-coordination/nominal-planning-priority-1": "test_data/che/flight_planning/nominal-planning-priority-1.json",
    }
    automated_tests: Dict[str, AutomatedTest] = {}
    for k, v in tests.items():
        with open(v, "r") as f:
            automated_tests[k] = ImplicitDict.parse(json.load(f), AutomatedTest)

    return automated_tests


def validate_configuration(test_configuration: SCDQualifierTestConfiguration):
    try:
        for injection_target in test_configuration.injection_targets:
            is_url(injection_target.injection_base_url)
    except ValueError:
        raise ValueError("A valid url for injection_target must be passed")


def combine_targets(
    targets: List[TestTarget], steps: List[TestStep]
) -> typing.Iterator[Dict[str, TestTarget]]:
    """Gets combination of targets assigned to the uss roles specified in steps"""

    injection_steps = filter(lambda step: "inject_flight" in step, steps)

    # Get unique uss roles in injection steps in deterministic order
    uss_roles = sorted(
        set(
            map(
                lambda step: step.inject_flight.injection_target.uss_role,
                injection_steps,
            )
        )
    )

    targets_count = len(targets)
    uss_roles_count = len(uss_roles)
    if targets_count < uss_roles_count:
        # TODO: Implement a strategy when there are less targets configured than the required uss_roles.
        raise RuntimeError(
            "A minimum of {} targets have to be configured for this test. Only {} found.".format(
                uss_roles_count, targets_count
            )
        )

    # Create combinations
    for t in itertools.permutations(targets, len(uss_roles)):
        target_set = {}
        for i, role in enumerate(uss_roles):
            target_set[role] = t[i]
        yield target_set


def format_combination(combination: Dict[str, TestTarget]) -> List[str]:
    """Returns a string in the form of `uss_role: target_name`"""
    return list(map(lambda t: "{}: {}".format(t[0], t[1].name), combination.items()))


def targets_information(targets: List[TestTarget]):
    return dict(
        map(lambda target: (target.name, target.get_target_information()), targets)
    )


def run_scd_tests(
    test_configuration: SCDQualifierTestConfiguration,
    auth_spec: str,
) -> bool:
    locale = "CHE"  # TODO: Obtain from configuration instead
    automated_tests = load_scd_test_definitions()
    configured_targets = list(
        map(
            lambda t: TestTarget(t.name, t, auth_spec),
            test_configuration.injection_targets,
        )
    )
    dss_target = (
        TestTarget(
            "DSS",
            InjectionTargetConfiguration(
                name="DSS", injection_base_url=test_configuration.dss_base_url
            ),
            auth_spec=auth_spec,
        )
        if "dss_base_url" in test_configuration
        else None
    )
    report = Report(
        qualifier_version=os.environ.get("USS_QUALIFIER_VERSION", "unknown"),
        configuration=test_configuration,
        targets_information=targets_information(configured_targets),
    )

    should_exit = False
    executed_test_run_count = 0
    for test_id, test in automated_tests.items():
        if should_exit:
            break
        target_combinations = combine_targets(configured_targets, test.steps)
        for i, targets_under_test in enumerate(target_combinations):
            context = AutomatedTestContext(
                test_id=test_id,
                test_name=test.name,
                locale=locale,
                targets_combination=dict(
                    map(lambda t: (t[0], t[1].name), targets_under_test.items())
                ),
            )
            print(
                "[SCD] Starting test combination {}: {} ({}/{}) {}".format(
                    i + 1,
                    test.name,
                    locale,
                    test_id,
                    format_combination(targets_under_test),
                )
            )

            runner = TestRunner(context, test, targets_under_test, dss_target, report)
            try:
                runner.run_automated_test()
            except TestRunnerError as e:
                report.findings.issues.append(e.issue)
                print(
                    "[SCD] TestRunnerError: {} Issue: {} Related interactions: {}".format(
                        e, e.issue.details, e.issue.interactions
                    )
                )
            finally:
                runner.teardown()

            executed_test_run_count = executed_test_run_count + 1
            should_exit = len(report.findings.critical_issues()) > 0
            if should_exit:
                print(
                    "[SCD] Critical issues found during test. Interrupting test sequence. {}".format(
                        report.findings
                    )
                )
                break

    report.save()
    return report, executed_test_run_count


def check_scd_test_run_issues(report, executed_test_run_count):
    issues_count = len(report.findings.issues)
    outcome = "SUCCESS" if issues_count == 0 else "FAIL"
    print(
        "[SCD] Result: {} {} {} executed tests".format(
            outcome, report.findings, executed_test_run_count
        )
    )

    # TODO: handle low priority issues.
    return issues_count == 0
