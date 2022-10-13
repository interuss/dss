import datetime
from typing import List
from urllib.parse import urlparse

from implicitdict import ImplicitDict

from monitoring.monitorlib import fetch, infrastructure
from monitoring.monitorlib.rid_automated_testing.injection_api import (
    TestFlight,
    CreateTestParameters,
    SCOPE_RID_QUALIFIER_INJECT,
    ChangeTestResponse,
)
from monitoring.uss_qualifier.resources import Resource
from monitoring.uss_qualifier.resources.communications import AuthAdapter


class ServiceProviderConfiguration(ImplicitDict):
    name: str
    """Name of the NetRID Service Provider into which test data can be injected"""

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
    name: str
    client: infrastructure.UTMClientSession

    def __init__(
        self, name: str, base_url: str, auth_adapter: infrastructure.AuthAdapter
    ):
        self.name = name
        self._base_url = base_url
        self.client = infrastructure.UTMClientSession(base_url, auth_adapter)

    @property
    def config(self) -> ServiceProviderConfiguration:
        return ServiceProviderConfiguration(
            name=self.name, injection_base_url=self._base_url
        )

    def submit_test(
        self, payload: CreateTestParameters, test_id: str, setup
    ) -> List[TestFlight]:
        # Note: this method imported from uss_qualifier/rid/aircraft_state_replayer.py::TestHarness
        # TODO: clean up according to new architecture and encapsulation models

        injection_path = "/tests/{}".format(test_id)

        initiated_at = datetime.datetime.utcnow()
        response = self.client.put(
            url=injection_path, json=payload, scope=SCOPE_RID_QUALIFIER_INJECT
        )
        setup.injections.append(fetch.describe_query(response, initiated_at))

        if response.status_code == 200:
            changed_test: ChangeTestResponse = ImplicitDict.parse(
                response.json(), ChangeTestResponse
            )
            print("New test with ID %s created" % test_id)
            return changed_test.injected_flights
        else:
            raise RuntimeError(
                "Error {} submitting test ID {} to {}: {}".format(
                    response.status_code,
                    test_id,
                    self._base_url,
                    response.content.decode("utf-8"),
                )
            )


class NetRIDServiceProviders(Resource[NetRIDServiceProvidersSpecification]):
    service_providers: List[NetRIDServiceProvider]

    def __init__(
        self,
        specification: NetRIDServiceProvidersSpecification,
        auth_adapter: AuthAdapter,
    ):
        self.service_providers = [
            NetRIDServiceProvider(s.name, s.injection_base_url, auth_adapter.adapter)
            for s in specification.service_providers
        ]
