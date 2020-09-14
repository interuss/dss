import datetime
import logging
import os
from typing import Dict

import yaml


logging.basicConfig()
_logger = logging.getLogger('tracer.logging')
_logger.setLevel(logging.DEBUG)


class Logger(object):
  def __init__(self, log_path: str):
    self.log_path = log_path
    _logger.info('Log path: {}'.format(self.log_path))
    os.makedirs(self.log_path, exist_ok=True)

  def log_same(self, t0: datetime.datetime, t1: datetime.datetime, code: str) -> None:
    with open(os.path.join(self.log_path, '000000_nochange_queries.yaml'), 'a') as f:
      body = {
        't0': t0.isoformat(),
        't1': t1.isoformat(),
        'code': code
      }
      f.write(yaml.dump(body, explicit_start=True))

  def log_new(self, code: str, content: Dict) -> str:
    n = len(os.listdir(self.log_path))
    logname = '{:06d}_{}_{}.yaml'.format(n, datetime.datetime.now().strftime('%H%M%S_%f'), code)
    fullname = os.path.join(self.log_path, logname)

    with open(fullname, 'w') as f:
      f.write(yaml.dump(content, indent=2))

    return logname
