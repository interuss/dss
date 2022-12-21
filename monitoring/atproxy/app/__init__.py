import logging
_logger = logging.getLogger('atproxy.context')
_logger.setLevel(logging.DEBUG)

import flask
from flask_httpauth import HTTPBasicAuth

from monitoring.atproxy import config

webapp = flask.Flask(__name__)
basic_auth = HTTPBasicAuth()

webapp.config.from_object(config.Config)
users = config.get_users(webapp.config[config.KEY_CLIENT_BASIC_AUTH])

from monitoring.atproxy import routes
