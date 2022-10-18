from abc import ABC, abstractmethod
from datetime import datetime
import inspect
import json
from typing import Dict, List, Optional, TypeVar, Generic

from implicitdict import StringBasedDateTime, ImplicitDict
import yaml

from monitoring.monitorlib.inspection import (
    fullname,
    get_module_object_by_name,
    import_submodules,
)
from monitoring.uss_qualifier.reports.report import (
    ActionGeneratorReport,
    TestScenarioReport,
    FailedCheck,
    TestSuiteReport,
    TestSuiteActionReport,
)
from monitoring.uss_qualifier.resources import ResourceID
from monitoring.uss_qualifier.resources.resource import ResourceType
from monitoring.uss_qualifier.scenarios.scenario import TestScenario
from monitoring.uss_qualifier.suites.definitions import (
    TestSuiteActionDeclaration,
    TestSuiteDefinition,
    ReactionToFailure,
    ActionGeneratorSpecificationType,
    ActionGeneratorDefinition,
)


def _print_failed_check(failed_check: FailedCheck) -> None:
    print("New failed check:")
    yaml_lines = yaml.dump(json.loads(json.dumps(failed_check))).split("\n")
    print("\n".join("  " + line for line in yaml_lines))


class TestSuiteAction(object):
    declaration: TestSuiteActionDeclaration
    _test_scenario: Optional[TestScenario] = None
    _test_suite: Optional["TestSuite"] = None
    _action_generator: Optional["ActionGeneratorType"] = None
    _resources: Dict[ResourceID, ResourceType]

    def __init__(
        self,
        action: TestSuiteActionDeclaration,
        resources: Dict[ResourceID, ResourceType],
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
        elif "action_generator" in action and action.action_generator:
            resources_for_generator = {
                generator_resource_id: resources[suite_resource_id]
                for generator_resource_id, suite_resource_id in action.action_generator.resources.items()
            }
            self._action_generator = ActionGenerator.make_from_definition(
                action.action_generator, resources_for_generator
            )
        else:
            raise ValueError(
                "Every TestSuiteComponent must specify `test_scenario`, `test_suite`, or `action_generator`"
            )
        self._resources = resources

    def run(self) -> TestSuiteActionReport:
        if self._test_scenario:
            return TestSuiteActionReport(test_scenario=self._run_test_scenario())
        elif self._test_suite:
            return TestSuiteActionReport(test_suite=self._run_test_suite())
        elif self._action_generator:
            return TestSuiteActionReport(action_generator=self._run_action_generator())

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

    def _run_action_generator(self) -> ActionGeneratorReport:
        report = ActionGeneratorReport(actions=[])
        while True:
            action_report = self._action_generator.run_next_action()
            if action_report is None:
                break
            report.actions.append(action_report)
        return report


class TestSuite(object):
    definition: TestSuiteDefinition
    _actions: List[TestSuiteAction]

    def __init__(
        self, definition: TestSuiteDefinition, resources: Dict[ResourceID, ResourceType]
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


class ActionGenerator(ABC, Generic[ActionGeneratorSpecificationType]):
    @abstractmethod
    def __init__(
        self,
        specification: ActionGeneratorSpecificationType,
        resources: Dict[ResourceID, ResourceType],
    ):
        """Create an instance of the action generator.

        Concrete subclasses of ActionGenerator must implement their constructor according to this specification.

        :param specification: A serializable (subclass of implicitdict.ImplicitDict) specification for how to create the action generator.  This parameter may be omitted if not needed.
        :param resources: All of the resources available in the test suite in which the action generator is run.
        """
        raise NotImplementedError(
            "A concrete action generator type must implement __init__ method"
        )

    @abstractmethod
    def run_next_action(self) -> Optional[TestSuiteActionReport]:
        """Run the next action from the generator, or else return None if there are no more actions"""
        raise NotImplementedError(
            "A concrete action generator must implement `actions` method"
        )

    @staticmethod
    def make_from_definition(
        definition: ActionGeneratorDefinition, resources: Dict[ResourceID, ResourceType]
    ) -> "ActionGeneratorType":
        from monitoring.uss_qualifier import (
            action_generators as action_generators_module,
        )

        import_submodules(action_generators_module)
        action_generator_type = get_module_object_by_name(
            action_generators_module, definition.generator_type
        )
        if not issubclass(action_generator_type, ActionGenerator):
            raise NotImplementedError(
                "Action generator type {} is not a subclass of the ActionGenerator base class".format(
                    action_generator_type.__name__
                )
            )
        constructor_signature = inspect.signature(action_generator_type.__init__)
        specification_type = None
        constructor_args = {}
        for arg_name, arg in constructor_signature.parameters.items():
            if arg_name == "specification":
                specification_type = arg.annotation
                break
        if specification_type is not None:
            constructor_args["specification"] = ImplicitDict.parse(
                definition.specification, specification_type
            )
        constructor_args["resources"] = resources
        return action_generator_type(**constructor_args)


ActionGeneratorType = TypeVar("ActionGeneratorType", bound=ActionGenerator)
