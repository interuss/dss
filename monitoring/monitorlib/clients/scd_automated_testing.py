import uuid
from typing import Tuple

from monitoring.monitorlib.clients.scd import OperationError
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResponse, \
    SCOPE_SCD_QUALIFIER_INJECT, InjectFlightResult, DeleteFlightResponse, DeleteFlightResult
from monitoring.monitorlib.typing import ImplicitDict


def create_flight(utm_client: DSSTestSession, uss_base_url: str, flight_request: InjectFlightRequest) -> Tuple[str, InjectFlightResponse]:
    flight_id = str(uuid.uuid4())
    url = '{}/v1/flights/{}'.format(uss_base_url, flight_id)
    print("[SCD] PUT {}".format(url))
    resp = utm_client.put(url, json=flight_request, scope=SCOPE_SCD_QUALIFIER_INJECT)
    if resp.status_code != 200:
        raise OperationError('createFlight failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    return flight_id, ImplicitDict.parse(resp.json(), InjectFlightResponse)


def delete_flight(utm_client: DSSTestSession, uss_base_url: str, flight_id: str) -> DeleteFlightResponse:
    url = '{}/v1/flights/{}'.format(uss_base_url, flight_id)
    print("[SCD] DELETE {}".format(url))
    resp = utm_client.delete(url, scope=SCOPE_SCD_QUALIFIER_INJECT)
    if resp.status_code != 200:
        raise OperationError('deleteFlight failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    return ImplicitDict.parse(resp.json(), DeleteFlightResponse)


