import flask

from monitoring.mock_uss import webapp, config
from monitoring.mock_uss.msgsigning.database import db
from loguru import logger
import os


@webapp.route(
    "/mock/msgsigning/.well-known/uas-traffic-management/pub.der", methods=["GET"]
)
def get_public_key():
    public_key_file_location = os.path.join(
        webapp.config.get(config.KEY_CERT_BASE_PATH), db.value.public_key_name
    )

    logger.info("Retrieving public key file from {}".format(public_key_file_location))

    return flask.send_file(public_key_file_location)


@webapp.route("/mock/msgsigning/set-keypair", methods=["POST"])
def set_keypair():
    # TODO: Update keypair to use via a received JSON object -> {"public_key": "/path", "private_key": "/path"}
    return flask.jsonify({"message": "Not yet implemented"}), 501


@webapp.route("/mock/msgsigning/report", methods=["GET"])
def get_report():
    # TODO: Return message signing report for uss_qualifier to access
    return flask.jsonify({"message": "Not yet implemented"}), 501
