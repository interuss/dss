from enum import Enum


class Locality(str, Enum):
    """The mock USS should behave as if it were operating in this locality."""
    CHE = 'CHE'
    """Switzerland"""

    @property
    def is_uspace_applicable(self) -> bool:
        return self in {Locality.CHE}

    @property
    def allow_same_priority_intersections(self) -> bool:
        return self in set()
