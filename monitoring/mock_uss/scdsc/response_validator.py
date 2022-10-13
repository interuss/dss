from loguru import logger
import json
import datetime
from monitoring.mock_uss.scdsc import report_settings
from  monitoring.messagesigning.message_validator import MessageValidatorService
message_validator = MessageValidatorService()

def validate_response(response):
    try:
        logger.info("Validating response...")
        associated_request = response.request
        response_body = get_response_body(response)
        signature_info = {
            'method': associated_request.method,
            'url': associated_request.url,
            'initiated_at': datetime.datetime.utcnow().isoformat(),
            'headers': json.dumps({k: v for k, v in response.headers.items()}),
            "status": response.status_code
        }
        signature_info['body'] = json.dumps(response_body) if response_body else ''
        query = {
           'response':  {
               'code': response.status_code,
               'headers': response.headers,
               'json': None if not response_body else response_body
           } 
        }
        test_description = 'Message signing headers in the response from the outgoing {} request to {} should be valid.'.format(
            associated_request.method, associated_request.url
        )
        test_context = {
            'test_name': 'Validating response signatures.',
            'test_case': test_description}
        interaction_id = report_settings.reprt_recorder.capture_interaction(query,
                                                            test_description,
                                                            test_context=test_context)
        message_validator.analyze_headers(interaction_id, signature_info, 'response')

    except Exception as e:
     logger.error("Error validating response: " + str(e))

def get_response_body(response):
    try:
        data = response.json()
        logger.info("The RESPONSE BODY STRING ALL CHARS IS: \n")
        print(repr(json.dumps(data)))
        return data
    except Exception:
        return ''