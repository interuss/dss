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
        query = {'request': signature_info}
        #query['reponse'] = response.json()
        test_context = {
            'test_name': 'Message signing in outgoing {} Op req from UUT',
            'test_case': 'Uss {} Op with Message Signing Expect non 403 response'}
        interaction_id = report_settings.reprt_recorder.capture_interaction(query,
                                                            "Checking message signing for {} response to {}".format(associated_request.method, associated_request.url),
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