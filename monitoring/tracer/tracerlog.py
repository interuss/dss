import copy
import datetime
import logging
import os
from typing import Dict

import yaml

from monitoring.monitorlib import infrastructure


logging.basicConfig()
_logger = logging.getLogger('tracer.logging')
_logger.setLevel(logging.DEBUG)


class Logger(object):
  def __init__(self, log_path: str, kml_session: infrastructure.KMLGenerationSession = None):
    self.log_path = log_path
    _logger.info('Log path: {}'.format(self.log_path))
    os.makedirs(self.log_path, exist_ok=True)
    self.kml_session = kml_session

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
    basename = '{:06d}_{}_{}'.format(n, datetime.datetime.now().strftime('%H%M%S_%f'), code)
    logname = '{}.yaml'.format(basename)
    fullname = os.path.join(self.log_path, logname)

    dump = copy.deepcopy(content)
    dump['object_type'] = type(content).__name__
    with open(fullname, 'w') as f:
      f.write(yaml.dump(dump, indent=2))

    if self.kml_session:
      kml_server_filename = os.path.join(self.kml_session.kml_folder, logname)
      try:
        with open(fullname, 'r') as f:
          resp = self.kml_session.post('/realtime_kml',
                                       data={'path': self.kml_session.kml_folder},
                                       files=[('files[]', f)])
        resp.raise_for_status()
        kml_path = os.path.join(self.log_path, 'kml')
        os.makedirs(kml_path, exist_ok=True)
        with open(os.path.join(kml_path, '{}.kml'.format(basename)), 'w') as f:
          f.write(resp.content.decode('utf-8'))
      except IOError as e:
        print('Error posting {} to KML server: {}'.format(kml_server_filename, e))

    return logname
