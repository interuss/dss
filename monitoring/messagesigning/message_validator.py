from monitoring.mock_uss.scdsc import report_settings
from loguru import logger
import json
import base64
import hashlib
import requests
from Crypto.PublicKey import RSA
from Crypto.Signature import pkcs1_15
from Crypto.Hash import SHA256
import flask

class MessageValidatorService:
    def __init__(self):
        self.required_message_signing_headers = set(['x-utm-message-signature', 'x-utm-message-signature-input', 'content-digest', 'x-utm-jws-header'])
        self.headers = {}
        self.request_covered_components = ['@method', '@path', '@query', 'authorization', 'content-type', 'content-digest', 'x-utm-jws-header']
        self.response_covered_components = ["@status", "content-type", "content-digest", "x-utm-jws-header"]
        self.public_key = None

    def analyze_headers(self, interaction_id, signature_info, request_type):
        self.scan_for_missing_headers(interaction_id, signature_info, request_type)

    def scan_for_missing_headers(self, interaction_id, signature_info, request_type):
        self.headers = json.loads(signature_info['headers'])
        self.headers = {key.lower(): value for key,value in self.headers.items()}
        incoming_headers_set = set(list(self.headers.keys()))
        is_missing_message_signing_headers = len(self.required_message_signing_headers.intersection(incoming_headers_set)) < len(self.required_message_signing_headers)
        if is_missing_message_signing_headers:
            error_message = "Incoming {} request to {} is missing message signing headers.\nRequired: {}\nProvided: {}".format(
            signature_info['method'], signature_info['url'], self.required_message_signing_headers, incoming_headers_set)
            logger.error(error_message)
        else:
            logger.info("{} request to {} has the required message signing headers. Analyzing...".format(
         signature_info['method'], signature_info['url']
        ))
            self.check_content_digests(interaction_id, signature_info, request_type)
            self.check_message_signing_headers(interaction_id, signature_info, request_type)
    
    def get_content_digest(self, payload):
        return base64.b64encode(hashlib.sha512(payload.encode('utf-8')).digest()).decode('utf-8')

    def check_content_digests(self, interaction_id, signature_info, request_type):
        payload = signature_info['body']
        content_digest_from_header = self.headers['content-digest']
        generated_content_digest = "sha-512=:{}:".format(self.get_content_digest(payload))
        if content_digest_from_header != generated_content_digest:
            error_message = "Content Digests did not match. \nFrom header: {} \nGenerated: {}".format(
            content_digest_from_header, generated_content_digest
            )
            logger.error(error_message)
            test_context = {
            'test_name': 'Checking for equality of content digest header in {} content-digest header'.format(request_type),
            'test_case': 'Content digest header in {} request to {} should be equal to generated value.'.format(flask.request.method, flask.request.path)}
            issue = {
                'context': test_context,
                'uss_role': "Notified USS",
                'target': "USS",
                'summary': error_message,
                'details': "{} != {}".format(content_digest_from_header, generated_content_digest),
                'interactions': [interaction_id]
            }
            report_settings.reprt_recorder.capture_issue(issue)
        else:
            logger.info("Content digests validated!")

    def check_message_signing_headers(self, interaction_id, signature_info, request_type):
        x_utm_jws_header = self.headers['x-utm-jws-header']
        x_utm_sig_input_header = self.headers['x-utm-message-signature-input']
        x_utm_sig_header = self.headers['x-utm-message-signature']
        logger.info("JWS Header: {}\nSigInput: \n{}, Sig:{}".format(
                x_utm_jws_header, x_utm_sig_input_header, x_utm_sig_header
        ))
        try:
            jws_map = self.http_dictionary_to_dict(x_utm_jws_header)
            endpoint_for_public_key = jws_map.get("x5u")
            logger.info("Public key is at: {}".format(endpoint_for_public_key))
        except Exception as e:
            error_message = "Could not get x5u from jws header during validation: {}".format(str(e))
            logger.error(error_message)
            issue = {
                'context': None,
                'uss_role': "Notified USS",
                'target': "USS",
                'summary': error_message,
                'details': "Method: {}, Url: {}, Headers: {}".format( signature_info['method'], signature_info['url'], json.dumps(self.headers, indent=4)),
                'interactions': [interaction_id]
            }
            report_settings.reprt_recorder.capture_issue(issue)
            return
        try:
            pub_key_response = requests.get(endpoint_for_public_key)
            pub_key_response.raise_for_status()
            self.public_key = RSA.import_key(pub_key_response.content)
        except Exception as e:
            error_message = "Could not get x5u from jws header during validation: {}".format(str(e))
            logger.error(error_message)
            issue = {
                'context': None,
                'uss_role': "Notified USS",
                'target': "USS",
                'summary': error_message,
                'details': "Method: {}, Url: {}, Headers: {}".format( signature_info['method'], signature_info['url'], json.dumps(self.headers, indent=4)),
                'interactions': [interaction_id]
            }
            report_settings.reprt_recorder.capture_issue(issue)
            return
        logger.info("Successfully retreived public key! Verifying signature...")
        self.verify(interaction_id, signature_info, request_type)
    
    def get_signature_base(self, signature_info, request_type):
        sig_input_from_header = self.headers['x-utm-message-signature-input']
        sig_input_start = sig_input_from_header.index("(")
        sig_input_end = len(sig_input_from_header)
        sig_input = sig_input_from_header[sig_input_start:sig_input_end]
        signature_param_str = "\"{}\": {}".format("@signature-params",sig_input)
        content_type =  self.headers.get("content-type", "")
        covered_components = self.request_covered_components if request_type == 'request' else self.response_covered_components
        logger.info("request_type is {}, covered components: {}".format(request_type, str(covered_components)))
        if ';' in content_type:
            content_type = content_type.split(';')[0].strip()
        if request_type == 'request':
            base_value_map = {
                "@method": flask.request.method,
                "@path": flask.request.path,
                "@query": "?" if not flask.request.query_string else flask.request.query_string,
                "authorization": self.headers['authorization'],
                "content-type": content_type,
                "content-digest": "sha-512=:{}:".format(self.get_content_digest(signature_info['body'])),
                "x-utm-jws-header": self.headers['x-utm-jws-header']
            }
        else:
            base_value_map = {
                "@status": signature_info['status'],
                "content-type": content_type,
                "content-digest": "sha-512=:{}:".format(self.get_content_digest(signature_info['body'])),
                "x-utm-jws-header": self.headers['x-utm-jws-header']
            }
        sig_base = ""
        for component in covered_components:
            sig_base += "\"{}\": {}\n".format(
                component, base_value_map[component]
            )
        sig_base += signature_param_str
        return sig_base
    
    def verify(self, interaction_id, signature_info, request_type):
        try:
            signature_value_from_header = self.headers['x-utm-message-signature']
            sig_start = signature_value_from_header.index(':') + 1
            sig_end = len(signature_value_from_header) - 1
            signature_value = signature_value_from_header[sig_start:sig_end]
            signature = base64.b64decode(signature_value)
            sig_base_str = self.get_signature_base(signature_info, request_type)
            recreated_sig_base = bytes(sig_base_str, 'utf-8')
            hash = SHA256.new(recreated_sig_base)
            logger.info("Validating this signature: {} with this recreated sig base: \n{}".format(signature_value, sig_base_str))
            pkcs1_15.new(self.public_key).verify(hash, signature)
            logger.info("Signature is valid!")

        except Exception as e:
            error_message = "Could not validate x-utm-message-signature header in {}!".format(request_type)
            logger.error(error_message)
            logger.error(str(e))
            test_context = {
            'test_name': 'Validating message signing headers',
            'test_case': 'Message signing headers should be valid'}
            issue = {
                'context': test_context,
                'uss_role': "Notified USS",
                'target': "USS",
                'summary': error_message,
                'details': str(e),
                'interactions': [interaction_id]
            }
            report_settings.reprt_recorder.capture_issue(issue)
    
    def http_dictionary_to_dict(self, http_dictionary):
       return {item.split('=')[0].strip().strip('\"'): item.split('=')[1].strip().strip('\"') for item in http_dictionary.split(',')}