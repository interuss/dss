import traceback
import flask
from werkzeug.exceptions import HTTPException

from monitoring.monitorlib import auth_validation, versioning
from monitoring.mock_uss import webapp, enabled_services


@webapp.route("/status")
def status():
    return "Mock USS ok {}; hosting {}".format(
        versioning.get_code_version(), ", ".join(enabled_services)
    )


@webapp.errorhandler(Exception)
def handle_exception(e):
    if isinstance(e, HTTPException):
        return e
    elif isinstance(e, auth_validation.InvalidScopeError):
        return (
            flask.jsonify(
                {
                    "message": "Invalid scope; expected one of {%s}, but received only {%s}"
                    % (" ".join(e.permitted_scopes), " ".join(e.provided_scopes))
                }
            ),
            403,
        )
    elif isinstance(e, auth_validation.InvalidAccessTokenError):
        return flask.jsonify({"message": e.message}), 401
    elif isinstance(e, auth_validation.ConfigurationError):
        return (
            flask.jsonify(
                {"message": "Auth validation configuration error: " + e.message}
            ),
            500,
        )
    elif isinstance(e, ValueError):
        traceback.print_exc()
        return flask.jsonify({"message": str(e)}), 400
    traceback.print_exc()
    return (
        flask.jsonify({"message": "Unhandled {}: {}".format(type(e).__name__, str(e))}),
        500,
    )
