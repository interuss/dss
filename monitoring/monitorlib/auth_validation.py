import json
from typing import List, NamedTuple
from functools import wraps
import flask
import jwcrypto.jwk
import jwt
import requests
import datetime

from loguru import logger

from monitoring.mock_uss.scdsc import report_settings
from monitoring.messagesigning.message_validator import MessageValidatorService

message_validator = MessageValidatorService()

class Authorization(NamedTuple):
    client_id: str
    scopes: List[str]
    issuer: str


class InvalidScopeError(Exception):
    def __init__(self, permitted_scopes, provided_scopes):
        self.permitted_scopes = permitted_scopes
        self.provided_scopes = provided_scopes


class InvalidAccessTokenError(Exception):
    def __init__(self, message):
        self.message = message


class ConfigurationError(Exception):
    def __init__(self, message):
        self.message = message


def analyze_message_signing_headers():
  request_info = {
        'method': flask.request.method,
        'url': flask.request.url,
        'initiated_at': datetime.datetime.utcnow().isoformat(),
        'headers': json.dumps({k: v for k, v in flask.request.headers.items()})
    }

  request_info['body'] = flask.request.data.decode('utf-8')
  query = {'request': request_info}
  test_context = {
            'test_name': 'Checking for existence of message signing headers',
            'test_case': 'Message signing headers in {} request to {} should exist.'.format(flask.request.method, flask.request.path)}
  interaction_id = report_settings.reprt_recorder.capture_interaction(
            query,
            "Checking message signing for {} {}".format(flask.request.method, flask.request.url),
            test_context=test_context)
  message_validator.analyze_headers(interaction_id, request_info, 'request')


def requires_scope_decorator(public_key: str, audience: str):
    """Function that produces a decorator to protect a Flask endpoint.

    If you decorate an endpoint with a decorator produced by this function, it
    will ensure that the requester has a valid access token with the required
    scope before allowing the endpoint to be called.
    """
    audiences = audience.split(",") if audience else []

    def decorator(permitted_scopes):
        def outer_wrapper(fn):
            @wraps(fn)
            def wrapper(*args, **kwargs):
                try:
                    if '/mock/scd/' in flask.request.path:
                        analyze_message_signing_headers()
                except Exception as e:
                    logger.error("Could not process message signing headers: {}".format(str(e)))
                if hasattr(flask.request, "jwt"):
                    # Token has already been processed; check additional scope
                    has_scope = False
                    for scope in permitted_scopes:
                        if scope in flask.request.jwt.scopes:
                            has_scope = True
                            break
                    if not has_scope:
                        raise InvalidScopeError(
                            permitted_scopes, flask.request.jwt.scopes
                        )
                else:
                    # Token has not yet been processed; process it
                    token = flask.request.headers.get("Authorization", None)
                    if token is None:
                        raise InvalidAccessTokenError("Missing Authorization header")
                    token = token.replace("Bearer ", "")
                    try:
                        if not public_key:
                            raise ConfigurationError(
                                "Public key for access tokens is not configured on server"
                            )
                        if not audiences:
                            raise ConfigurationError(
                                "Audience for access tokens is not configured on server"
                            )
                        r = jwt.decode(
                            token,
                            public_key,
                            algorithms="RS256",
                            options={"verify_aud": False},
                        )
                        if "aud" not in r:
                            raise InvalidAccessTokenError(
                                "Access token is missing aud claim."
                            )
                        if r["aud"] not in audiences:
                            raise InvalidAccessTokenError(
                                'Access token audience "{}" is invalid; expected {}'.format(
                                    r["aud"], ", ".join(audiences)
                                )
                            )
                        provided_scopes = r["scope"].split(" ")
                        has_scope = False
                        for scope in permitted_scopes:
                            if scope in provided_scopes:
                                has_scope = True
                                break
                        if not has_scope:
                            raise InvalidScopeError(permitted_scopes, provided_scopes)
                        client_id = (
                            r["client_id"] if "client_id" in r else r.get("sub", None)
                        )
                    except jwt.ImmatureSignatureError:
                        raise InvalidAccessTokenError("Access token is immature.")
                    except jwt.ExpiredSignatureError:
                        raise InvalidAccessTokenError("Access token has expired.")
                    except jwt.InvalidSignatureError:
                        raise InvalidAccessTokenError(
                            "Access token signature is invalid."
                        )
                    except jwt.DecodeError:
                        raise InvalidAccessTokenError("Access token cannot be decoded.")
                    except jwt.InvalidTokenError as e:
                        raise InvalidAccessTokenError(
                            "Unexpected InvalidTokenError: %s" % str(e)
                        )
                    issuer = r.get("iss", None)
                    flask.request.jwt = Authorization(
                        client_id, provided_scopes, issuer
                    )

                return fn(*args, **kwargs)

            return wrapper

        return outer_wrapper

    return decorator


def fix_key(public_key: str) -> str:
    """Convert a user-specified public key into a properly-formatted PEM string"""

    if public_key.startswith("http://") or public_key.startswith("https://"):
        resp = requests.get(public_key)
        if public_key.endswith(".json"):
            key = resp.json()
            if "keys" in key:
                key = key["keys"][0]
            jwk = jwcrypto.jwk.JWK.from_json(json.dumps(key))
            public_key = jwk.export_to_pem().decode("utf-8")
        else:
            public_key = resp.content.decode("utf-8")
    elif public_key.startswith("/") or public_key.endswith((".pem")):
        with open(public_key, "r") as f:
            public_key = f.read()
    # ENV variables sometimes don't pass newlines, spec says white space
    # doesn't matter, but pyjwt cares about it, so fix it
    public_key = public_key.replace(" PUBLIC ", "_PLACEHOLDER_")
    public_key = public_key.replace(" ", "\n")
    public_key = public_key.replace("_PLACEHOLDER_", " PUBLIC ")
    return public_key
