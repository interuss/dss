import itertools
import json
import os
import typing
from pathlib import Path
from typing import Dict, List

from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest, TestStep
from monitoring.uss_qualifier.scd.executor.runner import TestRunner
from monitoring.uss_qualifier.scd.executor.target import TestTarget
from monitoring.uss_qualifier.utils import is_url


def get_automated_tests(automated_tests_dir: Path, prefix: str) -> Dict[str, AutomatedTest]:
    """Gets automated tests from the specified directory"""

    # Read all JSON files in this directory
    automated_tests: Dict[str, AutomatedTest] = {}
    for file in automated_tests_dir.glob('*.json'):
        test_id = prefix + os.path.splitext(os.path.basename(file))[0]
        with open(file, 'r') as f:
            automated_tests[test_id] = ImplicitDict.parse(json.load(f), AutomatedTest)

    # Read subdirectories
    for subdir in automated_tests_dir.iterdir():
        if subdir.is_dir():
            new_tests = get_automated_tests(subdir, prefix + subdir.name + '/')
            for k, v in new_tests.items():
                automated_tests[k] = v

    return automated_tests


def load_scd_test_definitions(locale: Locality) -> Dict[str, AutomatedTest]:
    automated_tests_dir = Path(os.getcwd(), 'scd', 'test_definitions', locale.value)
    if not os.path.exists(automated_tests_dir):
        print('[SCD] No automated tests files found; generating them via simulator now')
        # TODO: Call the simulator
        raise NotImplementedError()

    return get_automated_tests(automated_tests_dir, '')


def validate_configuration(test_configuration: SCDQualifierTestConfiguration):
    try:
        for injection_target in test_configuration.injection_targets:
            is_url(injection_target.injection_base_url)
    except ValueError:
        raise ValueError("A valid url for injection_target must be passed")


def combine_targets(targets: List[TestTarget], steps: List[TestStep]) -> typing.Iterator[Dict[str, TestTarget]]:
    """Gets combination of targets assigned to the uss roles specified in steps"""
    injection_steps = filter(lambda step: 'inject_flight' in step, steps)

    # Get unique uss roles in injection steps in deterministic order
    uss_roles = sorted(set(map(lambda step: step.inject_flight.injection_target.uss_role, injection_steps)))

    # Create combinations
    for t in itertools.permutations(targets, len(uss_roles)):
        target_set = {}
        for i, role in enumerate(uss_roles):
            target_set[role] = t[i]
        yield target_set


def format_combination(combination: Dict[str, TestTarget]) -> List[str]:
    """Returns a string in the form of `uss_role: target_name`"""
    return list(map(lambda t: "{}: {}".format(t[0], t[1].name), combination.items()))


def run_scd_tests(locale: Locality, test_configuration: SCDQualifierTestConfiguration,
                  auth_spec: str):
    automated_tests = load_scd_test_definitions(locale)
    configured_targets = list(map(lambda t: TestTarget(t.name, t, auth_spec), test_configuration.injection_targets))

    for test_id, test in automated_tests.items():
        target_combinations = combine_targets(configured_targets, test.steps)
        for i, targets_under_test in enumerate(target_combinations):
            print('[SCD] Starting test combination {}: {} ({}/{}) {}'.format(i+1,  test.name, locale, test_id,
                format_combination(targets_under_test)))

            runner = TestRunner(test_id, test, targets_under_test)

            runner.run_automated_test()
            runner.teardown()
