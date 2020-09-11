import argparse
import atexit
import datetime
import logging
import os
import shlex
import signal
import sys
import threading
import time
from typing import Dict, Optional

import requests

from monitoring.monitorlib import ids, rid, scd, versioning
from monitoring.tracer.resources import ResourceSet


ENV_OPTIONS = 'TRACER_OPTIONS'
RID_SUBSCRIPTION_ID_CODE = 'tracer RID Subscription'
SCD_SUBSCRIPTION_ID_CODE = 'tracer SCD Subscription'

logging.basicConfig()
_logger = logging.getLogger('tracer.context')
_logger.setLevel(logging.DEBUG)

resources: Optional[ResourceSet] = None


class SubscriptionManagementError(RuntimeError):
  def __init__(self, msg):
    super(SubscriptionManagementError, self).__init__(msg)


def init() -> None:
  if not os.environ.get(ENV_OPTIONS, None):
    raise ValueError('{} environment variable must be specified'.format(ENV_OPTIONS))

  parser = argparse.ArgumentParser(description="Subscribe to changes in DSS-tracked Entity status")

  ResourceSet.add_arguments(parser)

  parser.add_argument('--base-url', help='Base URL at which this server may be reached externally')
  parser.add_argument('--monitor-rid', action='store_true', default=False, help='If specified, monitor ISA activity per the remote ID standard')
  parser.add_argument('--monitor-scd', action='store_true', default=False, help='If specified, monitor Operation and Constraint activity per the strategic deconfliction standard')

  args = parser.parse_args(shlex.split(os.environ[ENV_OPTIONS]))

  global resources
  resources = ResourceSet.from_arguments(args)

  config = vars(args)
  config['code_version'] = versioning.get_code_version()
  resources.logger.logconfig(config)

  try:
    _logger.info('Establishing Subscriptions from PID {} at {}...'.format(os.getpid(), datetime.datetime.utcnow()))
    _subscribe(resources, args.base_url, args.monitor_rid, args.monitor_scd)
    _logger.info('Subscriptions established.')
  except SubscriptionManagementError as e:
    msg = 'Failed to initialize: {}'.format(e)
    _logger.error(msg)
    sys.stderr.write(msg)
    sys.exit(-1)

  cleanup = {
    'lock': threading.Lock(),
    'complete': False,
  }
  def shutdown(signal_number, stack_frame) -> None:
    with cleanup['lock']:
      if cleanup['complete']:
        return
      _logger.info('Cleaning up Subscriptions from PID {} at {}...'.format(os.getpid(), datetime.datetime.utcnow()))
      _unsubscribe(resources, args.monitor_rid, args.monitor_scd)
      _logger.info('Subscription cleanup complete.')
      cleanup['complete'] = True
  atexit.register(shutdown, None, None)
  for sig in (signal.SIGABRT, signal.SIGINT, signal.SIGTERM):
    signal.signal(sig, shutdown)

  dt = (resources.end_time - datetime.datetime.utcnow()).total_seconds()
  def terminate_at_expiration():
    time.sleep(dt)
    _logger.info('Terminating server at expiration of Subscription(s)')
    os.kill(os.getpid(), signal.SIGINT)
  threading.Thread(target=terminate_at_expiration, daemon=True).start()


def _subscribe(resources: ResourceSet, base_url: str, monitor_rid: bool, monitor_scd: bool) -> None:
  if base_url.endswith('/'):
    base_url = base_url[0:-1]
  if monitor_rid:
    _subscribe_rid(resources, base_url)
  if monitor_scd:
    _subscribe_scd(resources, base_url)


def _unsubscribe(resources: ResourceSet, monitor_rid: bool, monitor_scd: bool) -> None:
  if monitor_rid:
    _clear_existing_rid_subscription(resources, 'cleanup')
  if monitor_scd:
    _clear_existing_scd_subscription(resources, 'cleanup')


def _describe_response(resp: requests.Response, description: str) -> Dict:
  info = {
    'description': description,
    'url': resp.url,
    'code': resp.status_code,
  }
  try:
    info['json'] = resp.json()
  except ValueError:
    info['body'] = resp.content
  return info


def _rid_subscription_url():
  sub_id = ids.make_id(RID_SUBSCRIPTION_ID_CODE)
  return '/v1/dss/subscriptions/{}'.format(sub_id)


def _subscribe_rid(resources: ResourceSet, callback_url: str) -> None:
  _clear_existing_rid_subscription(resources, 'old')

  body = {
    'extents': {
      'spatial_volume': {
        'footprint': {
          'vertices': rid.vertices_from_latlng_rect(resources.area)
        },
        'altitude_lo': 0,
        'altitude_hi': 3048,
      },
      'time_start': resources.start_time.strftime(rid.DATE_FORMAT),
      'time_end': resources.end_time.strftime(rid.DATE_FORMAT),
    },
    'callbacks': {
      'identification_service_area_url': callback_url
    },
  }
  resp = resources.dss_client.put(_rid_subscription_url(), json=body, scope=rid.SCOPE_READ)
  if resp.status_code != 200:
    msg = 'Failed to create RID Subscription'
    msg += ' -> ' + resources.logger.log_new('ridsubscription', _describe_response(resp, msg))
    raise SubscriptionManagementError(msg)

  msg = 'Created RID Subscription successfully'
  resources.logger.log_new('ridsubscription', _describe_response(resp, msg))


