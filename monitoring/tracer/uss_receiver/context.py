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
from typing import Optional

from monitoring.monitorlib import ids, versioning
from monitoring.monitorlib import fetch
import monitoring.monitorlib.fetch.rid
import monitoring.monitorlib.fetch.scd
from monitoring.monitorlib import mutate
import monitoring.monitorlib.mutate.rid
import monitoring.monitorlib.mutate.scd
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
  resources.logger.log_new('subscribe_start', config)

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
      resources.logger.log_new('subscribe_stop', {
        'timestamp': datetime.datetime.utcnow(),
        'signal_number': signal_number,
      })
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


def _rid_subscription_id() -> str:
  sub_id = ids.make_id(RID_SUBSCRIPTION_ID_CODE)
  return str(sub_id)


RID_SUBSCRIPTION_KEY = 'subscribe_ridsubscription'

def _subscribe_rid(resources: ResourceSet, callback_url: str) -> None:
  _clear_existing_rid_subscription(resources, 'old')

  create_result = mutate.rid.put_subscription(
    resources.dss_client, resources.area, resources.start_time,
    resources.end_time, callback_url, _rid_subscription_id())
  resources.logger.log_new(RID_SUBSCRIPTION_KEY, create_result)
  if not create_result.success:
    raise SubscriptionManagementError('Could not create RID Subscription')


def _clear_existing_rid_subscription(resources: ResourceSet, suffix: str) -> None:
  existing_result = fetch.rid.subscription(resources.dss_client, _rid_subscription_id())
  logfile = resources.logger.log_new('{}_{}_get'.format(RID_SUBSCRIPTION_KEY, suffix), existing_result)
  if not existing_result.success:
    raise SubscriptionManagementError('Could not query existing RID Subscription -> {}'.format(logfile))

  if existing_result.subscription is not None:
    del_result = mutate.rid.delete_subscription(
      resources.dss_client, _rid_subscription_id(), existing_result.subscription.version)
    logfile = resources.logger.log_new('{}_{}_del'.format(RID_SUBSCRIPTION_KEY, suffix), del_result)
    if not del_result.success:
      raise SubscriptionManagementError('Could not delete existing RID Subscription -> {}'.format(logfile))


SCD_SUBSCRIPTION_KEY = 'subscribe_scdsubscription'

def _scd_subscription_id() -> str:
  sub_id = ids.make_id(SCD_SUBSCRIPTION_ID_CODE)
  return str(sub_id)


def _subscribe_scd(resources: ResourceSet, base_url: str) -> None:
  _clear_existing_scd_subscription(resources, 'old')

  create_result = mutate.scd.put_subscription(
    resources.dss_client, resources.area, resources.start_time,
    resources.end_time, base_url, _scd_subscription_id())
  logfile = resources.logger.log_new(SCD_SUBSCRIPTION_KEY, create_result)
  if not create_result.success:
    raise SubscriptionManagementError('Could not create new SCD Subscription -> {}'.format(logfile))


def _clear_existing_scd_subscription(resources: ResourceSet, suffix: str) -> None:
  get_result = fetch.scd.subscription(resources.dss_client, _scd_subscription_id())
  logfile = resources.logger.log_new('{}_{}_get'.format(SCD_SUBSCRIPTION_KEY, suffix), get_result)
  if not get_result.success:
    raise SubscriptionManagementError('Could not query existing SCD Subscription -> {}'.format(logfile))

  if get_result.subscription is not None:
    del_result = mutate.scd.delete_subscription(resources.dss_client, _scd_subscription_id())
    logfile = resources.logger.log_new('{}_{}'.format(SCD_SUBSCRIPTION_KEY, suffix), del_result)
    if not del_result.success:
      raise SubscriptionManagementError('Could not delete existing SCD Subscription -> {}'.format(logfile))
