import os
import sys
import flask
from loguru import logger
from config import Config

webapp = flask.Flask(__name__)


@webapp.route("/mock/scd/.well-known/uas-traffic-management/mock_pub.der", methods=["GET"])
def get_public_key():
    """Implements notifyOperationalIntentDetailsChanged in ASTM SCD API."""
    public_key_file_location = Config.PUBLIC_KEY_PATH

    logger.info("Retreiving public key file from {}".format(public_key_file_location))

    return flask.send_file(public_key_file_location)

if __name__ == "__main__":
    port = int(os.environ.get('PORT', 8077))
    webapp.run(debug=True, host='0.0.0.0', port=port)
