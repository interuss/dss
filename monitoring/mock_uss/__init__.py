import flask

from monitoring.mock_uss import config

SERVICE_RIDSP = "ridsp"
SERVICE_RIDDP = "riddp"
SERVICE_SCDSC = "scdsc"

webapp = flask.Flask(__name__)
enabled_services = set()

webapp.config.from_object(config.Config)
print(
    "################################################################################\n"
    + "################################ Configuration  ################################\n"
    + "\n".join("## {}: {}".format(key, webapp.config[key]) for key in webapp.config)
    + "\n"
    + "################################################################################",
    flush=True,
)

from monitoring.mock_uss import routes as basic_routes

if SERVICE_RIDSP in webapp.config[config.KEY_SERVICES]:
    enabled_services.add(SERVICE_RIDSP)
    from monitoring.mock_uss import ridsp
    from monitoring.mock_uss.ridsp import routes as ridsp_routes

if SERVICE_RIDDP in webapp.config[config.KEY_SERVICES]:
    enabled_services.add(SERVICE_RIDDP)
    from monitoring.mock_uss import riddp
    from monitoring.mock_uss.riddp import routes as riddp_routes

if SERVICE_SCDSC in webapp.config[config.KEY_SERVICES]:
    enabled_services.add(SERVICE_SCDSC)
    from monitoring.mock_uss import scdsc
    from monitoring.mock_uss.scdsc import routes as scdsc_routes
