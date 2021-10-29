import flask

webapp = flask.Flask(__name__)

from . import routes
