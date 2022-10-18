from monitoring.mock_uss import config, SERVICE_RIDSP

if config.Config.DSS_URL is None:
    raise ValueError(f"DSS_URL is required for the {SERVICE_RIDSP} service")

if config.Config.AUTH_SPEC is None:
    raise ValueError(f"AUTH_SPEC is required for the {SERVICE_RIDSP} service")
