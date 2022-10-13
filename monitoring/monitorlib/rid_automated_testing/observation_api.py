from typing import List, Optional

from implicitdict import ImplicitDict


# Mirrors of types defined in remote ID automated testing observation API


class Position(ImplicitDict):
    lat: float
    lng: float
    alt: Optional[float]


class Path(ImplicitDict):
    positions: List[Position]


class Cluster(ImplicitDict):
    corners: List[Position]
    area_sqm: float
    number_of_flights: int


class Flight(ImplicitDict):
    id: str
    most_recent_position: Optional[Position]
    recent_paths: Optional[List[Path]]


class GetDetailsResponse(ImplicitDict):
    pass


class GetDisplayDataResponse(ImplicitDict):
    flights: List[Flight] = []
    clusters: List[Cluster] = []
