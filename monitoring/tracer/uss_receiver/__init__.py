import logging
_logger = logging.getLogger('tracer.context')
_logger.setLevel(logging.DEBUG)
_logger.info('Proto message')

import flask

webapp = flask.Flask(__name__)

from . import routes
