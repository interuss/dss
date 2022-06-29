import datetime
from typing import Dict, List, Optional

import s2sphere
import yaml
from yaml.representer import Representer

from monitoring.monitorlib import fetch, infrastructure, rid
from monitoring.monitorlib.typing import ImplicitDict


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


def put_subscription(utm_client: infrastructure.UTMClientSession,
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


def delete_subscription(utm_client: infrastructure.UTMClientSession,
                        subscription_id: str,
                        subscription_version: str) -> MutatedSubscription:
  url = '/v1/dss/subscriptions/{}/{}'.format(subscription_id, subscription_version)
  result = MutatedSubscription(fetch.query_and_describe(
    utm_client, 'DELETE', url, scope=rid.SCOPE_READ))
  result['mutation'] = 'delete'
  return result


class MutatedISAResponse(fetch.Query):
  """Response to a call to the DSS to mutate an ISA"""
  @property
  def success(self) -> bool:
    return not self.errors

  @property
  def errors(self) -> List[str]:
    if self.status_code != 200:
      return ['Failed to {} RID ISA ({})'.format(self.mutation, self.status_code)]
    if self.json_result is None:
      return ['Response did not contain valid JSON']
    try:
      _ = self.isa
    except ValueError as e:
      return ['Response returned an invalid ISA: {}'.format(e)]

  @property
  def isa(self) -> rid.IdentificationServiceArea:
    if self.json_result is None:
      raise ValueError('No JSON result present in response from DSS')
    isa_dict = self.json_result.get('service_area', None)
    if not isa_dict:
      raise ValueError('No `service_area` field present in response from DSS')
    return rid.IdentificationServiceArea(isa_dict)

  @property
  def subscribers(self) -> List[rid.SubscriberToNotify]:
    if self.json_result is None:
      raise ValueError('No JSON result present in response from DSS')
    subs = self.json_result.get('subscribers', None)
    if not subs:
      return []
    return [rid.SubscriberToNotify(sub) for sub in subs]

  @property
  def mutation(self) -> str:
    return self['mutation']
yaml.add_representer(MutatedISAResponse, Representer.represent_dict)


class MutatedISA(ImplicitDict):
  """Result of an attempt to mutate an ISA (including DSS & notifications)"""
  dss_response: MutatedISAResponse
  notifications: Dict[str, fetch.Query]


def put_isa(utm_client: infrastructure.UTMClientSession,
            area: s2sphere.LatLngRect,
            start_time: datetime.datetime,
            end_time: datetime.datetime,
            flights_url: str,
            entity_id: str,
            isa_version: Optional[str]=None) -> MutatedISA:
  extents = {
    'spatial_volume': {
      'footprint': {
        'vertices': rid.vertices_from_latlng_rect(area)
      },
      'altitude_lo': 0,
      'altitude_hi': 3048,
    },
    'time_start': start_time.strftime(rid.DATE_FORMAT),
    'time_end': end_time.strftime(rid.DATE_FORMAT),
  }
  body = {
    'extents': extents,
    'flights_url': flights_url,
  }
  if isa_version is None:
    url = '/v1/dss/identification_service_areas/{}'.format(entity_id)
  else:
    url = '/v1/dss/identification_service_areas/{}/{}'.format(entity_id, isa_version)
  dss_response = MutatedISAResponse(fetch.query_and_describe(
    utm_client, 'PUT', url, json=body, scope=rid.SCOPE_WRITE))
  dss_response['mutation'] = 'create' if isa_version is None else 'update'

  # Notify subscribers
  notifications: Dict[str, fetch.Query] = {}
  try:
    subscribers = dss_response.subscribers
    isa = dss_response.isa
  except ValueError:
    subscribers = []
    isa = None
  for subscriber in subscribers:
    body = {
      'service_area': isa,
      'subscriptions': subscriber.subscriptions,
      'extents': extents
    }
    url = '{}/{}'.format(subscriber.url, entity_id)
    notifications[subscriber.url] = fetch.query_and_describe(
      utm_client, 'POST', url, json=body, scope=rid.SCOPE_WRITE)

  return MutatedISA(dss_response=dss_response, notifications=notifications)


def delete_isa(utm_client: infrastructure.UTMClientSession,
               entity_id: str,
               isa_version: str) -> MutatedISA:
  url = '/v1/dss/identification_service_areas/{}/{}'.format(entity_id, isa_version)
  dss_response = MutatedISAResponse(fetch.query_and_describe(
    utm_client, 'DELETE', url, scope=rid.SCOPE_WRITE))
  dss_response['mutation'] = 'delete'

  # Notify subscribers
  notifications: Dict[str, fetch.Query] = {}
  try:
    subscribers = dss_response.subscribers
  except ValueError:
    subscribers = []
  for subscriber in subscribers:
    body = {
      'subscriptions': subscriber.subscriptions
    }
    url = '{}/{}'.format(subscriber.url, entity_id)
    notifications[subscriber.url] = fetch.query_and_describe(
      utm_client, 'POST', url, json=body, scope=rid.SCOPE_WRITE)

  return MutatedISA(dss_response=dss_response, notifications=notifications)
