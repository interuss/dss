import datetime
from typing import List
from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.monitorlib import fetch, infrastructure
from monitoring.monitorlib.rid_automated_testing.injection_api import (
    CreateTestParameters,
    SCOPE_RID_QUALIFIER_INJECT,
)
from monitoring.uss_qualifier.resources import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapter


class ServiceProviderConfiguration(ImplicitDict):
    participant_id: str
    """ID of the NetRID Service Provider into which test data can be injected"""

    injection_base_url: str
    """Base URL for the Service Provider's implementation of the interfaces/automated-testing/rid/injection.yaml API"""

    def __init__(self, *args, **kwargs):
        super().__init__(**kwargs)
        try:
            urlparse(self.injection_base_url)
        except ValueError:
            raise ValueError(
                "ServiceProviderConfiguration.injection_base_url must be a URL"
            )


class NetRIDServiceProvidersSpecification(ImplicitDict):
    service_providers: List[ServiceProviderConfiguration]


class NetRIDServiceProvider(object):
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

    def submit_test(self, request: CreateTestParameters, test_id: str) -> fetch.Query:
        injection_path = "/tests/{}".format(test_id)

        initiated_at = datetime.datetime.utcnow()
        response = self.client.put(
            url=injection_path, json=request, scope=SCOPE_RID_QUALIFIER_INJECT
        )
        return fetch.describe_query(response, initiated_at)


class NetRIDServiceProviders(Resource[NetRIDServiceProvidersSpecification]):
    service_providers: List[NetRIDServiceProvider]

    def __init__(
        self,
        specification: NetRIDServiceProvidersSpecification,
        auth_adapter: AuthAdapter,
    ):
        self.service_providers = [
            NetRIDServiceProvider(
                s.participant_id, s.injection_base_url, auth_adapter.adapter
            )
            for s in specification.service_providers
        ]
