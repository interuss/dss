from typing import Optional
from urllib.parse import urlparse

from monitoring.monitorlib.typing import ImplicitDict
from monitoring.uss_qualifier.rid.utils import RIDQualifierTestConfiguration
from monitoring.uss_qualifier.scd.configuration import SCDQualifierTestConfiguration


class USSQualifierTestConfiguration(ImplicitDict):
    locale: str
    """A three letter ISO 3166 country code to run the qualifier against.

  This should be the same one used to simulate the flight_data in
  the flight_data_generator.py module."""

    rid: Optional[RIDQualifierTestConfiguration]
    """Test configuration for RID"""

    scd: Optional[SCDQualifierTestConfiguration]
    """Test configuration for SCD"""


def is_url(url_string):
    try:
        urlparse(url_string)
    except ValueError:
        raise ValueError("A valid url must be passed")
