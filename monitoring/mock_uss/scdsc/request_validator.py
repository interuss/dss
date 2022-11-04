import flask
import datetime
import json
from monitoring.messagesigning.message_validator import MessageValidatorService

message_validator = MessageValidatorService()

def validate_message_signing_headers():
    request_info = {
        'method': flask.request.method,
        'url': flask.request.url,
        'initiated_at': datetime.datetime.utcnow().isoformat(),
        'headers': json.dumps({k: v for k, v in flask.request.headers.items()})
    }
    request_info['body'] = flask.request.data.decode('utf-8')
    query = {'request': request_info}
    test_context = {
        'test_name': "Validating incoming request signatures.",
        'test_case': 'Message signing headers in the incoming {} request to this mock_uss endpoint:  {} should be valid.'.format(flask.request.method, flask.request.path)
        }

    message_validator.analyze_headers(None, request_info, 'request')
    validation_metadata = {
        'results': message_validator.results,
        'request_info': request_info,
        'query': query,
        'test_context' : test_context
    }
    return validation_metadata