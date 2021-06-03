from monitoring.monitorlib import auth, infrastructure
from monitoring.mock_ridsp import webapp
from . import config


dss_client = infrastructure.DSSTestSession(
  webapp.config.get(config.KEY_DSS_URL),
  auth.make_auth_adapter(webapp.config.get(config.KEY_AUTH_SPEC)))
