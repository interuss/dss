#!env/bin/python3

import argparse
import datetime
import os
import sys
import time
from typing import Callable, Dict, List, Optional

import requests
import s2sphere

from monitoring.monitorlib import auth, infrastructure, rid
from monitoring.tracer import geo, tracerlog


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Test Interoperability of DSSs")

    # Required arguments
    parser.add_argument('--auth', help='Auth spec for obtaining authorization to DSS and USSs; see README.md')
    parser.add_argument('--dss', help='Base URL of DSS instance to query')
    parser.add_argument('--area', help='`lat,lng,lat,lng` for box containing the area to trace interactions for')
    parser.add_argument('--output-folder', help='Path of folder in which to write logs')

    # Feature arguments
    parser.add_argument('--rid-isa-poll-interval', type=float, default=0, help='Seconds beteween each poll of the DSS for ISAs, 0 to disable DSS polling for ISAs')
    parser.add_argument('--rid-subscription-poll-interval', type=float, default=0, help='Seconds beteween each poll of the DSS for RID Subscriptions, 0 to disable DSS polling for RID Subscriptions')
    parser.add_argument('--scd-operation-poll-interval', type=float, default=0, help='Seconds between each poll of the DSS for Operations, 0 to disable DSS polling for Operations')
    parser.add_argument('--scd-constraint-poll-interval', type=float, default=0, help='Seconds between each poll of the DSS for Constraints, 0 to disable DSS polling for Constraints')
    parser.add_argument('--scd-subscription-poll-interval', type=float, default=0, help='Seconds beteween each poll of the DSS for SCD Subscriptions, 0 to disable DSS polling for SCD Subscriptions')

    return parser.parse_args()


class ResourceSet(object):
  def __init__(self, dss_client: infrastructure.DSSTestSession, area: s2sphere.LatLngRect, logger: tracerlog.Logger):
    self.dss_client = dss_client
    self.area = area
    self.logger = logger


class PollError(object):
  def __init__(self, resp: requests.Response, description: str):
    self.description = description
    self.url = resp.url
    self.code = resp.status_code
    try:
      self.json = resp.json()
      self.body = None
    except ValueError:
      self.body = resp.content.decode('utf-8')
      self.json = None

  def __str__(self):
    return '{} after {} at {}'.format(self.description, self.code, self.url)


class PollSuccess(object):
  def __init__(self, objects: Dict):
    self._objects = objects

  def __str__(self):
    return '{} objects'.format(len(self._objects))


class PollResult(object):
  def __init__(self, success: PollSuccess=None, error: PollError=None):
    if success is None and error is None:
      raise ValueError('A poll result must indicate either success or error')
    if success is not None and error is not None:
      raise ValueError('A poll result may not indicate both success and error')
    self._success = success
    self._error = error

  def __str__(self):
    if self._success is not None:
      return 'Success {}'.format(self._success)
    return 'Error {}'.format(self._error)


class Poller(object):
  def __init__(self, name: str, interval: datetime.timedelta, poll: Callable[[], PollResult]):
    self.name = name
    self._interval = interval
    self._poll = poll
    self._next_poll: Optional[datetime.datetime] = None

  def time_to_next_poll(self) -> datetime.timedelta:
    if self._next_poll is None:
      return datetime.timedelta(seconds=0)
    now = datetime.datetime.utcnow()
    return self._next_poll - now

  def poll(self) -> PollResult:
    if self._next_poll is None:
      self._next_poll = datetime.datetime.utcnow() + self._interval
    else:
      now = datetime.datetime.utcnow()
      while self._next_poll < now:
        self._next_poll += self._interval
    return self._poll()


def poll_rid_isas(resources: ResourceSet) -> PollResult:
  area = rid.geo_polygon_string(rid.vertices_from_latlng_rect(resources.area))
  earliest_time = datetime.datetime.utcnow()
  latest_time = earliest_time + datetime.timedelta(hours=18)
  url = '/v1/dss/identification_service_areas?area={}&earliest_time={}&latest_time={}'.format(
    area, earliest_time.strftime(rid.DATE_FORMAT), latest_time.strftime(rid.DATE_FORMAT))
  resp = resources.dss_client.get(url, scope=rid.SCOPE_READ)
  if resp.status_code != 200:
    return PollResult(error=PollError(resp, 'Failed to search ISAs in DSS'))
  try:
    json = resp.json()
  except ValueError:
    return PollResult(error=PollError(resp, 'DSS response to search ISAs was not valid JSON'))
  isa_list = json.get('service_areas', [])
  isas = {}
  for isa in isa_list:
    if 'id' not in isa:
      return PollResult(error=PollError(resp, 'DSS response to search ISAs included ISA without id'))
    if 'owner' not in isa:
      return PollResult(error=PollError(resp, 'DSS response to search ISAs included ISA without owner'))
    isa_id = '{} ({})'.format(isa['id'], isa['owner'])
    del isa['id']
    del isa['owner']
    isas[isa_id] = isa
  print(isas)
  return PollResult(success=PollSuccess(isas))


def main() -> int:
    args = parseArgs()

    # Required resources
    adapter: auth.AuthAdapter = auth.make_auth_adapter(args.auth)
    dss_client = infrastructure.DSSTestSession(args.dss, adapter)
    area: s2sphere.LatLngRect = geo.make_latlng_rect(args.area)
    logger = tracerlog.Logger(args.output_folder)
    resources = ResourceSet(dss_client, area, logger)

    # Prepare pollers
    pollers: List[Poller] = []

    if args.rid_isa_poll_interval > 0:
      pollers.append(Poller(
        name='RID ISA',
        interval=datetime.timedelta(seconds=args.rid_isa_poll_interval),
        poll=lambda: poll_rid_isas(resources)))

    # Execute the polling loop
    abort = False
    while not abort:
      try:
        most_urgent_dt = datetime.timedelta(days=999999999)
        most_urgent_poller = None
        for poller in pollers:
          dt = poller.time_to_next_poll()
          if dt < most_urgent_dt:
            most_urgent_poller = poller
            most_urgent_dt = dt

        if most_urgent_dt.total_seconds() > 0:
          print('Sleeping for {} seconds'.format(most_urgent_dt.total_seconds()))
          time.sleep(most_urgent_dt.total_seconds())

        print('Polling {}'.format(most_urgent_poller.name))
        result = most_urgent_poller.poll()
        print('Result: {}'.format(result))
        print('Finished polling {}'.format(most_urgent_poller.name))
      except KeyboardInterrupt:
        abort = True

    return os.EX_OK

if __name__ == "__main__":
    sys.exit(main())
