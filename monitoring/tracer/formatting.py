import copy
import datetime
import enum
import hashlib
import json
from typing import Dict, List, Optional, Tuple

from termcolor import colored


class Change(enum.Enum):
  NOCHANGE = 0
  ADDED = 1
  CHANGED = 2
  REMOVED = 3

  @classmethod
  def color_of(cls, change) -> str:
    if change == Change.NOCHANGE:
      return 'grey'
    elif change == Change.ADDED:
      return 'green'
    elif change == Change.CHANGED:
      return 'yellow'
    elif change == Change.REMOVED:
      return 'red'
    raise ValueError('Invalid Change type')


def _update_overall(overall: Change, field: Change):
  if overall == Change.CHANGED:
    return Change.CHANGED
  if overall == Change.NOCHANGE:
    return field
  if overall == Change.ADDED:
    if field == Change.ADDED or field == Change.NOCHANGE:
      return Change.ADDED
    else:
      return Change.CHANGED
  if overall == Change.REMOVED:
    if field == Change.REMOVED or field == Change.NOCHANGE:
      return Change.REMOVED
    else:
      return Change.CHANGED
  raise ValueError('Unexpected change configuration')


def dict_changes(a: Dict, b: Dict) -> Tuple[Dict, Dict, Change]:
  values = {}
  changes = {}
  overall = Change.NOCHANGE

  for k, v1 in b.items():
    v0 = a.get(k, {})
    if isinstance(v1, dict):
      field_values, field_changes, change = dict_changes(v0, v1)
      if len(field_values) >= 2:
        values[k] = field_values
        changes[k] = field_changes
        changes[k]['__self__'] = change
      elif len(field_values) == 1:
        k = k + '.' + next(iter(values[k].keys()))
        values[k] = field_values
        changes[k] = field_changes
    else:
      if v0 == v1:
        change = Change.NOCHANGE
      else:
        values[k] = v1
        if k not in a:
          change = Change.ADDED
        else:
          change = Change.CHANGED
        changes[k] = change
    overall = _update_overall(overall, change)

  for k, v0 in a.items():
    if k not in b:
      if isinstance(v0, dict):
        values[k], changes[k], change = dict_changes(v0, {})
      else:
        values[k] = v0
        change = Change.REMOVED
        changes[k] = change
      overall = _update_overall(overall, change)

  return values, changes, overall


def diff_lines(values: Dict, changes: Dict) -> List[str]:
  lines = []
  for k, v in values.items():
    c = changes[k]
    if isinstance(v, dict):
      if '__self__' in c:
        lines.append(colored(k, Change.color_of(c['__self__'])) + ':')
      else:
        lines.append(k + ':')
      lines.extend('  ' + line for line in diff_lines(v, c))
    else:
      if c == Change.ADDED:
        lines.append(colored('{}: {}'.format(k, v), 'green'))
      elif c == Change.CHANGED:
        lines.append(k + ': ' + colored(str(v), 'yellow'))
      elif c == Change.REMOVED:
        lines.append(colored('{}: {}'.format(k, v), 'red'))
  return lines


def isa_diff_text(a: Optional[Dict], b: Optional[Dict]) -> str:
  """Create text to display to a real-time user describing a change in ISAs.

  The parameters a and b are ISA "objects" produced by polling.poll_rid_isas in
  a PollingSuccess; a is the previous one, b is the new one.  This function
  should generate text to be printed to a console that summarizes the difference
  between a and b.
  """
  if a is None:
    a = {}
  if b is None:
    b = {}
  values, changes, _ = dict_changes(a, b)
  return '\n'.join(diff_lines(values, changes))


def _abbreviated_entity(entity: Dict) -> Dict:
  entity_lite = copy.deepcopy(entity)

  if 'uss' in entity_lite:
    try:
      details = entity_lite['uss']['details']
      volumes = details['volumes']
      n_circles = sum(1 if v['volume'].get('outline_circle', None) else 0 for v in volumes)
      n_polygons = sum(1 if v['volume'].get('outline_polygon', None) else 0 for v in volumes)
      t_start = min(datetime.datetime.fromisoformat(v['time_start']['value']) for v in volumes)
      t_end = min(datetime.datetime.fromisoformat(v['time_end']['value']) for v in volumes)
      signature = hashlib.sha1(json.dumps(volumes)).hexdigest()[-8:]
      details['volumes'] = {
        'shape': '{} circle{}, {} polygon{} ({})'.format(
            n_circles, '' if n_circles == 1 else 's',
            n_polygons, '' if n_polygons == 1 else 's',
            signature
          ),
        'start': t_start.isoformat(),
        'end': t_end.isoformat(),
      }

      uss_ref = entity_lite['uss']['reference']
      dss_ref = entity_lite['dss']['reference']
      to_remove = []
      for key in uss_ref:
        if key in dss_ref and dss_ref[key] == uss_ref[key]:
          to_remove.append(key)
      for key in to_remove:
        del uss_ref[key]
    except KeyError as e:
      entity_lite['uss'] = 'Response format error: {}'.format(e)

  return entity_lite


def entity_diff_text(a: Optional[Dict], b: Optional[Dict]) -> str:
  """Create text to display to a real-time user describing a change in an Entity.

  The parameters a and b are Operation or Constraint "objects" produced by
  polling.poll_scd_operations or polling.poll_scd_constraints in a
  PollingSuccess; a is the previous one, b is the new one.  This function should
  generate text to be printed to a console that summarizes the difference
  between a and b.
  """
  if a is None:
    a = {}
  if b is None:
    b = {}
  a = _abbreviated_entity(a)
  b = _abbreviated_entity(b)
  values, changes, _ = dict_changes(a, b)
  return '\n'.join(diff_lines(values, changes))


def format_timedelta(td: datetime.timedelta) -> str:
  """Produce a human-readable string describing a timedelta.
  Args:
    td: datetime.timedelta to format.
  Return:
    Formatted timedelta that looks like HH:MM:SS or XXXdHH:MM:SS where XXX is
    number of days, with or without a leading negative sign.
  """
  seconds = int(td.total_seconds())
  if seconds < 0:
    seconds = -seconds
    sign = '-'
  else:
    sign = ''
  periods = (('%d', 60*60*24), ('%02d', 60*60), ('%02d', 60), ('%02d', 1))
  has_days = seconds >= periods[0][1]

  segments = []
  for format_string, period_seconds in periods:
    period_value, seconds = divmod(seconds, period_seconds)
    segments.append(format_string % period_value)

  if has_days:
    return sign + '{:s}d{:s}:{:s}:{:s}'.format(*segments)
  else:
    return sign + '{:s}:{:s}:{:s}'.format(*segments[1:])
