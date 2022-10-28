import arrow

from implicitdict import ImplicitDict
from monitoring.monitorlib import fetch
from monitoring.monitorlib.infrastructure import AuthAdapter, UTMClientSession
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    SCOPE_SCD_QUALIFIER_INJECT,
)
from monitoring.uss_qualifier.reports.report import ParticipantID
from monitoring.uss_qualifier.resources.communications import AuthAdapterResource
from monitoring.uss_qualifier.resources.resource import Resource


class MockUSSClient(object):
    """Means to communicate with an InterUSS mock_uss instance"""

    def __init__(
        self,
        participant_id: str,
        base_url: str,
        auth_adapter: AuthAdapter,
    ):
        self.session = UTMClientSession(base_url, auth_adapter)
        self.participant_id = participant_id

    def get_status(self) -> fetch.Query:
        initiated_at = arrow.utcnow().datetime
        resp = self.session.get("/scdsc/v1/status", scope=SCOPE_SCD_QUALIFIER_INJECT)
        return fetch.describe_query(resp, initiated_at)

    # TODO: Add other methods to interact with the mock USS in other ways (like starting/stopping message signing data collection)


class MockUSSSpecification(ImplicitDict):
    mock_uss_base_url: str
    """The base URL for the mock USS.
    
    If the mock USS had scdsc enabled, for instance, then these URLs would be
    valid:
      * <mock_uss_base_url>/mock/scd/uss/v1/reports
      * <mock_uss_base_url>/scdsc/v1/status
    """

    participant_id: ParticipantID
    """Test participant responsible for this mock USS."""


class MockUSSResource(Resource[MockUSSSpecification]):
    mock_uss: MockUSSClient

    def __init__(
        self,
        specification: MockUSSSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self.mock_uss = MockUSSClient(
            specification.participant_id,
            specification.mock_uss_base_url,
            auth_adapter.adapter,
        )
