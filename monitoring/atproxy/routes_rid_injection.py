import logging
from typing import Tuple

import flask

from . import webapp, handling
from .oauth import requires_scope
from .requests import RIDInjectionCreateTestRequest, RIDInjectionDeleteTestRequest
from monitoring.monitorlib.rid_automated_testing import injection_api
from implicitdict import ImplicitDict


logging.basicConfig()
_logger = logging.getLogger('atproxy.rid_injection')
_logger.setLevel(logging.DEBUG)


@webapp.route('/ridsp/injection/tests/<test_id>', methods=['PUT'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def rid_injection_create_test(test_id: str) -> Tuple[str, int]:
    """Implements test creation in RID automated testing injection API."""
    try:
        json = flask.request.json
        if json is None:
            raise ValueError('Request did not contain a JSON payload')
        req_body: injection_api.CreateTestParameters = ImplicitDict.parse(json, injection_api.CreateTestParameters)
    except ValueError as e:
        msg = 'Create test {} unable to parse JSON: {}'.format(test_id, e)
        return msg, 400

    return handling.fulfill_query(RIDInjectionCreateTestRequest(test_id=test_id, request_body=req_body), _logger)


@webapp.route('/ridsp/injection/tests/<test_id>/<version>', methods=['DELETE'])
@requires_scope([injection_api.SCOPE_RID_QUALIFIER_INJECT])
def rid_injection_delete_test(test_id: str, version: str) -> Tuple[str, int]:
    """Implements test deletion in RID automated testing injection API."""
    return handling.fulfill_query(RIDInjectionDeleteTestRequest(test_id=test_id, version=version), _logger)
