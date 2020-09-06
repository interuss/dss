import datetime
from typing import Callable, Dict, Optional

import requests
import s2sphere
from termcolor import colored
import yaml

from monitoring.monitorlib import infrastructure, rid, scd
from monitoring.tracer import formatting, tracerlog


def indent(s: str, level: int) -> str:
  return '\n'.join(' ' * level + line for line in s.split('\n'))


class ResourceSet(object):
  def __init__(self,
               dss_client: infrastructure.DSSTestSession,
               area: s2sphere.LatLngRect,
               logger: tracerlog.Logger,
               start_time: datetime.datetime,
               end_time: datetime.datetime):
    self.dss_client = dss_client
    self.area = area
    self.logger = logger
    self.start_time = start_time
    self.end_time = end_time

    self.scd_op_cache = {}


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

  def __eq__(self, other):
    return isinstance(other, PollError) and self.to_json() == other.to_json()

  def __ne__(self, other):
    return not self == other

  def __str__(self) -> str:
    return '{} after {} at {}'.format(self.description, self.code, self.url)

  def to_json(self) -> Dict:
    return {
      'description': self.description,
      'url': self.url,
      'code': self.code,
      'json': self.json,
      'body': self.body,
    }

class PollSuccess(object):
  def __init__(self, objects: Dict):
    self.objects = objects

  def __str__(self) -> str:
    return '{} objects'.format(len(self.objects))

  def __eq__(self, other):
    return isinstance(other, PollSuccess) and self.to_json() == other.to_json()

  def __ne__(self, other):
    return not self == other

  def to_json(self) -> Dict:
    return {
      'objects': self.objects
    }


class PollResult(object):
  def __init__(self, initiated_at: datetime.datetime, completed_at: datetime.datetime, success: PollSuccess=None, error: PollError=None):
    self.initiated_at = initiated_at
    if success is None and error is None:
      raise ValueError('A poll result must indicate either success or error')
    if success is not None and error is not None:
      raise ValueError('A poll result may not indicate both success and error')
    self._success = success
    self._error = error
    self.completed_at = completed_at

  def __str__(self):
    if self._success is not None:
      return 'Success {}'.format(self._success)
    return 'Error {}'.format(self._error)

  def to_json(self) -> Dict:
    return  {
      't0': self.initiated_at.isoformat(),
      't1': self.completed_at.isoformat(),
      'error': self._error.to_json() if self._error else None,
      'success': self._success.to_json() if self._success else None,
    }

  def has_different_content_than(self, other) -> bool:
    if other is None:
      return True
    return self._error != other._error or self._success != other._success

  def diff_text(self, other, name: str, object_diff_text: Callable[[Optional[Dict], Optional[Dict]], str]) -> str:
    dt_seconds = round((self.completed_at - self.initiated_at).total_seconds(), 2)

    if self._error and not other._error:
      e = self._error
      lines = [colored('{} {}'.format(name, e.description), 'white', 'on_red')]
      lines.append('  {} ({} s) {}\n'.format(colored(str(e.code), 'red'), dt_seconds, e.url))
      if e.json:
        lines.extend('    ' + colored(line, 'red') for line in yaml.dump(e.json, indent=2).strip().split('\n'))
      else:
        lines.extend('    ' + colored(line, 'red') for line in e.body.replace('\r\n', '\n').split('\n'))
      return '\n'.join(lines)

    if other is None or other._success is None:
      other = {}
    else:
      other = {name + ' ' + k: v for k, v in other._success.objects.items()}
    this = {name + ' ' + k: v for k, v in self._success.objects.items()}
    values, changes, _ = formatting.dict_changes(other, this)
    return '\n'.join(formatting.diff_lines(values, changes))


class Poller(object):
  def __init__(self, name: str,
               object_diff_text: Callable[[Optional[Dict], Optional[Dict]], str],
               interval: datetime.timedelta,
               poll: Callable[[], PollResult]):
    self.name = name
    self._object_diff_text = object_diff_text
    self._interval = interval
    self._poll = poll
    self._next_poll: Optional[datetime.datetime] = None
    self.last_result = None

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

  def diff_text(self, new_result: PollResult) -> str:
    return new_result.diff_text(self.last_result, self.name, self._object_diff_text)


