from abc import ABC, abstractmethod
import inspect
from typing import Dict, TypeVar

from implicitdict import ImplicitDict

from monitoring.monitorlib import inspection
from monitoring.uss_qualifier import scenarios as scenarios_module
from monitoring.uss_qualifier.scenarios.documentation import (
    TestScenarioDocumentation,
    parse_documentation,
)
from monitoring.uss_qualifier.resources import Resource


class TestScenario(ABC):
    """Instance of a test scenario, ready to run after construction.

    Concrete subclasses of TestScenario must:
      1) Implement a constructor that accepts only parameters with types that are subclasses of Resource
      2) Call TestScenario.__init__ from the subclass's __init__
    """

    documentation: TestScenarioDocumentation

    def __init__(self):
        self.documentation = parse_documentation(self.__class__)

    @abstractmethod
    # TODO: have `run` interact with an encapsulated portion of a report, or return contributions to a report
    def run(self):
        raise NotImplementedError(
            "A concrete test scenario must implement `run` method"
        )


TestScenarioType = TypeVar("TestScenarioType", bound=TestScenario)


class TestScenarioDeclaration(ImplicitDict):
    scenario_type: str
    """Type of test scenario, expressed as a Python class name qualified relative to this `scenarios` module"""

    resources: Dict[str, str] = {}
    """Mapping of resource parameter (additional argument to concrete test scenario constructor) to globally unique name of resource to use"""

    def make_test_scenario(self, resource_pool: Dict[str, Resource]) -> TestScenario:
        inspection.import_submodules(scenarios_module)
        scenario_type = inspection.get_module_object_by_name(
            scenarios_module, self.scenario_type
        )
        if not issubclass(scenario_type, TestScenario):
            raise NotImplementedError(
                "Scenario type {} is not a subclass of the TestScenario base class".format(
                    scenario_type.__name__
                )
            )

        constructor_signature = inspect.signature(scenario_type.__init__)
        constructor_args = {}
        for arg_name, arg in constructor_signature.parameters.items():
            if arg_name == "self":
                continue
            if arg_name not in self.resources:
                raise ValueError(
                    'Test scenario declaration for {} is missing a source for resource "{}"'.format(
                        self.scenario_type, arg
                    )
                )
            if self.resources[arg_name] not in resource_pool:
                raise ValueError(
                    'Resource "{}" was not found in the resource pool when trying to create test scenario "{}" ({})'.format(
                        self.resources[arg_name], self.name, self.type
                    )
                )
            constructor_args[arg_name] = resource_pool[self.resources[arg_name]]

        return scenario_type(**constructor_args)
