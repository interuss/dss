import uuid
from typing import Tuple

from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import InjectFlightRequest, InjectFlightResponse, \
    SCOPE_SCD_QUALIFIER_INJECT, InjectFlightResult, DeleteFlightResponse, DeleteFlightResult


def create_flight(utm_client: DSSTestSession, uss_base_url: str, flight_request: InjectFlightRequest, dry: bool=False) -> Tuple[str, InjectFlightResponse]:
    flight_id = str(uuid.uuid4())
    url = '{}/v1/flights/{}'.format(uss_base_url, flight_id)
    print("[SCD] PUT {}".format(url))
    if dry:
        return flight_id, InjectFlightResponse(
            result=InjectFlightResult.DryRun,
            notes=[],
            operational_intent_id=str(uuid.uuid4())
        )


def delete_flight(utm_client: DSSTestSession, uss_base_url: str, flight_id: str, dry: bool) -> DeleteFlightResponse:
    url = '{}/v1/flights/{}'.format(uss_base_url, flight_id)
    print("[SCD] DEL {}".format(url))
    if dry:
        return DeleteFlightResponse(
            result=DeleteFlightResult.DryRun,
            notes=[]
        )


