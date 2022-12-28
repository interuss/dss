from monitoring.mock_uss import config, SERVICE_ATPROXY_CLIENT

if not config.Config.ATPROXY_BASE_URL:
    raise ValueError(
        f"{config.ENV_KEY_ATPROXY_BASE_URL} is required for the {SERVICE_ATPROXY_CLIENT} service"
    )

if not config.Config.ATPROXY_BASIC_AUTH:
    raise ValueError(
        f"{config.ENV_KEY_ATPROXY_BASIC_AUTH} is required for the {SERVICE_ATPROXY_CLIENT} service"
    )
