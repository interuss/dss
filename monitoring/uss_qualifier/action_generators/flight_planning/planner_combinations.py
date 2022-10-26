from typing import Dict, List, Optional

from implicitdict import ImplicitDict

from monitoring.monitorlib.inspection import fullname
from monitoring.uss_qualifier.reports.report import TestSuiteActionReport
from monitoring.uss_qualifier.resources.definitions import ResourceID
from monitoring.uss_qualifier.resources.flight_planning import FlightPlannersResource
from monitoring.uss_qualifier.resources.resource import ResourceType

from monitoring.uss_qualifier.suites.definitions import TestSuiteActionDeclaration
from monitoring.uss_qualifier.suites.suite import (
    ActionGenerator,
    TestSuiteAction,
    ReactionToFailure,
)


class FlightPlannerCombinationsSpecification(ImplicitDict):
    action_to_repeat: TestSuiteActionDeclaration
    """Test suite action to run for each combination of flight planners"""

    flight_planners_source: ResourceID
    """ID of the resource providing all available flight planners"""

    roles: int
    """Number of flight planners to make available to the action, via whichever resource ID is mapped to the parent `flight_planners_source`"""

    resources: Dict[ResourceID, ResourceID]
    """Mapping of the ID a resource will be known by in the child action -> the ID a resource is known by in the parent action generator.
    
    The child action resource <key> is supplied by the parent action generator <value>.
    
    Resources not included in this field or in `roles` will not be available to the child action.
    """


class FlightPlannerCombinations(
    ActionGenerator[FlightPlannerCombinationsSpecification]
):
    _actions: List[TestSuiteAction]
    _current_action: int
    _failure_reaction: ReactionToFailure

    def __init__(
        self,
        specification: FlightPlannerCombinationsSpecification,
        resources: Dict[ResourceID, ResourceType],
    ):
        if specification.flight_planners_source not in resources:
            raise ValueError(
                f"Resource ID {specification.flight_planners_source} was not present in the available resource pool"
            )
        flight_planners_resource: FlightPlannersResource = resources[
            specification.flight_planners_source
        ]
        if not isinstance(flight_planners_resource, FlightPlannersResource):
            raise ValueError(
                f"Expected resource ID {specification.flight_planners_source} to be a {fullname(FlightPlannersResource)} but it was a {fullname(flight_planners_resource.__class__)} instead"
            )
        flight_planners = flight_planners_resource.flight_planners

        self._actions = []
        role_assignments = [0] * specification.roles
        while True:
            modified_parent_resources = {k: v for k, v in resources.items()}
            modified_parent_resources[
                specification.flight_planners_source
            ] = flight_planners_resource.make_subset(role_assignments)
            resources_for_child = {
                child_resource_id: modified_parent_resources[parent_resource_id]
                for child_resource_id, parent_resource_id in specification.resources.items()
            }
            self._actions.append(
                TestSuiteAction(specification.action_to_repeat, resources_for_child)
            )

            index_to_increment = len(role_assignments) - 1
            while index_to_increment >= 0:
                role_assignments[index_to_increment] += 1
                if role_assignments[index_to_increment] >= len(flight_planners):
                    role_assignments[index_to_increment] = 0
                    index_to_increment -= 1
                else:
                    break
            if index_to_increment < 0:
                break

        self._current_action = 0
        self._failure_reaction = specification.action_to_repeat.on_failure

    def run_next_action(self) -> Optional[TestSuiteActionReport]:
        if self._current_action < len(self._actions):
            report = self._actions[self._current_action].run()
            self._current_action += 1
            if not report.successful():
                if self._failure_reaction == ReactionToFailure.Abort:
                    self._current_action = len(self._actions)
            return report
        else:
            return None
