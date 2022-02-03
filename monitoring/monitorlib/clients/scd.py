from typing import Dict, List

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


# === USS operations defined in the ASTM API ===


def get_operational_intent_details(utm_client: DSSTestSession, uss_base_url: str, id: str) -> scd.OperationalIntent:
    resp = utm_client.get('{}/uss/v1/operational_intents/{}'.format(uss_base_url, id), scope=scd.SCOPE_SC)
    if resp.status_code != 200:
        raise OperationError('getOperationalIntentDetails failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))
    resp_body = ImplicitDict.parse(resp.json(), scd.GetOperationalIntentDetailsResponse)
    return resp_body.operational_intent


def notify_operational_intent_details_changed(utm_client: DSSTestSession, uss_base_url: str, update: scd.PutOperationalIntentDetailsParameters) -> None:
    resp = utm_client.post('{}/uss/v1/operational_intents'.format(uss_base_url), json=update, scope=scd.SCOPE_SC)
    if resp.status_code != 200:
        raise OperationError('notifyOperationalIntentDetailsChanged failed {}:\n{}'.format(resp.status_code, resp.content.decode('utf-8')))


# === Custom actions ===


def query_operational_intents(utm_client: DSSTestSession, area_of_interest: scd.Volume4D, cache: Dict[str, scd.OperationalIntent]=None) -> List[scd.OperationalIntent]:
    """Retrieve a complete set of operational intents in an area, including details.

    :param utm_client: Means by which to execute HTTP calls to the DSS and other USSs
    :param area_of_interest: Area where intersecting operational intents must be discovered
    :param cache: If specified, this cache is mutated to store the details of operational intents so that the details don't necessarily need to be retrieved next time
    :return: Full definition for every operational intent discovered
    """
    if cache is None:
        cache = dict()
    op_intent_refs = query_operational_intent_references(utm_client, area_of_interest)
    for op_intent_ref in op_intent_refs:
        if op_intent_ref.id not in cache or cache[op_intent_ref.id].reference.version != op_intent_ref.version:
            op_intent = get_operational_intent_details(utm_client, op_intent_ref.uss_base_url, op_intent_ref.id)
            cache[op_intent.reference.id] = op_intent
    return [cache[op_intent_ref.id] for op_intent_ref in op_intent_refs]


def notify_subscribers(utm_client: DSSTestSession, id: str, operational_intent: scd.OperationalIntent, subscribers: List[scd.SubscriberToNotify]):
    for subscriber in subscribers:
        update = scd.PutOperationalIntentDetailsParameters(
            operational_intent_id=id,
            operational_intent=operational_intent,
            subscriptions=subscriber.subscriptions)
        notify_operational_intent_details_changed(utm_client, subscriber.uss_base_url, update)