def _clear_existing_rid_subscription(resources: ResourceSet, suffix: str):
  url = _rid_subscription_url()

  resp = resources.dss_client.get(url, scope=rid.SCOPE_READ)
  if resp.status_code == 404:
    return # This is the expected condition (no pre-existing Subscription)
  elif resp.status_code == 200:
    # There is a pre-existing Subscription; delete it
    try:
      resp_json = resp.json()
    except ValueError:
      msg = 'Response to get existing RID Subscription did not return valid JSON'
      msg += ' -> ' + resources.logger.log_new('ridsubscription_{}'.format(suffix), _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)
    version = resp_json.get('subscription', {}).get('version', None)
    if not version:
      msg = 'Response to get existing RID Subscription did not include a version'
      msg += ' -> ' + resources.logger.log_new('ridsubscription_old', _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)

    resources.logger.log_new('ridsubscription_{}'.format(suffix), _describe_response(resp, 'RID Subscription retrieved successfully'))

    del_url = url + '/{}'.format(resp_json['subscription']['version'])
    resp = resources.dss_client.delete(del_url, scope=rid.SCOPE_READ)
    if resp.status_code != 200:
      msg = 'Response to delete existing RID Subscription indicated {}'.format(resp.status_code)
      msg += ' -> ' + resources.logger.log_new('ridsubscription_{}_del'.format(suffix), _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)

    resources.logger.log_new('ridsubscription_{}_del'.format(suffix), _describe_response(resp, 'RID Subscription deleted successfully'))
  else:
    # We expected to get a 200 or 404 but got something else instead
    msg = 'Response to get existing RID Subscription did not return 200 or 404'
    msg += ' -> ' + resources.logger.log_new('ridsubscription_{}'.format(suffix), _describe_response(resp, msg))
    raise SubscriptionManagementError(msg)


def _scd_subscription_url():
  sub_id = ids.make_id(SCD_SUBSCRIPTION_ID_CODE)
  return '/dss/v1/subscriptions/{}'.format(sub_id)


def _subscribe_scd(resources: ResourceSet, base_url: str) -> None:
  _clear_existing_scd_subscription(resources, 'old')

  body = {
    'extents': scd.make_vol4(
        resources.start_time, resources.end_time, 0, 3048,
        polygon=scd.make_polygon(latlngrect=resources.area)),
    'old_version': 0,
    'uss_base_url': base_url,
    'notify_for_operations': True,
    'notify_for_constraints': True,
  }
  resp = resources.dss_client.put(_scd_subscription_url(), json=body, scope=scd.SCOPE_SC)
  if resp.status_code != 200:
    msg = 'Failed to create SCD Subscription'
    msg += ' -> ' + resources.logger.log_new('scdsubscription', _describe_response(resp, msg))
    raise SubscriptionManagementError(msg)

  msg = 'Created SCD Subscription successfully'
  resources.logger.log_new('scdsubscription', _describe_response(resp, msg))


def _clear_existing_scd_subscription(resources: ResourceSet, suffix: str):
  url = _scd_subscription_url()

  resp = resources.dss_client.get(url, scope=scd.SCOPE_SC)
  if resp.status_code == 404:
    return # This is the expected condition (no pre-existing Subscription)
  elif resp.status_code == 200:
    # There is a pre-existing Subscription; delete it
    try:
      resp_json = resp.json()
    except ValueError:
      msg = 'Response to get existing SCD Subscription did not return valid JSON'
      msg += ' -> ' + resources.logger.log_new('scdsubscription_{}'.format(suffix), _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)
    version = resp_json.get('subscription', {}).get('version', None)
    if version is None:
      msg = 'Response to get existing SCD Subscription did not include a version'
      msg += ' -> ' + resources.logger.log_new('scdsubscription_old', _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)

    resources.logger.log_new('scdsubscription_{}'.format(suffix), _describe_response(resp, 'SCD Subscription retrieved successfully'))

    resp = resources.dss_client.delete(url, scope=scd.SCOPE_SC)
    if resp.status_code != 200:
      msg = 'Response to delete existing SCD Subscription indicated {}'.format(resp.status_code)
      msg += ' -> ' + resources.logger.log_new('scdsubscription_{}_del'.format(suffix), _describe_response(resp, msg))
      raise SubscriptionManagementError(msg)

    resources.logger.log_new('scdsubscription_{}_del'.format(suffix), _describe_response(resp, 'SCD Subscription deleted successfully'))
  else:
    # We expected to get a 200 or 404 but got something else instead
    msg = 'Response to get existing SCD Subscription did not return 200 or 404'
    msg += ' -> ' + resources.logger.log_new('scdsubscription_{}'.format(suffix), _describe_response(resp, msg))
    raise SubscriptionManagementError(msg)
