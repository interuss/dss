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
    session = datetime.datetime.now().strftime('%Y%m%d_%H%M%S')
    self.log_path = os.path.join(log_path, session)
    _logger.info('Log path: {}'.format(self.log_path))
    os.makedirs(self.log_path, exist_ok=True)
    self.index = 0

  def logconfig(self, config: Dict) -> None:
    with open(os.path.join(self.log_path, 'config.yaml'), 'w') as f:
      f.write(yaml.dump(config, indent=2))

  def log_same(self, t0: datetime.datetime, t1: datetime.datetime, code: str) -> None:
    with open(os.path.join(self.log_path, 'nochange_queries.yaml'), 'a') as f:
      body = {
        't0': t0.isoformat(),
        't1': t1.isoformat(),
        'code': code
      }
      f.write(yaml.dump(body, explicit_start=True))

  def log_new(self, code: str, content: Dict) -> str:
    logname = '{}_{:03}_{}.yaml'.format(datetime.datetime.now().strftime('%H%M%S'), self.index % 1000, code)
    fullname = os.path.join(self.log_path, logname)
    self.index += 1

    with open(fullname, 'w') as f:
      f.write(yaml.dump(content, indent=2))

    return logname
