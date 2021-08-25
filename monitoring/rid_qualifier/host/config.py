import os
import rq

import redis
from redis import Redis


# These keys map to entries in the Config class
ENV_KEY_PREFIX = 'MOCK_HOST_RID_QUALIFIER'
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)
ENV_KEY_REDIS_URL = '{}_REDIS_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_RID_QUALIFIER_HOST_URL = '{}_HOST_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_RID_QUALIFIER_HOST_PORT = '{}_HOST_PORT'.format(ENV_KEY_PREFIX)


workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')
os.environ["OAUTHLIB_INSECURE_TRANSPORT"] = "1"


class Config(object):
  AUTH_SPEC = os.environ[ENV_KEY_AUTH]
  REDIS_URL = os.environ[ENV_KEY_REDIS_URL]
  RID_QUALIFIER_HOST_URL = os.environ[ENV_KEY_RID_QUALIFIER_HOST_URL]
  RID_QUALIFIER_HOST_PORT = os.environ[ENV_KEY_RID_QUALIFIER_HOST_PORT]
  REDIS_QUEUE = 'qualifer-tasks'
  SECRET_KEY = os.environ.get('SECRET_KEY') or 'a-test-secret-string'
  qualifier_queue = rq.Queue(
    REDIS_QUEUE, connection=Redis.from_url(REDIS_URL), default_timeout=3600)
  POOL = redis.ConnectionPool(host='redis', port=6379, db=0)
  redis_client = redis.StrictRedis(connection_pool=POOL)
  INPUT_PATH = '/mnt/app/input-files'
  FILE_PATH = '/app/rid-host-files'
