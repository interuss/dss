from abc import ABC, abstractmethod
from datetime import datetime
import inspect
import json
from typing import Dict, List, Optional, TypeVar, Generic

from implicitdict import StringBasedDateTime, ImplicitDict
import yaml

from monitoring import uss_qualifier as uss_qualifier_module
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
from monitoring.uss_qualifier.resources.definitions import ResourceID
from monitoring.uss_qualifier.resources.resource import (
    ResourceType,
    make_child_resources,
)
from monitoring.uss_qualifier.scenarios.scenario import (
    TestScenario,
    ScenarioCannotContinueError,
    TestRunCannotContinueError,
)
from monitoring.uss_qualifier.suites.definitions import (
    TestSuiteActionDeclaration,
    TestSuiteDefinition,
    ReactionToFailure,
    ActionGeneratorSpecificationType,
    ActionGeneratorDefinition,
    ActionType,
    TestSuiteDeclaration,
)


def _print_failed_check(failed_check: FailedCheck) -> None:
    print("New failed check:")
    yaml_lines = yaml.dump(json.loads(json.dumps(failed_check))).split("\n")
    print("\n".join("  " + line for line in yaml_lines))


class TestSuiteAction(object):
    declaration: TestSuiteActionDeclaration
    test_scenario: Optional[TestScenario] = None
    test_suite: Optional["TestSuite"] = None
    action_generator: Optional["ActionGeneratorType"] = None

    def __init__(
        self,
        action: TestSuiteActionDeclaration,
        resources: Dict[ResourceID, ResourceType],
    ):
        self.declaration = action
        resources_for_child = make_child_resources(
            resources,
            action.get_resource_links(),
            f"Test suite action to run {action.get_action_type()} {action.get_child_type()}",
        )

        action_type = action.get_action_type()
        if action_type == ActionType.TestScenario:
            self.test_scenario = TestScenario.make_test_scenario(
                declaration=action.test_scenario, resource_pool=resources_for_child
            )
        elif action_type == ActionType.TestSuite:
            self.test_suite = TestSuite(
                declaration=action.test_suite,
                resources=resources,
            )
        elif action_type == ActionType.ActionGenerator:
            self.action_generator = ActionGenerator.make_from_definition(
                definition=action.action_generator, resources=resources_for_child
            )
        else:
            ActionType.raise_invalid_action_declaration()

    def run(self) -> TestSuiteActionReport:
        if self.test_scenario:
            return TestSuiteActionReport(test_scenario=self._run_test_scenario())
        elif self.test_suite:
            return TestSuiteActionReport(test_suite=self._run_test_suite())
        elif self.action_generator:
            return TestSuiteActionReport(action_generator=self._run_action_generator())

    def _run_test_scenario(self) -> TestScenarioReport:
        scenario = self.test_scenario
        print(f'Running "{scenario.documentation.name}" scenario...')
        scenario.on_failed_check = _print_failed_check
        try:
            try:
                scenario.run()
            except (ScenarioCannotContinueError, TestRunCannotContinueError):
                pass
            scenario.go_to_cleanup()
            scenario.cleanup()
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
        print(f"Beginning test suite {self.test_suite.definition.name}...")
        report = self.test_suite.run()
        print(f"Completed test suite {self.test_suite.definition.name}")
        return report

    def _run_action_generator(self) -> ActionGeneratorReport:
        report = ActionGeneratorReport(
            actions=[], generator_type=self.action_generator.definition.generator_type
        )
        while True:
            action_report = self.action_generator.run_next_action()
            if action_report is None:
                break
            report.actions.append(action_report)
            if action_report.has_critical_problem():
                break
        return report


class TestSuite(object):
    declaration: TestSuiteDeclaration
    definition: TestSuiteDefinition
    actions: List[TestSuiteAction]

    def __init__(
        self,
        declaration: TestSuiteDeclaration,
        resources: Dict[ResourceID, ResourceType],
    ):
        self.declaration = declaration
        self.definition = TestSuiteDefinition.load(declaration.suite_type)
        local_resources = {
            local_resource_id: resources[parent_resource_id]
            for local_resource_id, parent_resource_id in declaration.resources.items()
        }
        for resource_id, resource_type in self.definition.resources.items():
            is_optional = resource_type.endswith("?")
            if is_optional:
                resource_type = resource_type[:-1]
            if not is_optional and resource_id not in local_resources:
                raise ValueError(
                    f'Test suite "{self.definition.name}" is missing resource {resource_id} ({resource_type})'
                )
            if resource_id in local_resources and not local_resources[
                resource_id
            ].is_type(resource_type):
                raise ValueError(
                    f'Test suite "{self.definition.name}" expected resource {resource_id} to be {resource_type}, but instead it was provided {fullname(resources[resource_id].__class__)}'
                )
        self.actions = [
            TestSuiteAction(action=a, resources=local_resources)
            for a in self.definition.actions
        ]

    def run(self) -> TestSuiteReport:
        report = TestSuiteReport(
            name=self.definition.name,
            suite_type=self.declaration.suite_type,
            documentation_url="",  # TODO: Populate correctly
            start_time=StringBasedDateTime(datetime.utcnow()),
            actions=[],
        )
        success = True
        for a, action in enumerate(self.actions):
            action_report = action.run()
            report.actions.append(action_report)
            if action_report.has_critical_problem():
                success = False
                break
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
    definition: ActionGeneratorDefinition

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
            parent_module=uss_qualifier_module,
            object_name=definition.generator_type,
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
        generator = action_generator_type(**constructor_args)
        generator.definition = definition
        return generator


ActionGeneratorType = TypeVar("ActionGeneratorType", bound=ActionGenerator)
