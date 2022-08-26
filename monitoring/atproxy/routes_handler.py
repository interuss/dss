from datetime import datetime, timedelta
import logging
import time
from typing import Tuple

import flask
from typing import List, Optional

from monitoring.monitorlib.typing import ImplicitDict
from . import webapp, basic_auth
from .database import db, Query, QueryState


logging.basicConfig()
_logger = logging.getLogger('atproxy.handler')
_logger.setLevel(logging.DEBUG)


class PendingRequest(ImplicitDict):
    id: str
    type: str
    request: dict


class ListQueriesResponse(ImplicitDict):
    requests: List[PendingRequest]


class PutQueryRequest(ImplicitDict):
    response: Optional[dict] = None
    return_code: int


@webapp.route('/handler/queries', methods=['GET'])
@basic_auth.login_required
def list_queries() -> Tuple[str, int]:
    """Lists outstanding queries to be handled"""
    t_start = datetime.utcnow()
    _logger.debug('Handler requesting queries')
    max_timeout = timedelta(seconds=5)
    while datetime.utcnow() < t_start + max_timeout:
        with db as tx:
            response = ListQueriesResponse(requests=[
                PendingRequest(id=id, type=q.type, request=q.request)
                for id, q in tx.queries.items()
                if q.state == QueryState.Queued])
            if response.requests:
                _logger.debug('Provided handler {} queries'.format(len(response.requests)))
                return flask.jsonify(response)
        time.sleep(0.1)
    _logger.debug('No queries available for handler')
    return flask.jsonify(ListQueriesResponse(requests=[]))


@webapp.route('/handler/queries/<id>', methods=['PUT'])
@basic_auth.login_required
def put_query_result(id: str) -> Tuple[str, int]:
    """"""
    try:
        request: PutQueryRequest = ImplicitDict.parse(flask.request.json, PutQueryRequest)
    except ValueError as e:
        return flask.jsonify({'message': 'Could not parse PutQueryRequest: {}'.format(e)}), 400
    with db as tx:
        if id not in tx.queries:
            return flask.jsonify({'message': 'No outstanding request with ID {} exists'.format(id)}), 400
        query: Query = tx.queries[id]
        _logger.debug('{} query {} handled with code {}'.format(query.type, id, request.return_code))
        query.return_code = request.return_code
        query.response = request.response
        query.state = QueryState.Complete
    return '', 204
