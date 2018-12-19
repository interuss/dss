"""Utilities for printing data from TCL4 InterUSS Platform Data Nodes.


Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

import datetime

import pytz
import termcolor


def parse_timestamp(timestamp):
  """Parse a timestamp from the data node into a tz-aware datetime.

  Args:
    timestamp: String describing timestamp.

  Returns:
    Time-zone aware datetime in UTC time zone.
  """
  return datetime.datetime.strptime(
    timestamp, '%Y-%m-%dT%H:%M:%S.%fZ').replace(tzinfo=pytz.UTC)


def format_timedelta(td):
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


def print_operator_list(operators, now):
  """Print human-readable summary of operators.

  Timestamps are printed as timedeltas relative to the provided now, and are
  highlighted when there may be a problem (start time too far in the future, end
  time in the past)

  Args:
    operators: List of TCL4 InterUSS Platform Operator dicts.
    now: The datetime relative to which timestamps should be displayed.
  """
  for operator in operators:
    time_begin = parse_timestamp(operator['minimum_operation_timestamp'])
    time_end = parse_timestamp(operator['maximum_operation_timestamp'])
    td_begin = time_begin - now
    td_end = time_end - now
    begin_text = format_timedelta(time_begin - now)
    end_text = format_timedelta(time_end - now)
    if td_begin > datetime.timedelta(minutes=5):
      begin_text = termcolor.colored(begin_text, 'red')
    elif td_begin.total_seconds() > 0:
      begin_text = termcolor.colored(begin_text, 'yellow')
    if td_end.total_seconds() < 0:
      end_text = termcolor.colored(end_text, 'red')
    print '== %s %d(%d, %d) v%d %s (%s => %s)' % (
      termcolor.colored(operator['uss'], 'blue'),
      operator['zoom'], operator['x'], operator['y'], operator['version'],
      operator['announcement_level'], begin_text, end_text)
    for operation in operator['operations']:
      time_begin = parse_timestamp(operation['effective_time_begin'])
      time_end = parse_timestamp(operation['effective_time_end'])
      td_begin = time_begin - now
      td_end = time_end - now
      begin_text = format_timedelta(time_begin - now)
      end_text = format_timedelta(time_end - now)
      if td_begin > datetime.timedelta(minutes=5):
        begin_text = termcolor.colored(begin_text, 'red')
      elif td_begin.total_seconds() > 0:
        begin_text = termcolor.colored(begin_text, 'yellow')
      if td_end.total_seconds() < 0:
        end_text = termcolor.colored(end_text, 'red')
      print '     %s (%s => %s)' % (
        termcolor.colored(operation['gufi'], 'magenta'),
        begin_text, end_text)


def print_uvrs(uvrs, now):
  """Print human-readable summary of UVRs.

  Timestamps are printed as timedeltas relative to the provided now, and are
  highlighted when there may be a problem (start time too far in the future, end
  time in the past)

  Args:
    uvrs: List of TCL4 InterUSS Platform UVR dicts.
    now: The datetime relative to which timestamps should be displayed.
  """
  for uvr in uvrs:
    time_begin = parse_timestamp(uvr['effective_time_begin'])
    time_end = parse_timestamp(uvr['effective_time_begin'])
    td_begin = time_begin - now
    td_end = time_end - now
    begin_text = format_timedelta(time_begin - now)
    end_text = format_timedelta(time_end - now)
    if td_begin > datetime.timedelta(minutes=5):
      begin_text = termcolor.colored(begin_text, 'red')
    elif td_begin.total_seconds() > 0:
      begin_text = termcolor.colored(begin_text, 'yellow')
    if td_end.total_seconds() < 0:
      end_text = termcolor.colored(end_text, 'red')
    print '++ %s %s %.0f %s %s-%.0f %s %s (%s => %s) %s' % (
        termcolor.colored('UVR', 'green'),
        termcolor.colored(uvr['uss_name'], 'blue'),
        uvr['min_altitude']['altitude_value'],
        uvr['min_altitude']['units_of_measure'],
        uvr['min_altitude']['vertical_reference'],
        uvr['max_altitude']['altitude_value'],
        uvr['max_altitude']['units_of_measure'],
        uvr['max_altitude']['vertical_reference'],
        begin_text, end_text, uvr['message_id'])


def print_operators(metadata):
  """Print human-readable summary of all operators in provided metadata.

  Inactive (no-operation) operators are displayed first in alphabetical order,
  followed by active (at least one operation) operators in alphabetical order.

  Args:
    metadata: USS Metadata dict translated from JSON returned from
      GridCellOperators endpoint.
  """
  now = pytz.timezone('US/Pacific').localize(datetime.datetime.now())
  print('%s @ %s received %s' % (
    termcolor.colored('sync_token %d' % metadata['sync_token'], 'cyan', attrs=['bold']),
    metadata['data']['timestamp'], str(now)))
  passive_operators = sorted([o for o in metadata['data']['operators']
                              if not o['operations']], key=lambda o: o['uss'])
  active_operators = sorted([o for o in metadata['data']['operators']
                             if o['operations']], key=lambda o: o['uss'])
  print_operator_list(passive_operators, now)
  print_uvrs(metadata['data']['uvrs'], now)
  print_operator_list(active_operators, now)
