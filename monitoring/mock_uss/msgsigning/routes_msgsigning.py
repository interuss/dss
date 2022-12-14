import flask

from monitoring.mock_uss import webapp
from monitoring.mock_uss.config import Config
from loguru import logger


@webapp.route(
    "/mock/msgsigning/.well-known/uas-traffic-management/pub.der", methods=["GET"]
)
def get_public_key():
    public_key_file_location = Config.PUBLIC_KEY_PATH

    logger.info("Retreiving public key file from {}".format(public_key_file_location))

    return flask.send_file(public_key_file_location)
