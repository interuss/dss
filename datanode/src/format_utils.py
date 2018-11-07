"""Utilities for managing and converting between formats.


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
from dateutil import parser
import pytz


def format_ts(timestamp=None):
  """Formats a Python datetime as a NASA-style string.

  Args:
    timestamp: Python datetime to format; defaults to now

  Returns:
    String formatted like YYYY-mm-ddTHH:MM:SS.fffZ
  """
  r = datetime.datetime.now(pytz.utc) if timestamp is None else timestamp
  r = r.astimezone(pytz.utc)
  return '{0}Z'.format(r.strftime('%Y-%m-%dT%H:%M:%S.%f')[:23])


def parse_timestamp(timestamp_str):
  """Parses a timestamp into a Python datetime.

  Args:
    timestamp_str: timestamp string, with or without Z suffix

  Returns:
    Python datetime representation of timestamp
  """
  timestamp = parser.parse(timestamp_str)
  if timestamp.tzinfo is None:
    timestamp = timestamp.replace(tzinfo=pytz.utc)
  return timestamp
