import flask

from monitoring.mock_uss import webapp, config
from loguru import logger
import os


@webapp.route(
    "/mock/msgsigning/.well-known/uas-traffic-management/pub.der", methods=["GET"]
)
def get_public_key():
    public_key_file_location = os.path.join(
        webapp.config.get(config.KEY_CERT_BASE_PATH), "messagesigning/mock_faa_pub.der"
    )

    logger.info("Retreiving public key file from {}".format(public_key_file_location))

    return flask.send_file(public_key_file_location)
