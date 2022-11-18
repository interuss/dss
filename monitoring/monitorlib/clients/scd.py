from typing import List, Optional

from monitoring.monitorlib import scd
from monitoring.monitorlib.infrastructure import UTMClientSession

from implicitdict import ImplicitDict
import datetime
import json
from loguru import logger
import os


def create_subscription(utm_client: UTMClientSession, id: str):
    time_start = datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")
    time_end = (datetime.datetime.utcnow() + datetime.timedelta(minutes=30)).strftime(
        "%Y-%m-%dT%H:%M:%SZ"
    )
    payload = {
        "extents": {
            "volume": {
                "outline_circle": None,
                "outline_polygon": {
                    "vertices": [
                        {"lat": 7.40, "lng": 46.5},
                        {"lat": 7.50, "lng": 46.5},
                        {"lat": 7.50, "lng": 47.5},
                        {"lat": 7.40, "lng": 47.5},
                    ]
                },
                "altitude_lower": {"value": 0.0, "reference": "W84", "units": "M"},
                "altitude_upper": {"value": 1000.0, "reference": "W84", "units": "M"},
            },
            "time_start": {"value": "{}".format(time_start), "format": "RFC3339"},
            "time_end": {"value": "{}".format(time_end), "format": "RFC3339"},
        },
        "uss_base_url": "http://host.docker.internal:10206/interop",
        "notify_for_operational_intents": True,
        "notify_for_constraints": True,
    }
    if os.environ.get("MESSAGE_SIGNING", None) == "true":
        subscription_response = utm_client.put(
            "/dss/v1/subscriptions/{}".format(id), json=payload, scope=scd.SCOPE_SC
        )
        logger.info(
            "Create Subscription response: {}".format(
                str(subscription_response.status_code)
            )
        )
        subscription_response.raise_for_status()


class OperationError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""

    def __init__(self, msg):
        super(OperationError, self).__init__(msg)


# === DSS operations defined in ASTM API ===


def query_operational_intent_references(
    utm_client: UTMClientSession, area_of_interest: scd.Volume4D
) -> List[scd.OperationalIntentReference]:
    req = scd.QueryOperationalIntentReferenceParameters(
        area_of_interest=area_of_interest
    )
    resp = utm_client.post(
        "/dss/v1/operational_intent_references/query", json=req, scope=scd.SCOPE_SC
    )
    if resp.status_code != 200:
        raise OperationError(
            "queryOperationalIntentReferences failed {}:\n{}".format(
                resp.status_code, resp.content.decode("utf-8")
            )
        )
    resp_body = ImplicitDict.parse(
        resp.json(), scd.QueryOperationalIntentReferenceResponse
    )
    return resp_body.operational_intent_references


def create_operational_intent_reference(
    utm_client: UTMClientSession,
    id: str,
    req: scd.PutOperationalIntentReferenceParameters,
) -> scd.ChangeOperationalIntentReferenceResponse:
    resp = utm_client.put(
        "/dss/v1/operational_intent_references/{}".format(id),
        json=req,
        scope=scd.SCOPE_SC,
    )
    if resp.status_code != 200 and resp.status_code != 201:
        raise OperationError(
            "createOperationalIntentReference failed {}:\n{}".format(
                resp.status_code, resp.content.decode("utf-8")
            )
        )
    return ImplicitDict.parse(resp.json(), scd.ChangeOperationalIntentReferenceResponse)


def delete_operational_intent_reference(
    utm_client: UTMClientSession, id: str, ovn: str
) -> scd.ChangeOperationalIntentReferenceResponse:
    resp = utm_client.delete(
        "/dss/v1/operational_intent_references/{}/{}".format(id, ovn),
        scope=scd.SCOPE_SC,
    )
    if resp.status_code != 200:
        raise OperationError(
            "deleteOperationalIntentReference failed {}:\n{}".format(
                resp.status_code, resp.content.decode("utf-8")
            )
        )
    return ImplicitDict.parse(resp.json(), scd.ChangeOperationalIntentReferenceResponse)


# === USS operations defined in the ASTM API ===


def get_operational_intent_details(
    utm_client: UTMClientSession, uss_base_url: str, id: str
) -> scd.OperationalIntent:
    resp = utm_client.get(
        "{}/uss/v1/operational_intents/{}".format(uss_base_url, id), scope=scd.SCOPE_SC
    )
    if resp.status_code != 200:
        logger.error(resp)
        raise OperationError(
            "getOperationalIntentDetails failed {}:\n{}".format(
                resp.status_code, resp.content.decode("utf-8")
            )
        )
    resp_body = ImplicitDict.parse(resp.json(), scd.GetOperationalIntentDetailsResponse)
    return resp_body.operational_intent, resp


def notify_operational_intent_details_changed(
    utm_client: UTMClientSession,
    uss_base_url: str,
    update: scd.PutOperationalIntentDetailsParameters,
) -> None:
    resp = utm_client.post(
        "{}/uss/v1/operational_intents".format(uss_base_url),
        json=update,
        scope=scd.SCOPE_SC,
    )
    if resp.status_code != 204 and resp.status_code != 200:
        logger.error(resp)
        raise OperationError(
            "notifyOperationalIntentDetailsChanged failed {}:\n{}".format(
                resp.status_code, resp.content.decode("utf-8")
            )
        )
    return resp


# === Custom actions ===


def notify_subscribers(
    utm_client: UTMClientSession,
    id: str,
    operational_intent: Optional[scd.OperationalIntent],
    subscribers: List[scd.SubscriberToNotify],
):
    notify_responses = []
    for subscriber in subscribers:
        kwargs = {
            "operational_intent_id": id,
            "subscriptions": subscriber.subscriptions,
        }
        if operational_intent is not None:
            kwargs["operational_intent"] = operational_intent
        update = scd.PutOperationalIntentDetailsParameters(**kwargs)
        notify_responses.append(
            notify_operational_intent_details_changed(
                utm_client, subscriber.uss_base_url, update
            )
        )
    return notify_responses
