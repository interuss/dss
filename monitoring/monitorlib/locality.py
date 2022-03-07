from enum import Enum


class Locality(str, Enum):
    """Operating locations and their respective regulation and technical variations."""
    CHE = 'CHE'
    """Switzerland"""

    @property
    def is_uspace_applicable(self) -> bool:
        return self in {Locality.CHE}

    @property
    def allow_same_priority_intersections(self) -> bool:
        return self in set()
