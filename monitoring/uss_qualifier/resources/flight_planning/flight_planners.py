from typing import List, Iterable, Dict, Optional

from implicitdict import ImplicitDict
from monitoring.uss_qualifier.reports.report import ParticipantID
from monitoring.uss_qualifier.resources.definitions import ResourceID

from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapterResource
from monitoring.uss_qualifier.resources.flight_planning.flight_planner import (
    FlightPlannerConfiguration,
    FlightPlanner,
)


class FlightPlannerSpecification(ImplicitDict):
    flight_planner: FlightPlannerConfiguration


class FlightPlannerResource(Resource[FlightPlannerSpecification]):
    flight_planner: FlightPlanner

    def __init__(
        self,
        specification: FlightPlannerSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self.flight_planner = FlightPlanner(
            specification.flight_planner, auth_adapter.adapter
        )


class FlightPlannersSpecification(ImplicitDict):
    flight_planners: List[FlightPlannerConfiguration]


class FlightPlannersResource(Resource[FlightPlannersSpecification]):
    flight_planners: List[FlightPlannerResource]

    def __init__(
        self,
        specification: FlightPlannersSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self._specification = specification
        self._auth_adapter = auth_adapter
        self.flight_planners = [
            FlightPlannerResource(
                FlightPlannerSpecification(flight_planner=p), auth_adapter
            )
            for p in specification.flight_planners
        ]

    def make_subset(self, select_indices: Iterable[int]) -> List[FlightPlannerResource]:
        return [self.flight_planners[i] for i in select_indices]


class FlightPlannerCombinationSelectorSpecification(ImplicitDict):
    must_include: Optional[List[ParticipantID]]
    """The set of flight planners which must be included in every combination"""

    maximum_roles: Optional[Dict[ParticipantID, int]]
    """Maximum number of roles a particular participant may fill in any given combination"""


class FlightPlannerCombinationSelectorResource(
    Resource[FlightPlannerCombinationSelectorSpecification]
):
    _specification: FlightPlannerCombinationSelectorSpecification

    def __init__(self, specification: FlightPlannerCombinationSelectorSpecification):
        self._specification = specification

    def is_valid_combination(
        self, flight_planners: Dict[ResourceID, FlightPlannerResource]
    ):
        participants = [
            p.flight_planner.participant_id for p in flight_planners.values()
        ]

        # Apply must_include criteria
        if "must_include" in self._specification:
            for required_participant in self._specification.must_include:
                if required_participant not in participants:
                    return False

        # Apply maximum_roles criteria
        if "maximum_roles" in self._specification:
            for (
                limited_participant,
                max_count,
            ) in self._specification.maximum_roles.items():
                count = sum(
                    (1 if participant == limited_participant else 0)
                    for participant in participants
                )
                if count > max_count:
                    return False

        return True
