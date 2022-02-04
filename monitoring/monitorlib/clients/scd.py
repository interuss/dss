from typing import List, Optional

from monitoring.monitorlib import scd
from monitoring.monitorlib.infrastructure import DSSTestSession
from monitoring.monitorlib.typing import ImplicitDict


class OperationError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""
    def __init__(self, msg):
        super(OperationError, self).__init__(msg)


# === DSS operations defined in ASTM API ===


def query_operational_intent_references(utm_client: DSSTestSession, area_of_interest: scd.Volume4D) -> List[scd.OperationalIntentReference]:
    req = scd.QueryOperationalIntentReferenceParameters(area_of_interest=area_of_interest)
    resp = utm_client.post('/dss/v1/operational_intent_references/query', json=req, scope=scd.SCOPE_SC)
    if resp.status_code != 200:
        raise OperationError('queryOperationalIntentReferences failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    resp_body = ImplicitDict.parse(resp.json(), scd.QueryOperationalIntentReferenceResponse)
    return resp_body.operational_intent_references


def create_operational_intent_reference(utm_client: DSSTestSession, id: str, req: scd.PutOperationalIntentReferenceParameters) -> scd.ChangeOperationalIntentReferenceResponse:
    resp = utm_client.put('/dss/v1/operational_intent_references/{}'.format(id), json=req, scope=scd.SCOPE_SC)
    if resp.status_code != 200 and resp.status_code != 201:
        raise OperationError('createOperationalIntentReference failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    return ImplicitDict.parse(resp.json(), scd.ChangeOperationalIntentReferenceResponse)


def delete_operational_intent_reference(utm_client: DSSTestSession, id: str, ovn: str) -> scd.ChangeOperationalIntentReferenceResponse:
    resp = utm_client.delete('/dss/v1/operational_intent_references/{}/{}'.format(id, ovn), scope=scd.SCOPE_SC)
    if resp.status_code != 200:
        raise OperationError('deleteOperationalIntentReference failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    return ImplicitDict.parse(resp.json(), scd.ChangeOperationalIntentReferenceResponse)


# === USS operations defined in the ASTM API ===


def get_operational_intent_details(utm_client: DSSTestSession, uss_base_url: str, id: str) -> scd.OperationalIntent:
    resp = utm_client.get('{}/uss/v1/operational_intents/{}'.format(uss_base_url, id), scope=scd.SCOPE_SC)
    if resp.status_code != 200:
        raise OperationError('getOperationalIntentDetails failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    resp_body = ImplicitDict.parse(resp.json(), scd.GetOperationalIntentDetailsResponse)
    return resp_body.operational_intent


def notify_operational_intent_details_changed(utm_client: DSSTestSession, uss_base_url: str, update: scd.PutOperationalIntentDetailsParameters) -> None:
    resp = utm_client.post('{}/uss/v1/operational_intents'.format(uss_base_url), json=update, scope=scd.SCOPE_SC)
    if resp.status_code != 204 and resp.status_code != 200:
        raise OperationError('notifyOperationalIntentDetailsChanged failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))


# === Custom actions ===


def notify_subscribers(utm_client: DSSTestSession, id: str, operational_intent: Optional[scd.OperationalIntent], subscribers: List[scd.SubscriberToNotify]):
    for subscriber in subscribers:
        kwargs = {
            'operational_intent_id': id,
            'subscriptions': subscriber.subscriptions,
        }
        if operational_intent is not None:
            kwargs['operational_intent'] = operational_intent
        update = scd.PutOperationalIntentDetailsParameters(**kwargs)
        notify_operational_intent_details_changed(utm_client, subscriber.uss_base_url, update)
