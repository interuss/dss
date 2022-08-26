from datetime import datetime, timedelta
import logging
import time
from typing import Tuple
import uuid

import flask

from monitoring.monitorlib.typing import ImplicitDict
from .database import db, Query, QueryState


def fulfill_query(req: ImplicitDict, logger: logging.Logger) -> Tuple[str, int]:
    t_start = datetime.utcnow()
    query = Query(type=req.request_type_name(), request=req)
    timeout = timedelta(seconds=59)
    id = str(uuid.uuid4())
    with db as tx:
        tx.queries[id] = query
        logger.debug('Added {} query {} to handler queue'.format(query.type, id))
    while datetime.utcnow() < t_start + timeout:
        time.sleep(0.1)
        with db as tx:
            if tx.queries[id].state == QueryState.Complete:
                logger.debug('Fulfilling {} query {}'.format(query.type, id))
                query = tx.queries.pop(id)
                return flask.jsonify(query.response), query.return_code

    # Time expired; remove request from queue and indicate error
    with db as tx:
        tx.queries.pop(id)
    logger.debug('Failed to fulfill {} query {} in time (backend handler did not provide a response)'.format(query.type, id))
    return flask.jsonify({'message': 'Backend handler did not respond within the alotted time'}), 500
