import datetime

import flask

from monitoring.monitorlib import rid
from monitoring.mock_ridsp import webapp
from monitoring.mock_ridsp.auth import requires_scope


@webapp.route('/v1/uss/identification_service_areas/<id>', methods=['POST'])
@requires_scope([rid.SCOPE_WRITE])
def notify_isa(id: str):
  return 'Notification acknowledged', 204


@webapp.route('/v1/uss/flights', methods=['GET'])
@requires_scope([rid.SCOPE_READ])
def flights():
  return flask.jsonify({
    'timestamp': datetime.datetime.utcnow().isoformat() + 'Z',
    'flights': []
  })


@webapp.route('/v1/uss/flights/<id>/details', methods=['GET'])
@requires_scope([rid.SCOPE_READ])
def flight_details(id: str):
  return flask.jsonify({
    'message': 'Flight {} not found'.format(id),
  }), 404
