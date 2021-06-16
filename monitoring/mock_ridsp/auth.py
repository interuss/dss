from monitoring.monitorlib import auth_validation
from monitoring.mock_ridsp import webapp
from . import config


requires_scope = auth_validation.requires_scope_decorator(
  webapp.config.get(config.KEY_TOKEN_PUBLIC_KEY),
  webapp.config.get(config.KEY_TOKEN_AUDIENCE))
