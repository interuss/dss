import flask

from .config import Config

webapp = flask.Flask(__name__)

webapp.config.from_object(Config)

from monitoring.mock_riddp import routes
