import os

from monitoring.monitorlib import auth_validation


ENV_KEY_PREFIX = 'MOCK_HOST'
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)

# These keys map to entries in the Config class
KEY_AUTH_SPEC = 'AUTH_SPEC'


workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')


class Config(object):
  AUTH_SPEC = os.environ[ENV_KEY_AUTH]
  REDIS_URL = os.environ['REDIS_URL']
  REDIS_QUEUE = 'qualifer-tasks'
