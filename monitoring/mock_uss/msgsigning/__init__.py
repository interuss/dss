from monitoring.mock_uss import config, SERVICE_MESSAGESIGNING

if config.Config.CERT_BASE_PATH is None:
    raise ValueError(
        f"CERT_BASE_PATH is required for the {SERVICE_MESSAGESIGNING} service"
    )
