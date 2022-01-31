import flask

from monitoring.mock_uss import config

SERVICE_RIDSP = 'ridsp'

webapp = flask.Flask(__name__)

webapp.config.from_object(config.Config)
print(
  '################################################################################\n' + \
  '################################ Configuration  ################################\n' + \
  '\n'.join('## {}: {}'.format(key, webapp.config[key]) for key in webapp.config) + '\n' + \
  '################################################################################', flush=True)

from monitoring.mock_uss import routes as basic_routes

if SERVICE_RIDSP in webapp.config[config.KEY_SERVICES]:
    from monitoring.mock_uss import ridsp
    from monitoring.mock_uss.ridsp import routes as ridsp_routes