def poll_rid_isas(resources: ResourceSet) -> PollResult:
  # Query DSS for ISAs in 2D+time area of interest
  area = rid.geo_polygon_string(rid.vertices_from_latlng_rect(resources.area))
  url = '/v1/dss/identification_service_areas?area={}&earliest_time={}&latest_time={}'.format(
    area, resources.start_time.strftime(rid.DATE_FORMAT), resources.end_time.strftime(rid.DATE_FORMAT))
  t0 = datetime.datetime.utcnow()
  resp = resources.dss_client.get(url, scope=rid.SCOPE_READ)
  t1 = datetime.datetime.utcnow()

  # Handle overall errors
  if resp.status_code != 200:
    return PollResult(t0, t1, error=PollError(resp, 'Failed to search ISAs in DSS'))
  try:
    json = resp.json()
  except ValueError:
    return PollResult(t0, t1, error=PollError(resp, 'DSS response to search ISAs was not valid JSON'))

  # Extract ISAs from response
  isa_list = json.get('service_areas', [])
  isas = {}
  for isa in isa_list:
    if 'id' not in isa:
      return PollResult(t0, t1, error=PollError(resp, 'DSS response to search ISAs included ISA without id'))
    if 'owner' not in isa:
      return PollResult(t0, t1, error=PollError(resp, 'DSS response to search ISAs included ISA without owner'))
    isa_id = '{} ({})'.format(isa['id'], isa['owner'])
    del isa['id']
    del isa['owner']
    isas[isa_id] = isa
  return PollResult(t0, t1, success=PollSuccess(isas))


def poll_scd_operations(resources: ResourceSet) -> PollResult:
  # Query DSS for Operations in 4D volume of interest
  request_body = {
    'area_of_interest': scd.make_vol4(
      resources.start_time, resources.end_time, 0, 3048,
      polygon=scd.make_polygon(latlngrect=resources.area)
    )
  }
  url = '/dss/v1/operation_references/query'
  t0 = datetime.datetime.utcnow()
  resp = resources.dss_client.post(url, json=request_body, scope=scd.SCOPE_SC)
  t1 = datetime.datetime.utcnow()

  # Handle any errors
  if resp.status_code != 200:
    return PollResult(t0, t1, error=PollError(resp, 'Failed to search Operations in DSS'))
  try:
    ref_json = resp.json()
  except ValueError:
    return PollResult(t0, t1, error=PollError(resp, 'DSS response to search Operations was not valid JSON'))
  op_ref_list = ref_json.get('operation_references', [])
  for op_ref in op_ref_list:
    if 'id' not in op_ref:
      return PollResult(t0, t1, error=PollError(resp, 'DSS response to search Operations included Operation without id'))
    if 'owner' not in op_ref:
      return PollResult(t0, t1, error=PollError(resp, 'DSS response to search Operations included ISA without owner'))
    if 'uss_base_url' not in op_ref:
      return PollResult(t0, t1, error=PollError(resp, 'DSS response to search Operations included ISA without uss_base_url'))

  # Obtain details for all Operations (using cache when appropriate)
  ops = {}
  for op_ref in op_ref_list:
    op_key = '{} ({})'.format(op_ref['id'], op_ref['owner'])

    if (op_key in resources.scd_op_cache and
        resources.scd_op_cache[op_key]['dss']['reference'] == op_ref and
        'error' not in resources.scd_op_cache[op_key]['uss']):
      # Operation reference data in DSS is identical to the cached reference; do
      # not re-retrieve Operation details from USS
      ops[op_key] = resources.scd_op_cache[op_key]
      continue

    ops[op_key] = {'dss': {'reference': op_ref}}

    # Query the USS for Operation details
    details_url = op_ref['uss_base_url'] + '/uss/v1/operations/{}'.format(op_ref['id'])
    t2 = datetime.datetime.utcnow()
    details_resp = resources.dss_client.get(details_url, scope=scd.SCOPE_SC)
    t3 = datetime.datetime.utcnow()

    # Handle any errors
    details_error_condition = None
    try:
      details_json = details_resp.json()
    except ValueError:
      details_json = None
      details_error_condition = 'did not return valid JSON'
    if resp.status_code != 200:
      details_error_condition = 'indicated failure'
    if not details_error_condition:
      if 'reference' not in details_json:
        details_error_condition = 'did not contain reference field'
      if 'details' not in details_json:
        details_error_condition = 'did not contain details field'
    if details_error_condition:
      ops[op_key]['uss']['error'] = {
        'description': 'USS query for Operation details ' + details_error_condition,
        'url': details_url,
        'code': resp.status_code,
      }
      if details_json is not None:
        ops[op_key]['uss']['error']['json'] = details_json
      else:
        ops[op_key]['uss']['error']['body'] = details_resp.content
      continue

    # Record details, and information about querying details, in the result
    ops[op_key]['uss'] = details_json
    ops[op_key]['uss']['tracer'] = {
      'time_queried': t2.isoformat(),
      'dt_s': round((t3 - t2).total_seconds(), 2),
    }

    # Cache the full result for this Operation
    resources.scd_op_cache[op_key] = ops[op_key]
  return PollResult(t0, t1, success=PollSuccess(ops))
