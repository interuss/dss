import json
import os
from pathlib import Path
from typing import Dict

from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest
from monitoring.uss_qualifier.scd.utils import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.utils import is_url


def get_automated_tests(automated_tests_dir: Path, prefix: str) -> Dict[str, AutomatedTest]:
    """Gets automated tests from the specified directory"""

    # Read all JSON files in this directory
    automated_tests: Dict[str, AutomatedTest] = {}
    for file in automated_tests_dir.glob('*.json'):
        id = prefix + os.path.splitext(os.path.basename(file))[0]
        with open(file, 'r') as f:
            automated_tests[id] = ImplicitDict.parse(json.load(f), AutomatedTest)

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


def run_scd_tests(locale: Locality, test_configuration: SCDQualifierTestConfiguration,
                  auth_spec: str):

    automated_tests = load_scd_test_definitions(locale)

    for test_id, test in automated_tests.items():
        print('[SCD] Running {} ({})'.format(test_id, test.name))
        # TODO Replace with actual implementation
