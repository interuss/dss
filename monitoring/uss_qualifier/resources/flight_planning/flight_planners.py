from typing import List, Iterable

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapterResource
from monitoring.uss_qualifier.resources.flight_planning.target import (
    FlightPlannerConfiguration,
    TestTarget,
)


class FlightPlannersSpecification(ImplicitDict):
    flight_planners: List[FlightPlannerConfiguration]


class FlightPlannersResource(Resource[FlightPlannersSpecification]):
    flight_planners: List[TestTarget]

    def __init__(
        self,
        specification: FlightPlannersSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self.flight_planners = [
            TestTarget(p, auth_adapter.adapter) for p in specification.flight_planners
        ]

    def make_subset(self, select_indices: Iterable[int]) -> "FlightPlannersResource":
        subset = [self.flight_planners[i] for i in select_indices]
        subset_resource = FlightPlannersResource.__new__(FlightPlannersResource)
        subset_resource.flight_planners = subset
        return subset_resource
