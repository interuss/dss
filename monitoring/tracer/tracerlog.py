import os


class Logger(object):
  def __init__(self, log_path: str):
    os.makedirs(log_path, exist_ok=True)
    self._log_path = log_path
