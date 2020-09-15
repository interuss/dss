import datetime
from typing import Any, Callable, Dict, Optional

import s2sphere
from termcolor import colored
import yaml

from monitoring.monitorlib import fetch
import monitoring.monitorlib.fetch.rid
import monitoring.monitorlib.fetch.scd
from monitoring.tracer.resources import ResourceSet


def indent(s: str, level: int) -> str:
  return '\n'.join(' ' * level + line for line in s.split('\n'))


class Poller(object):
  def __init__(self, name: str,
               object_diff_text: Callable[[Any, Any], str],
               interval: datetime.timedelta,
               poll: Callable[[], Any]):
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

  def poll(self) -> Any:
    if self._next_poll is None:
      self._next_poll = datetime.datetime.utcnow() + self._interval
    else:
      now = datetime.datetime.utcnow()
      while self._next_poll < now:
        self._next_poll += self._interval
    return self._poll()

  def diff_text(self, new_result: Any) -> str:
    return self._object_diff_text(self.last_result, new_result)


def poll_rid_isas(resources: ResourceSet, box: s2sphere.LatLngRect) -> Any:
  result = fetch.rid.isas(resources.dss_client, box, resources.start_time, resources.end_time)
  result['log_type'] = 'poll_isas'
  return result


def poll_scd_operations(resources: ResourceSet) -> Any:
  if 'operations' not in resources.scd_cache:
    resources.scd_cache['operations']: Dict[str, fetch.scd.FetchedEntity] = {}
  result = fetch.scd.operations(
    resources.dss_client, resources.area, resources.start_time,
    resources.end_time, operation_cache=resources.scd_cache['operations'])
  result['log_type'] = 'poll_operations'
  return result


def poll_scd_constraints(resources: ResourceSet) -> Any:
  if 'constraints' not in resources.scd_cache:
    resources.scd_cache['constraints']: Dict[str, fetch.scd.FetchedEntity] = {}
  result = fetch.scd.constraints(
    resources.dss_client, resources.area, resources.start_time,
    resources.end_time, constraint_cache=resources.scd_cache['constraints'])
  result['log_type'] = 'poll_constraints'
  return result
