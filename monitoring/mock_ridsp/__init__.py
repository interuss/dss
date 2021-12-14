import flask

from .config import Config

webapp = flask.Flask(__name__)

webapp.config.from_object(Config)
print(
  '################################################################################\n' + \
  '################################ Configuration  ################################\n' + \
  '\n'.join('## {}: {}'.format(key, webapp.config[key]) for key in webapp.config) + '\n' + \
  '################################################################################', flush=True)

from monitoring.mock_ridsp import routes
