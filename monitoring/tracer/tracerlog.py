import datetime
import os
from typing import Dict, Optional

import yaml


suffixes = ['']
for c in 'bcdefghijklmnopqrstuvwxyz':
  suffixes.append(c)
for c1 in 'abcdefghijklmnopqrstuvwxyz':
  for c2 in 'abcdefghijklmnopqrstuvwxyz':
    suffixes.append(c1 + c2)


class Logger(object):
  def __init__(self, log_path: str):
    session = datetime.datetime.now().strftime('%Y%m%d_%H%M%S')
    self._log_path = os.path.join(log_path, session)
    os.makedirs(self._log_path, exist_ok=True)

  def logconfig(self, config: Dict) -> None:
    with open(os.path.join(self._log_path, 'config.yaml'), 'w') as f:
      f.write(yaml.dump(config, indent=2))

  def log(self, timestamp: datetime.datetime, code: str, content: Optional[Dict]) -> None:
    if not content:
      with open(os.path.join(self._log_path, 'nochange_queries.yaml'), 'a') as f:
        body = {
          't': timestamp.isoformat(),
          'code': code
        }
        f.write(yaml.dump(body, explicit_start=True))
    else:
      for suffix in suffixes:
        logname = '{}{}_{}.yaml'.format(datetime.datetime.now().strftime('%H%M%S'), suffix, code)
        fullname = os.path.join(self._log_path, logname)
        if not os.path.exists(fullname):
          break

      with open(fullname, 'w') as f:
        f.write(yaml.dump(content, indent=2))
