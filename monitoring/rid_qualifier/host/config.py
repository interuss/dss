import os


# These keys map to entries in the Config class
ENV_KEY_PREFIX = 'MOCK_HOST_RID_QUALIFIER'
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)
ENV_KEY_REDIS_URL = '{}_REDIS_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_RID_QUALIFIER_HOST_URL = '{}_HOST_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_RID_QUALIFIER_HOST_PORT = '{}_HOST_PORT'.format(ENV_KEY_PREFIX)

workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')

KEY_REDIS_QUEUE = 'REDIS_URL'
KEY_REDIS_URL = 'REDIS_URL'
KEY_RID_QUALIFIER_HOST_URL = 'RID_QUALIFIER_HOST_URL'
KEY_FILE_PATH = 'FILE_PATH'

class Config(object):
    AUTH_SPEC = os.environ[ENV_KEY_AUTH]
    SECRET_KEY = os.environ.get('SECRET_KEY') or 'a-test-secret-string'
    FILE_PATH = '/app/rid-host-files'
    os.environ["OAUTHLIB_INSECURE_TRANSPORT"] = "1"

    REDIS_QUEUE = 'qualifer-tasks'
    RID_QUALIFIER_HOST_PORT = os.environ[ENV_KEY_RID_QUALIFIER_HOST_PORT]
    REDIS_URL = os.environ[ENV_KEY_REDIS_URL]
    RID_QUALIFIER_HOST_URL = os.environ[ENV_KEY_RID_QUALIFIER_HOST_URL]
