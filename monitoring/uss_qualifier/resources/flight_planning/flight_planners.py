from typing import List

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.resources import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapter
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
        auth_adapter: AuthAdapter,
    ):
        self.flight_planners = [
            TestTarget(p, auth_adapter.adapter) for p in specification.flight_planners
        ]
