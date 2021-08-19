import flask
from werkzeug.exceptions import HTTPException

from monitoring.monitorlib import versioning, auth_validation
from monitoring.mock_riddp import webapp


@webapp.route('/status')
def status():
  return 'Mock RID Display Provider ok {}'.format(versioning.get_code_version())


@webapp.errorhandler(Exception)
def handle_exception(e):
  if isinstance(e, HTTPException):
    return e
  elif isinstance(e, auth_validation.InvalidScopeError):
    return flask.jsonify({
      'message': 'Invalid scope; expected one of {%s}, but received only {%s}' % (' '.join(e.permitted_scopes),
                                                                                  ' '.join(e.provided_scopes))}), 403
  elif isinstance(e, auth_validation.InvalidAccessTokenError):
    return flask.jsonify({'message': e.message}), 401
  elif isinstance(e, auth_validation.ConfigurationError):
    return flask.jsonify({'message': e.message}), 500
  elif isinstance(e, ValueError):
    return flask.jsonify({'message': str(e)}), 400

  return flask.jsonify({'message': str(e)}), 500


from . import routes_observation
