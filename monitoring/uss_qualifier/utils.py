from typing import Optional

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.resources.definitions import ResourceCollection


class USSQualifierTestConfiguration(ImplicitDict):
    # TODO: Remove when SCD uses a test suite
    resources: Optional[ResourceCollection]

    config: str = ""
    """Configuration string according to monitoring/uss_qualifier/configurations/README.md"""
