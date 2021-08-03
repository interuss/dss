import os
import rq

import redis
from redis import Redis


ENV_KEY_PREFIX = 'MOCK_HOST'
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)

# These keys map to entries in the Config class
KEY_AUTH_SPEC = 'AUTH_SPEC'


workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')


class Config(object):
  AUTH_SPEC = os.environ[ENV_KEY_AUTH]
  REDIS_URL = os.environ['REDIS_URL']
  REDIS_QUEUE = 'qualifer-tasks'
  SECRET_KEY = os.environ.get('SECRET_KEY') or 'a-test-secret-string'
  qualifier_queue = rq.Queue(REDIS_QUEUE, connection=Redis.from_url(REDIS_URL), default_timeout=3600)
  POOL = redis.ConnectionPool(host='redis', port=6379, db=0)
  redis_client = redis.StrictRedis(connection_pool=POOL)
  INPUT_PATH = '/mnt/app/input-files'
