from typing import List, Iterable, Dict

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.reports.report import ParticipantID

from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapterResource
from monitoring.uss_qualifier.resources.flight_planning.flight_planner import (
    FlightPlannerConfiguration,
    FlightPlanner,
)


class FlightPlannersSpecification(ImplicitDict):
    flight_planners: List[FlightPlannerConfiguration]


class FlightPlannersResource(Resource[FlightPlannersSpecification]):
    flight_planners: List[FlightPlanner]

    def __init__(
        self,
        specification: FlightPlannersSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self.flight_planners = [
            FlightPlanner(p, auth_adapter.adapter)
            for p in specification.flight_planners
        ]

    def make_subset(self, select_indices: Iterable[int]) -> "FlightPlannersResource":
        subset = [self.flight_planners[i] for i in select_indices]
        subset_resource = FlightPlannersResource.__new__(FlightPlannersResource)
        subset_resource.flight_planners = subset
        return subset_resource


class FlightPlannerCombinationSelectorSpecification(ImplicitDict):
    must_include: List[ParticipantID]
    """The set of flight planners which must be included in every combination"""

    maximum_roles: Dict[ParticipantID, int]
    """Maximum number of roles a particular participant may fill in any given combination"""


class FlightPlannerCombinationSelectorResource(
    Resource[FlightPlannerCombinationSelectorSpecification]
):
    _specification: FlightPlannerCombinationSelectorSpecification

    def __init__(self, specification: FlightPlannerCombinationSelectorSpecification):
        self._specification = specification

    def is_valid_combination(self, flight_planners: FlightPlannersResource):
        participants = [p.participant_id for p in flight_planners.flight_planners]

        accept_combination = True

        # Apply must_include criteria
        for required_participant in self._specification.must_include:
            if required_participant not in participants:
                accept_combination = False
                break

        # Apply maximum_roles criteria
        for limited_participant, max_count in self._specification.maximum_roles.items():
            count = sum(
                (1 if participant == limited_participant else 0)
                for participant in participants
            )
            if count > max_count:
                accept_combination = False
                break

        return accept_combination
