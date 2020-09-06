import datetime
import json
from typing import Callable, Dict, List, Optional

import requests
import s2sphere
from termcolor import colored
import yaml

from monitoring.monitorlib import infrastructure, rid
from monitoring.tracer import geo, tracerlog


def indent(s: str, level: int) -> str:
  return '\n'.join(' ' * level + line for line in s.split('\n'))


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
        lines.extend('    ' + colored(line, 'red') for line in yaml.dump(e.json, indent=2).split('\n'))
      else:
        lines.extend('    ' + colored(line, 'red') for line in e.body.replace('\r\n', '\n').split('\n'))
      return '\n'.join(lines)

    if other is None or other._success is None:
      other = {}
    else:
      other = other._success.objects
    this = self._success.objects
    lines = []
    def add_value(key_text: str, value_text: str) -> None:
      new_lines = value_text.split('\n')
      if len(new_lines) == 1:
        lines.append('{}: {}'.format(key_text, new_lines[0]))
      else:
        lines.append('{}:'.format(key_text))
        lines.extend('  ' + line for line in new_lines)
    for k1, v1 in this.items():
      if k1 not in other:
        add_value('{} {}'.format(name, colored(k1, 'green')), object_diff_text(None, v1))
      elif v1 != other[k1]:
        add_value('{} {}'.format(name, colored(k1, 'yellow')), object_diff_text(other[k1], v1))
    for k0, v0 in other.items():
      if k0 not in this:
        add_value('{} {}'.format(name, colored(k0, 'red')), object_diff_text(v0, None))
    return '\n'.join(lines)


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
  area = rid.geo_polygon_string(rid.vertices_from_latlng_rect(resources.area))
  earliest_time = datetime.datetime.utcnow()
  latest_time = earliest_time + datetime.timedelta(hours=18)
  url = '/v1/dss/identification_service_areas?area={}&earliest_time={}&latest_time={}'.format(
    area, earliest_time.strftime(rid.DATE_FORMAT), latest_time.strftime(rid.DATE_FORMAT))
  t0 = datetime.datetime.utcnow()
  resp = resources.dss_client.get(url, scope=rid.SCOPE_READ)
  t1 = datetime.datetime.utcnow()
  if resp.status_code != 200:
    return PollResult(t0, t1, error=PollError(resp, 'Failed to search ISAs in DSS'))
  try:
    json = resp.json()
  except ValueError:
    return PollResult(t0, t1, error=PollError(resp, 'DSS response to search ISAs was not valid JSON'))
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
