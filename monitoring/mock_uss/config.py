from enum import Enum
import os

from monitoring.monitorlib import auth_validation
from monitoring.monitorlib.locality import Locality

ENV_KEY_PREFIX = "MOCK_USS"
ENV_KEY_PUBLIC_KEY = "{}_PUBLIC_KEY".format(ENV_KEY_PREFIX)
ENV_KEY_TOKEN_AUDIENCE = "{}_TOKEN_AUDIENCE".format(ENV_KEY_PREFIX)
ENV_KEY_BASE_URL = "{}_BASE_URL".format(ENV_KEY_PREFIX)
ENV_KEY_AUTH = "{}_AUTH_SPEC".format(ENV_KEY_PREFIX)
ENV_KEY_SERVICES = "{}_SERVICES".format(ENV_KEY_PREFIX)
ENV_KEY_DSS = "{}_DSS_URL".format(ENV_KEY_PREFIX)
ENV_KEY_BEHAVIOR_LOCALITY = "{}_BEHAVIOR_LOCALITY".format(ENV_KEY_PREFIX)

# These keys map to entries in the Config class
KEY_TOKEN_PUBLIC_KEY = "TOKEN_PUBLIC_KEY"
KEY_TOKEN_AUDIENCE = "TOKEN_AUDIENCE"
KEY_BASE_URL = "USS_BASE_URL"
KEY_AUTH_SPEC = "AUTH_SPEC"
KEY_SERVICES = "SERVICES"
KEY_DSS_URL = "DSS_URL"
KEY_BEHAVIOR_LOCALITY = "BEHAVIOR_LOCALITY"

KEY_CODE_VERSION = "MONITORING_VERSION"

workspace_path = os.path.join(os.path.abspath(os.path.dirname(__file__)), "workspace")


class Config(object):
    TOKEN_PUBLIC_KEY = auth_validation.fix_key(
        os.environ.get(ENV_KEY_PUBLIC_KEY, "")
    ).encode("utf-8")
    TOKEN_AUDIENCE = os.environ.get(ENV_KEY_TOKEN_AUDIENCE, "")
    USS_BASE_URL = os.environ.get(ENV_KEY_BASE_URL, None)
    AUTH_SPEC = os.environ.get(ENV_KEY_AUTH, None)
    SERVICES = set(
        svc.strip().lower() for svc in os.environ.get(ENV_KEY_SERVICES, "").split(",")
    )
    DSS_URL = os.environ.get(ENV_KEY_DSS, None)
    BEHAVIOR_LOCALITY = Locality(os.environ.get(ENV_KEY_BEHAVIOR_LOCALITY, "CHE"))
    CODE_VERSION = os.environ.get(KEY_CODE_VERSION, "Unknown")
    CERT_PATH = "/var/test-certs"
    PRIVATE_KEY_PATH = "{}/messagesigning/mock_faa_priv.pem".format(CERT_PATH)
    PUBLIC_KEY_PATH = "{}/messagesigning/mock_faa_pub.der".format(CERT_PATH)
