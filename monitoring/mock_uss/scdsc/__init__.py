from monitoring.mock_uss import config, SERVICE_SCDSC

if config.Config.DSS_URL is None:
    raise ValueError(f"DSS_URL is required for the {SERVICE_SCDSC} service")

if config.Config.AUTH_SPEC is None:
    raise ValueError(f"AUTH_SPEC is required for the {SERVICE_SCDSC} service")
