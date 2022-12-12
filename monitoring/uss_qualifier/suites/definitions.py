from enum import Enum
from typing import Dict, List, Optional, TypeVar

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.fileio import load_dict_with_references, FileReference
from monitoring.uss_qualifier.resources.definitions import ResourceID, ResourceTypeName
from monitoring.uss_qualifier.scenarios.definitions import (
    TestScenarioDeclaration,
)


class TestSuiteDeclaration(ImplicitDict):
    suite_type: FileReference
    """Type/location of test suite.  Usually expressed as the file name of the suite definition (without extension) qualified relative to the `uss_qualifier` folder"""

    resources: Dict[ResourceID, ResourceID]
    """Mapping of the ID a resource will be known by in the child test suite -> the ID a resource is known by in the parent test suite.
    
    The child suite resource <key> is supplied by the parent suite resource <value>.
    """


GeneratorTypeName = str
"""This plain string represents a type of action generator, expressed as a Python class name qualified relative to the `uss_qualifier.action_generators` module"""


ActionGeneratorSpecificationType = TypeVar(
    "ActionGeneratorSpecificationType", bound=ImplicitDict
)


class ActionGeneratorDefinition(ImplicitDict):
    generator_type: GeneratorTypeName
    """Type of action generator"""

    specification: dict = {}
    """Specification of action generator; format is the ActionGeneratorSpecificationType that corresponds to the `generator_type`"""

    resources: Dict[ResourceID, ResourceID]
    """Mapping of the ID a resource will be known by in the child action -> the ID a resource is known by in the parent test suite.
    
    The child action resource ID <key> is supplied by the parent test suite resource ID <value>.
    
    Resources not included in this field will not be available to the child action.
    
    If the parent resource ID is suffixed with ? then the resource will not be required (and will not be populated for the child action when not present in the parent)
    """


class ReactionToFailure(str, Enum):
    Continue = "Continue"
    """If the test suite action fails, continue to the next action in that test suite"""

    Abort = "Abort"
    """If the test suite action fails, do not execute any more actions in that test suite"""


class ActionType(str, Enum):
    TestScenario = "test_scenario"
    TestSuite = "test_suite"
    ActionGenerator = "action_generator"

    @staticmethod
    def raise_invalid_action_declaration():
        raise ValueError(
            f"Exactly one of ({', '.join(a for a in ActionType)}) must be specified in a TestSuiteActionDeclaration"
        )


class TestSuiteActionDeclaration(ImplicitDict):
    """Defines a step in the sequence of things to do for a test suite.

    Exactly one of `test_scenario`, `test_suite`, or `action_generator` must be specified.
    """

    test_scenario: Optional[TestScenarioDeclaration]
    """If this field is populated, declaration of the test scenario to run"""

    test_suite: Optional[TestSuiteDeclaration]
    """If this field is populated, declaration of the test suite to run"""

    action_generator: Optional[ActionGeneratorDefinition]
    """If this field is populated, declaration of a generator that will produce 0 or more test suite actions"""

    on_failure: ReactionToFailure = ReactionToFailure.Continue
    """What to do if this action fails"""

    def get_action_type(self) -> ActionType:
        matches = [v for v in ActionType if v in self and self[v]]
        if len(matches) != 1:
            ActionType.raise_invalid_action_declaration()
        return ActionType(matches[0])

    def get_resource_links(self) -> Dict[ResourceID, ResourceID]:
        action_type = self.get_action_type()
        if action_type == ActionType.TestScenario:
            return self.test_scenario.resources
        elif action_type == ActionType.TestSuite:
            return self.test_suite.resources
        elif action_type == ActionType.ActionGenerator:
            return self.action_generator.resources
        else:
            ActionType.raise_invalid_action_declaration()

    def get_child_type(self) -> str:
        action_type = self.get_action_type()
        if action_type == ActionType.TestScenario:
            return self.test_scenario.scenario_type
        elif action_type == ActionType.TestSuite:
            return self.test_suite.suite_type
        elif action_type == ActionType.ActionGenerator:
            return self.action_generator.generator_type
        else:
            ActionType.raise_invalid_action_declaration()


class TestSuiteDefinition(ImplicitDict):
    """Schema for the definition of a test suite, analogous to the Python TestScenario subclass for scenarios"""

    name: str
    """Name of the test suite"""

    resources: Dict[ResourceID, ResourceTypeName]
    """Enumeration of the resources used by this test suite"""

    actions: List[TestSuiteActionDeclaration]
    """The actions to take when running the test suite.  Components will be executed in order."""

    @staticmethod
    def load(suite_type: FileReference) -> "TestSuiteDefinition":
        return ImplicitDict.parse(
            load_dict_with_references(suite_type), TestSuiteDefinition
        )
