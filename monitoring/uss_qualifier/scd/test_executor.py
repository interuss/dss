import json
import os
from pathlib import Path
from typing import List

from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest
from monitoring.uss_qualifier.scd.utils import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.utils import is_url


def get_automated_tests(automated_tests_dir: Path) -> List[AutomatedTest]:
    """Gets automated tests from the specified directory if they exist"""

    if not os.path.exists(automated_tests_dir):
        raise ValueError('The automated tests directory does not exist: {}'.format(automated_tests_dir))

    all_files = os.listdir(automated_tests_dir)
    files = [os.path.join(automated_tests_dir, f) for f in all_files if os.path.isfile(os.path.join(automated_tests_dir, f))]
    if not files:
        raise ValueError('There are no automated tests in the directory, create automated tests first using the simulator module.')

    automated_tests = []
    for file in files:
        with open(file, 'r') as f:
            automated_tests.append(ImplicitDict.parse(json.load(f), AutomatedTest))

    return automated_tests

def load_scd_test_definitions(locale: str) -> List[AutomatedTest]:
    automated_tests_dir = Path(os.getcwd(), 'scd/test_definitions', locale.value, 'automated_test')
    try:
        automated_tests = get_automated_tests(automated_tests_dir)
    except ValueError:
        print('[SCD] No automated tests files found; generating them via simulator now')
        # TODO: Call the simulator
        raise NotImplementedError()
    return automated_tests

def validate_configuration(test_configuration: SCDQualifierTestConfiguration):
    try:
        for injection_target in test_configuration.injection_targets:
            is_url(injection_target.injection_base_url)
    except ValueError:
        raise ValueError("A valid url for injection_target must be passed")

def run_scd_tests(locale: Locality, test_configuration: SCDQualifierTestConfiguration,
                  auth_spec: str):

    automated_tests = load_scd_test_definitions(locale)

    # TODO: Move to simulator
    if locale.is_uspace_applicable():
        print("[SCD] U-Space tests")

    print("[SCD] Running ASTM tests")
    # TODO Replace with actual implementation
