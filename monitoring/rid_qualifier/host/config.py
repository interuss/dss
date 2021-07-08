import os

from monitoring.monitorlib import auth_validation


ENV_KEY_PREFIX = 'MOCK_HOST'
ENV_KEY_PUBLIC_KEY = '{}_PUBLIC_KEY'.format(ENV_KEY_PREFIX)
ENV_KEY_TOKEN_AUDIENCE = '{}_TOKEN_AUDIENCE'.format(ENV_KEY_PREFIX)
ENV_KEY_BASE_URL = '{}_BASE_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)
ENV_KEY_DSS = '{}_DSS_URL'.format(ENV_KEY_PREFIX)

# These keys map to entries in the Config class
KEY_TOKEN_PUBLIC_KEY = 'TOKEN_PUBLIC_KEY'
KEY_TOKEN_AUDIENCE = 'TOKEN_AUDIENCE'
KEY_BASE_URL = 'USS_BASE_URL'
KEY_AUTH_SPEC = 'AUTH_SPEC'
KEY_DSS_URL = 'DSS_URL'


workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')


class Config(object):
  TOKEN_PUBLIC_KEY = auth_validation.fix_key(
    os.environ.get(ENV_KEY_PUBLIC_KEY, '')).encode('utf-8')
  TOKEN_AUDIENCE = os.environ.get(ENV_KEY_TOKEN_AUDIENCE, '')
  USS_BASE_URL = os.environ[ENV_KEY_BASE_URL]
  AUTH_SPEC = os.environ[ENV_KEY_AUTH]
  DSS_URL = os.environ[ENV_KEY_DSS]
