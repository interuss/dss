import inspect
from typing import List

from monitoring.monitorlib.inspection import fullname
from monitoring.uss_qualifier.requirements.documentation import get_requirement
from monitoring.uss_qualifier.scenarios import documentation
from monitoring.uss_qualifier.scenarios.scenario import TestScenarioType


def validate(test_scenarios: List[TestScenarioType]):
    for test_scenario in test_scenarios:
        # Verify that documentation parses
        docs = documentation.get_documentation(test_scenario)

        # Verify that all resources are documented
        constructor_signature = inspect.signature(test_scenario.__init__)
        args = []
        for arg_name, arg in constructor_signature.parameters.items():
            if arg_name == "self":
                continue
            if "resources" not in docs:
                raise ValueError(
                    f'Test scenario {fullname(test_scenario)} declares resources in its constructor, but there is no "{documentation.RESOURCES_HEADING}" section in its documentation'
                )
            if arg_name not in docs.resources:
                raise ValueError(
                    f"Test scenario {fullname(test_scenario)} declares resource {arg_name} ({fullname(arg.annotation)}), but this resource is not documented"
                )
            args.append(arg_name)
        for documented_resource in docs.resources:
            if documented_resource not in args:
                raise ValueError(
                    f"Documentation for test scenario {fullname(test_scenario)} specifies a resource named {documented_resource}, but this resource is not declared as a resource in the constructor"
                )

        # Verify that all requirements are documented
        for case in docs.cases:
            for step in case.steps:
                for check in step.checks:
                    for req in check.applicable_requirements:
                        try:
                            get_requirement(req)
                        except ValueError as e:
                            raise ValueError(
                                f'Error verifying documentation for requirement "{req}" in check "{check.name}" of step "{step.name}" for case "{case.name}" in scenario "{fullname(test_scenario)}": {str(e)}'
                            )
        if "cleanup" in docs:
            for check in docs.cleanup.checks:
                for req in check.applicable_requirements:
                    try:
                        get_requirement(req)
                    except ValueError as e:
                        raise ValueError(
                            f'Error verifying documentation for requirement "{req}" in check "{check.name}" of cleanup phase in scenario "{fullname(test_scenario)}": {str(e)}'
                        )
