import os


# These keys map to entries in the Config class
ENV_KEY_PREFIX = 'MOCK_HOST_USS_QUALIFIER'
ENV_KEY_AUTH = '{}_AUTH_SPEC'.format(ENV_KEY_PREFIX)
ENV_KEY_REDIS_URL = '{}_REDIS_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_USS_QUALIFIER_HOST_URL = '{}_HOST_URL'.format(ENV_KEY_PREFIX)
ENV_KEY_USS_QUALIFIER_HOST_PORT = '{}_HOST_PORT'.format(ENV_KEY_PREFIX)

workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')

KEY_REDIS_QUEUE = 'REDIS_QUEUE'
KEY_REDIS_URL = 'REDIS_URL'
KEY_USS_QUALIFIER_HOST_URL = 'USS_QUALIFIER_HOST_URL'
KEY_FILE_PATH = 'FILE_PATH'
KEY_USS_QUALIFIER_HOST_PORT = 'USS_QUALIFIER_HOST_PORT'

class Config(object):
    AUTH_SPEC = os.environ[ENV_KEY_AUTH]
    SECRET_KEY = os.environ.get('SECRET_KEY') or 'a-test-secret-string'
    FILE_PATH = '/app/uss-host-files'
    os.environ["OAUTHLIB_INSECURE_TRANSPORT"] = "1"

    REDIS_QUEUE = 'qualifer-tasks'
    USS_QUALIFIER_HOST_PORT = os.environ[ENV_KEY_USS_QUALIFIER_HOST_PORT]
    REDIS_URL = os.environ[ENV_KEY_REDIS_URL]
    USS_QUALIFIER_HOST_URL = os.environ[ENV_KEY_USS_QUALIFIER_HOST_URL]
