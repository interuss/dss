import datetime
from typing import List, Optional

import s2sphere
import yaml
from yaml.representer import Representer

from monitoring.monitorlib import infrastructure, scd
from monitoring.monitorlib import fetch


class MutatedSubscription(fetch.Query):
    @property
    def success(self) -> bool:
        return not self.errors

    @property
    def errors(self) -> List[str]:
        if self.status_code != 200:
            return [
                "Failed to {} SCD Subscription ({})".format(
                    self.mutation, self.status_code
                )
            ]
        if self.json_result is None:
            return ["Response did not contain valid JSON"]
        sub = self.subscription
        if sub is None or not sub.valid:
            return ["Response returned an invalid Subscription"]

    @property
    def subscription(self) -> Optional[scd.Subscription]:
        if self.json_result is None:
            return None
        sub = self.json_result.get("subscription", None)
        if not sub:
            return None
        return scd.Subscription(sub)

    @property
    def mutation(self) -> str:
        return self["mutation"]


yaml.add_representer(MutatedSubscription, Representer.represent_dict)


def put_subscription(
    utm_client: infrastructure.UTMClientSession,
    area: s2sphere.LatLngRect,
    start_time: datetime.datetime,
    end_time: datetime.datetime,
    base_url: str,
    subscription_id: str,
    min_alt_m: float = 0,
    max_alt_m: float = 3048,
    old_version: int = 0,
) -> MutatedSubscription:
    body = {
        "extents": scd.make_vol4(
            start_time,
            end_time,
            min_alt_m,
            max_alt_m,
            polygon=scd.make_polygon(latlngrect=area),
        ),
        "old_version": old_version,
        "uss_base_url": base_url,
        "notify_for_operations": True,
        "notify_for_constraints": True,
    }
    url = "/dss/v1/subscriptions/{}".format(subscription_id)
    result = MutatedSubscription(
        fetch.query_and_describe(utm_client, "PUT", url, json=body, scope=scd.SCOPE_SC)
    )
    result["mutation"] = "create" if old_version == 0 else "update"
    return result


def delete_subscription(
    utm_client: infrastructure.UTMClientSession, subscription_id: str
) -> MutatedSubscription:
    url = "/dss/v1/subscriptions/{}".format(subscription_id)
    result = MutatedSubscription(
        fetch.query_and_describe(utm_client, "DELETE", url, scope=scd.SCOPE_SC)
    )
    result["mutation"] = "delete"
    return result
