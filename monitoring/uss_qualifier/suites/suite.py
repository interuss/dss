from datetime import datetime
from enum import Enum
import json
import os
from typing import Dict, List, Optional

from implicitdict import ImplicitDict, StringBasedDateTime
import yaml

from monitoring.monitorlib.inspection import fullname
from monitoring.uss_qualifier.reports import (
    TestScenarioReport,
    FailedCheck,
    TestSuiteReport,
    TestSuiteActionReport,
)
from monitoring.uss_qualifier.resources import Resource, ResourceID, ResourceType
from monitoring.uss_qualifier.scenarios.scenario import (
    TestScenario,
    TestScenarioDeclaration,
)


SuiteType = str
"""This plain string represents a type of test suite, expressed as the file name of the suite definition (without extension) qualified relative to this `suites` folder"""


class TestSuiteDeclaration(ImplicitDict):
    suite_type: SuiteType
    """Type of test suite"""

    resources: Dict[ResourceID, ResourceID]
    """Mapping of the ID a resource will be known by in the child test suite -> the ID a resource is known by in the parent test suite.
    
    The child suite resource <key> is supplied by the parent suite resource <value>.
    """


class ReactionToFailure(str, Enum):
    Continue = "Continue"
    """If the test suite action fails, continue to the next action in that test suite"""

    Abort = "Abort"
    """If the test suite action fails, do not execute any more actions in that test suite"""


class TestSuiteActionDeclaration(ImplicitDict):
    """Defines a step in the sequence of things to do for a test suite.

    Exactly one of `test_scenario` or `test_suite` must be specified.
    """

    test_scenario: Optional[TestScenarioDeclaration]
    """If this field is populated, declaration of the test scenario to run"""

    test_suite: Optional[TestSuiteDeclaration]
    """If this field is populated, declaration of the test suite to run"""

    on_failure: ReactionToFailure
    """What to do if this action fails"""


class TestSuiteDefinition(ImplicitDict):
    """Schema for the definition of a test suite, analogous to the Python TestScenario subclass for scenarios"""

    name: str
    """Name of the test suite"""

    resources: Dict[ResourceID, ResourceType]
    """Enumeration of the resources used by this test suite"""

    actions: List[TestSuiteActionDeclaration]
    """The actions to take when running the test suite.  Components will be executed in order."""

    @staticmethod
    def load(suite_type: str) -> "TestSuiteDefinition":
        path_parts = [os.path.dirname(__file__)]
        path_parts += suite_type.split(".")
        yaml_file = os.path.join(*path_parts) + ".yaml"
        if os.path.exists(yaml_file):
            with open(yaml_file, "r") as f:
                suite_dict = yaml.safe_load(f)
        else:
            json_file = os.path.join(*path_parts) + ".json"
            with open(json_file, "r") as f:
                suite_dict = json.load(f)
        return ImplicitDict.parse(suite_dict, TestSuiteDefinition)


def _print_failed_check(failed_check: FailedCheck) -> None:
    print("New failed check:")
    yaml_lines = yaml.dump(json.loads(json.dumps(failed_check))).split("\n")
    print("\n".join("  " + line for line in yaml_lines))


class TestSuiteAction(object):
    declaration: TestSuiteActionDeclaration
    _test_scenario: Optional[TestScenario] = None
    _test_suite: Optional["TestSuite"] = None
    _resources: Dict[str, Resource]

    def __init__(
        self, action: TestSuiteActionDeclaration, resources: Dict[str, Resource]
    ):
        self.declaration = action
        if "test_scenario" in action and action.test_scenario:
            self._test_scenario = action.test_scenario.make_test_scenario(resources)
        elif "test_suite" in action and action.test_suite:
            resources_for_child = {
                child_resource_id: resources[parent_resource_id]
                for child_resource_id, parent_resource_id in action.test_suite.resources.items()
            }
            self._test_suite = TestSuite(
                definition=TestSuiteDefinition.load(action.test_suite.suite_type),
                resources=resources_for_child,
            )
        else:
            raise ValueError(
                "Every TestSuiteComponent must specify `test_scenario` or `test_suite`"
            )
        self._resources = resources

    def run(self) -> TestSuiteActionReport:
        if self._test_scenario:
            return TestSuiteActionReport(test_scenario=self._run_test_scenario())
        elif self._test_suite:
            return TestSuiteActionReport(test_suite=self._run_test_suite())

    def _run_test_scenario(self) -> TestScenarioReport:
        scenario = self._test_scenario
        print(f'Running "{scenario.documentation.name}" scenario...')
        scenario.on_failed_check = _print_failed_check
        try:
            scenario.run()
        except KeyboardInterrupt:
            raise
        except Exception as e:
            scenario.record_execution_error(e)
        report = scenario.get_report()
        if report.successful:
            print(f'SUCCESS for "{scenario.documentation.name}" scenario')
        else:
            if "execution_error" in report:
                lines = report.execution_error.stacktrace.split("\n")
                print("\n".join("  " + line for line in lines))
            print(f'FAILURE for "{scenario.documentation.name}" scenario')
        return report

    def _run_test_suite(self) -> TestSuiteReport:
        print(f"Beginning test suite {self._test_suite.definition.name}...")
        report = self._test_suite.run()
        print(f"Completed test suite {self._test_suite.definition.name}")
        return report


class TestSuite(object):
    definition: TestSuiteDefinition
    _actions: List[TestSuiteAction]

    def __init__(
        self, definition: TestSuiteDefinition, resources: Dict[ResourceID, Resource]
    ):
        self.definition = definition
        for resource_id, resource_type in definition.resources.items():
            if resource_id not in resources:
                raise ValueError(
                    f'Test suite "{definition.name}" is missing resource {resource_id} ({resource_type})'
                )
            if not resources[resource_id].is_type(resource_type):
                raise ValueError(
                    f'Test suite "{definition.name}" expected resource {resource_id} to be {resource_type}, but instead it was provided {fullname(resources[resource_id].__class__)}'
                )
        self._actions = [TestSuiteAction(a, resources) for a in definition.actions]

    def run(self) -> TestSuiteReport:
        report = TestSuiteReport(
            name=self.definition.name,
            documentation_url="",  # TODO: Populate correctly
            start_time=StringBasedDateTime(datetime.utcnow()),
            actions=[],
        )
        success = True
        for a, action in enumerate(self._actions):
            action_report = action.run()
            report.actions.append(action_report)
            if not action_report.successful():
                success = False
                if action.declaration.on_failure == ReactionToFailure.Abort:
                    break
                elif action.declaration.on_failure == ReactionToFailure.Continue:
                    continue
                else:
                    raise ValueError(
                        f"Action {a} of test suite {self.definition.name} indicate an unrecognized reaction to failure: {str(action.declaration.on_failure)}"
                    )
        report.successful = success
        report.end_time = StringBasedDateTime(datetime.utcnow())
        return report
