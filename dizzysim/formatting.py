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
