from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.monitorlib import infrastructure
from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapter


class DSSInstanceSpecification(ImplicitDict):
    participant_id: str
    """ID of the USS responsible for this DSS instance"""

    base_url: str
    """Base URL for the DSS instance according to the ASTM F3548-21 API"""

    def __init__(self, *args, **kwargs):
        super().__init__(**kwargs)
        try:
            urlparse(self.base_url)
        except ValueError:
            raise ValueError("DSSInstanceConfiguration.base_url must be a URL")


class DSSInstance(object):
    participant_id: str
    client: infrastructure.UTMClientSession

    def __init__(
        self,
        participant_id: str,
        base_url: str,
        auth_adapter: infrastructure.AuthAdapter,
    ):
        self.participant_id = participant_id
        self._base_url = base_url
        self.client = infrastructure.UTMClientSession(base_url, auth_adapter)


class DSSInstanceResource(Resource[DSSInstanceSpecification]):
    dss: DSSInstance

    def __init__(
        self,
        specification: DSSInstanceSpecification,
        auth_adapter: AuthAdapter,
    ):
        self.dss = DSSInstance(
            specification.participant_id, specification.base_url, auth_adapter.adapter
        )
