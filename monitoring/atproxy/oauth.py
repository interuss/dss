from monitoring.monitorlib import auth_validation
from . import config, webapp


requires_scope = auth_validation.requires_scope_decorator(
    webapp.config.get(config.KEY_TOKEN_PUBLIC_KEY),
    webapp.config.get(config.KEY_TOKEN_AUDIENCE))
