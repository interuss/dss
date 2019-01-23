"""Utilities to format data according to API.

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

def timestamp(timestamp=None):
  r = datetime.datetime.now(pytz.utc) if timestamp is None else timestamp
  if r.tzinfo:
    r = r.astimezone(pytz.utc)
  else:
    r = pytz.utc.localize(r)
  s = r.isoformat()
  if '.' in s:
    s = '%s.%03d%s' % (s[:s.index('.')], r.microsecond // 1000, s[-6:])
  else:
    s = '%s.%03d%s' % (s[:-6], r.microsecond // 1000, s[-6:])
  return s
