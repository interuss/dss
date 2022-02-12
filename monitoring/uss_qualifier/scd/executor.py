import itertools
import json
import os
import typing
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List

from monitoring.monitorlib import infrastructure, auth
from monitoring.monitorlib.clients.scd import OperationError
from monitoring.monitorlib.clients.scd_automated_testing import create_flight, delete_flight
from monitoring.monitorlib.locality import Locality
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResult, \
    DeleteFlightResult, InjectFlightResponse
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import InjectionTargetConfiguration
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.data_interfaces import AutomatedTest, TestStep
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


class TestTarget():

    def __init__(self, name: str, config: InjectionTargetConfiguration, auth_spec: str):
        self.name = name
        self.config = config
        self.client = infrastructure.DSSTestSession(
            self.config.injection_base_url,
            auth.make_auth_adapter(auth_spec))
        self.created_flight_ids: Dict[str, str] = {}

    def __repr__(self):
        return "TestTarget({}, {})".format(self.name, self.config.injection_base_url)

    def inject_flight(self, flight_request: InjectFlightRequest, dry=False):
        flight_id, resp = create_flight(self.client, self.config.injection_base_url, flight_request.test_injection, dry=dry)
        print (flight_id, self.name, self.created_flight_ids)
        if resp.result in [InjectFlightResult.Planned, InjectFlightResult.DryRun]:
            self.created_flight_ids[flight_request.name] = flight_id
        # elif resp.result == InjectFlightResult.ConflictWithFlight:
        #     raise OperationError("Unable to inject flight due to conflicting flight: {}".format(resp))
        # elif resp.result == InjectFlightResult.Failed:
        #     raise OperationError("Unable to inject flight: {}".format(resp))
        return resp

    def delete_flight(self, flight_name: str, dry=False):
        flight_id = self.created_flight_ids[flight_name]
        resp = delete_flight(self.client, self.config.injection_base_url, flight_id, dry=dry)
        if resp.result in [DeleteFlightResult.Closed, InjectFlightResult.DryRun]:
            del self.created_flight_ids[flight_name]
        # TODO: Handle errors

    def delete_all_flights(self, dry=False) -> int:
        flights_count = len(self.created_flight_ids.keys())
        print("[SCD]    - Deleting {} flights for target {}.".format(flights_count, self.name))
        for flight_name, flight_id in list(self.created_flight_ids.items()):
            self.delete_flight(flight_name, dry=dry)
        return flights_count

    def has_created_flight(self, flight_name: str):
        return flight_name in self.created_flight_ids.keys()

class TestRunner:
    def __init__(self, auth_spec: str, automated_test_id: str, automated_test: AutomatedTest, targets: Dict[str, TestTarget]):
        self.auth_spec = auth_spec
        self.automated_test_id = automated_test_id
        self.automated_test = automated_test
        self.targets = targets

    def print_test_plan(self):
        self.run_automated_test(dry=True)
        self.teardown(dry=True)

    def run_automated_test(self, dry=False):
        for step in self.automated_test.steps:
            print('[SCD] - {}'.format(step.name))
            self.execute_step(step, dry=dry)

    def evaluate_inject_flight_response(self, req: InjectFlightRequest, resp: InjectFlightResponse, dry=False) -> bool:
        if dry and resp.result == InjectFlightResult.DryRun:
            print("[SCD] Result: SKIP")
            return
        if resp.result not in req.known_responses.acceptable_results:
            raise OperationError("[SCD] ERROR: Received {}, expected one of {}. Reason: {}".format(resp.result, req.known_responses.acceptable_results, resp.get('notes', None)))
        print("[SCD] Result: SUCCESS")

    def execute_step(self, step: TestStep, dry=False):
        target = self.get_target(step)
        if target is None:
            # TODO implement reporting
            self.print_targets_state()
            raise RuntimeError("[SCD] Error: Unable to identify the target managing flight {}".format(
                step.inject_flight.name if 'inject_flight' in step else step.delete_flight.flight_name
            ))

        if 'inject_flight' in step:
            print("[SCD]   - Inject flight {} to {}".format(step.inject_flight.name, target.name))
            resp = target.inject_flight(step.inject_flight, dry=dry)
            self.evaluate_inject_flight_response(step.inject_flight, resp, dry=dry)
        elif 'delete_flight' in step:
            print("[SCD]   - Delete flight {} to {}".format(step.delete_flight.flight_name, target.name))
            target.delete_flight(step.delete_flight.flight_name, dry=dry)
        else:
            print("[SCD] Warning: no action defined for step {}".format(step.name))


    def get_managing_target(self, flight_name: str) -> typing.Optional[TestTarget]:
        for role, target in self.targets.items():
            if target.has_created_flight(flight_name):
                return target
        return None

    def get_target(self, step: TestStep) -> typing.Optional[TestTarget]:
        if 'inject_flight' in step:
            return self.targets[step.inject_flight.injection_target.uss_role]
        elif 'delete_flight' in step:
            return self.get_managing_target(step.delete_flight.flight_name)
        else:
            raise NotImplementedError("Unsupported step. A Test Step shall contain either a inject_flight or a delete_flight object.")

    def print_targets_state(self):
        print("[SCD] Targets States:")
        for name, target in self.targets.items():
            print(f"[SCD]   - {name}: {target.created_flight_ids}")

    def teardown(self, dry=False):
        print ("[SCD] Teardown {}".format(self.automated_test.name))
        for role, target in self.targets.items():
            target.delete_all_flights(dry=dry)


def combine_targets(targets: List[InjectionTargetConfiguration], steps: List[TestStep]) -> typing.Iterator[Dict[str, TestTarget]]:
    injection_steps = filter(lambda step: 'inject_flight' in step, steps)
    # Get unique uss roles in injection steps
    uss_roles = sorted(set(map(lambda step: step.inject_flight.injection_target.uss_role, injection_steps)))
    for t in itertools.permutations(targets, len(uss_roles)):
        target_set = {}
        for i, role in enumerate(uss_roles):
            target_set[role] = t[i]
        print(target_set)
        yield target_set


def run_scd_tests(locale: Locality, test_configuration: SCDQualifierTestConfiguration,
                  auth_spec: str, dry=False):
    automated_tests = load_scd_test_definitions(locale)
    configured_targets = list(map(lambda t: TestTarget(t.name, t, auth_spec), test_configuration.injection_targets))

    for test_id, test in automated_tests.items():
        combinations = combine_targets(configured_targets, test.steps)
        for i, targets_under_test in enumerate(combinations):
            print('[SCD] Starting test combination {}: {} ({}/{}) {}'.format(i+1,  test.name, locale, test_id, list(map(lambda t: "{}: {}".format(t[0], t[1].name), targets_under_test.items()))))
            runner = TestRunner(auth_spec, test_id, test, targets_under_test)

            if dry:
                runner.print_test_plan()
            else:
                runner.run_automated_test()
                runner.teardown()



