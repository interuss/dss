from typing import List, Optional
from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration


class USSQualifierTestConfiguration(ImplicitDict):
    locale: str
    """A three letter ISO 3166 country code to run the qualifier against.

  This should be the same one used to simulate the flight_data in
  the flight_data_generator.py module."""

    scd: Optional[SCDQualifierTestConfiguration]
    """Test configuration for SCD"""

    config: str = ""
    """Configuration string according to monitoring/uss_qualifier/configurations/README.md"""


def is_url(url_string):
    try:
        urlparse(url_string)
    except ValueError:
        raise ValueError("A valid url must be passed")
