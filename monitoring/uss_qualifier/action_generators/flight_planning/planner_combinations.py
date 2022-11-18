from typing import Dict, List, Optional

from implicitdict import ImplicitDict

from monitoring.monitorlib.inspection import fullname
from monitoring.uss_qualifier.reports.report import TestSuiteActionReport
from monitoring.uss_qualifier.resources.definitions import ResourceID
from monitoring.uss_qualifier.resources.flight_planning import FlightPlannersResource
from monitoring.uss_qualifier.resources.flight_planning.flight_planners import (
    FlightPlannerCombinationSelectorResource,
)
from monitoring.uss_qualifier.resources.resource import (
    ResourceType,
    make_child_resources,
)

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

    combination_selector_source: Optional[ResourceID] = None
    """If specified and contained in the provided resources, the resource containing a FlightPlannerCombinationSelectorResource to select only a subset of combinations"""

    roles: int
    """Number of flight planners to make available to the action, via whichever resource ID is mapped to the parent `flight_planners_source`"""


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
                f"Resource ID {specification.flight_planners_source} specified as `flight_planners_source` was not present in the available resource pool"
            )
        flight_planners_resource: FlightPlannersResource = resources[
            specification.flight_planners_source
        ]
        if not isinstance(flight_planners_resource, FlightPlannersResource):
            raise ValueError(
                f"Expected resource ID {specification.flight_planners_source} to be a {fullname(FlightPlannersResource)} but it was a {fullname(flight_planners_resource.__class__)} instead"
            )
        flight_planners = flight_planners_resource.flight_planners

        if (
            specification.combination_selector_source is not None
            and specification.combination_selector_source in resources
        ):
            combination_selector = resources[specification.combination_selector_source]
            if not isinstance(
                combination_selector, FlightPlannerCombinationSelectorResource
            ):
                raise ValueError(
                    f"Expected resource ID {specification.combination_selector_source} to be a {fullname(FlightPlannerCombinationSelectorResource)} but it was a {fullname(combination_selector.__class__)} instead"
                )
        else:
            combination_selector = None

        self._actions = []
        role_assignments = [0] * specification.roles
        while True:
            flight_planners_combination = flight_planners_resource.make_subset(
                role_assignments
            )

            if (
                combination_selector is None
                or combination_selector.is_valid_combination(
                    flight_planners_combination
                )
            ):
                modified_resources = {k: v for k, v in resources.items()}
                modified_resources[
                    specification.flight_planners_source
                ] = flight_planners_combination

                self._actions.append(
                    TestSuiteAction(specification.action_to_repeat, modified_resources)
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
