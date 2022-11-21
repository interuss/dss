from datetime import datetime
from typing import Tuple, List
from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.monitorlib import infrastructure, fetch
from monitoring.monitorlib.scd import (
    Volume4D,
    QueryOperationalIntentReferenceParameters,
    SCOPE_SC,
    QueryOperationalIntentReferenceResponse,
    OperationalIntentReference,
    OperationalIntent,
    GetOperationalIntentDetailsResponse,
)
from monitoring.uss_qualifier.resources.resource import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapterResource


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

    def find_op_intent(
        self, extent: Volume4D
    ) -> Tuple[List[OperationalIntentReference], fetch.Query]:
        url = "/dss/v1/operational_intent_references/query"
        req = QueryOperationalIntentReferenceParameters(area_of_interest=extent)

        initiated_at = datetime.utcnow()
        resp = self.client.post(url, scope=SCOPE_SC, json=req)
        if resp.status_code != 200:
            result = None
        else:
            result = ImplicitDict.parse(
                resp.json(), QueryOperationalIntentReferenceResponse
            ).operational_intent_references
        return result, fetch.describe_query(resp, initiated_at)

    def get_full_op_intent(
        self, op_intent_ref: OperationalIntentReference
    ) -> Tuple[OperationalIntent, fetch.Query]:
        url = f"{op_intent_ref.uss_base_url}/uss/v1/operational_intents/{op_intent_ref.id}"

        initiated_at = datetime.utcnow()
        resp = self.client.get(url, scope=SCOPE_SC)
        if resp.status_code != 200:
            result = None
        else:
            result = ImplicitDict.parse(
                resp.json(), GetOperationalIntentDetailsResponse
            ).operational_intent
        return result, fetch.describe_query(resp, initiated_at)


class DSSInstanceResource(Resource[DSSInstanceSpecification]):
    dss: DSSInstance

    def __init__(
        self,
        specification: DSSInstanceSpecification,
        auth_adapter: AuthAdapterResource,
    ):
        self.dss = DSSInstance(
            specification.participant_id, specification.base_url, auth_adapter.adapter
        )
