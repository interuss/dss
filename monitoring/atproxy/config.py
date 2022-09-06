import os
from typing import Dict

from werkzeug.security import generate_password_hash

from monitoring.monitorlib import auth_validation

ENV_KEY_PREFIX = 'ATPROXY'
ENV_KEY_PUBLIC_KEY = '{}_PUBLIC_KEY'.format(ENV_KEY_PREFIX)
ENV_KEY_TOKEN_AUDIENCE = '{}_TOKEN_AUDIENCE'.format(ENV_KEY_PREFIX)
ENV_KEY_CLIENT_BASIC_AUTH = '{}_CLIENT_BASIC_AUTH'.format(ENV_KEY_PREFIX)

# These keys map to entries in the Config class
KEY_TOKEN_PUBLIC_KEY = 'TOKEN_PUBLIC_KEY'
KEY_TOKEN_AUDIENCE = 'TOKEN_AUDIENCE'
KEY_CLIENT_BASIC_AUTH = 'CLIENT_BASIC_AUTH'

KEY_CODE_VERSION = 'MONITORING_VERSION'

workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'workspace')


class Config(object):
    TOKEN_PUBLIC_KEY = auth_validation.fix_key(
        os.environ.get(ENV_KEY_PUBLIC_KEY, '')).encode('utf-8')
    TOKEN_AUDIENCE = os.environ.get(ENV_KEY_TOKEN_AUDIENCE, '')
    CLIENT_BASIC_AUTH = os.environ[ENV_KEY_CLIENT_BASIC_AUTH]
    CODE_VERSION = os.environ.get(KEY_CODE_VERSION, 'Unknown')


def get_users(basic_auth: str) -> Dict[str, str]:
    user_pass = [v.strip() for v in basic_auth.split(':')]
    if len(user_pass) != 2:
        raise ValueError('Expected "username:password", got "{}"'.format(basic_auth))
    return {user_pass[0]: generate_password_hash(user_pass[1])}
