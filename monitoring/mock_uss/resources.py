from monitoring.monitorlib import auth, infrastructure
from monitoring.mock_uss import webapp
from . import config


utm_client = infrastructure.UTMClientSession(
    webapp.config[config.KEY_DSS_URL],
    auth.make_auth_adapter(webapp.config[config.KEY_AUTH_SPEC]))
