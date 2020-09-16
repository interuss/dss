import datetime
from typing import List, Optional

import s2sphere
import yaml
from yaml.representer import Representer

from monitoring.monitorlib import infrastructure, rid
from monitoring.monitorlib import fetch


class MutatedSubscription(fetch.Query):
  @property
  def success(self) -> bool:
    return not self.errors

  @property
  def errors(self) -> List[str]:
    if self.status_code != 200:
      return ['Failed to {} RID Subscription ({})'.format(self.mutation, self.status_code)]
    if self.json_result is None:
      return ['Response did not contain valid JSON']
    sub = self.subscription
    if sub is None or not sub.valid:
      return ['Response returned an invalid Subscription']

  @property
  def subscription(self) -> Optional[rid.Subscription]:
    if self.json_result is None:
      return None
    sub = self.json_result.get('subscription', None)
    if not sub:
      return None
    return rid.Subscription(sub)

  @property
  def mutation(self) -> str:
    return self['mutation']
yaml.add_representer(MutatedSubscription, Representer.represent_dict)


def put_subscription(utm_client: infrastructure.DSSTestSession,
                     area: s2sphere.LatLngRect,
                     start_time: datetime.datetime,
                     end_time: datetime.datetime,
                     callback_url: str,
                     subscription_id: str,
                     subscription_version: Optional[str]=None) -> MutatedSubscription:
  body = {
    'extents': {
      'spatial_volume': {
        'footprint': {
          'vertices': rid.vertices_from_latlng_rect(area)
        },
        'altitude_lo': 0,
        'altitude_hi': 3048,
      },
      'time_start': start_time.strftime(rid.DATE_FORMAT),
      'time_end': end_time.strftime(rid.DATE_FORMAT),
    },
    'callbacks': {
      'identification_service_area_url': callback_url
    },
  }
  if subscription_version is None:
    url = '/v1/dss/subscriptions/{}'.format(subscription_id)
  else:
    url = '/v1/dss/subscriptions/{}/{}'.format(subscription_id, subscription_version)
  result = MutatedSubscription(fetch.query_and_describe(
    utm_client, 'PUT', url, json=body, scope=rid.SCOPE_READ))
  result['mutation'] = 'create' if subscription_version is None else 'update'
  return result


def delete_subscription(utm_client: infrastructure.DSSTestSession,
                        subscription_id: str,
                        subscription_version: str) -> MutatedSubscription:
  url = '/v1/dss/subscriptions/{}/{}'.format(subscription_id, subscription_version)
  result = MutatedSubscription(fetch.query_and_describe(
    utm_client, 'DELETE', url, scope=rid.SCOPE_READ))
  result['mutation'] = 'delete'
  return result
