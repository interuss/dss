import base64
import hashlib
import time
from Crypto.PublicKey import RSA
from Crypto.Signature import pkcs1_15
from Crypto.Hash import SHA256
import os
import json
from loguru import logger
import urllib.parse
from monitoring.messagesigning.config import Config

def get_x_utm_jws_header():
      return '\"alg\"=\"{}\", \"typ\"=\"{}\", \"kid\"=\"{}\", \"x5u\"=\"{}\"'.format(
        'RS256', 'JOSE', get_key_id(), 'http://host.docker.internal:8077/mock/scd/.well-known/uas-traffic-management/mock_pub.der'
      )

def get_signed_headers(object_to_sign):
    signed_type = str(type(object_to_sign))
    sig, sig_input = get_signature(object_to_sign, signed_type)
    content_digest = get_content_digest(convert_request_dot_body(object_to_sign.body)) if 'PreparedRequest' in signed_type else get_content_digest('' if not object_to_sign.json else json.dumps(object_to_sign.json))
    signed_headers = {
      'x-utm-message-signature-input': 'utm-message-signature={}'.format(sig_input),
      'x-utm-message-signature': 'utm-message-signature=:{}:'.format(sig),
      'x-utm-jws-header': get_x_utm_jws_header(),
      'content-digest': 'sha-512=:{}:'.format(content_digest)
    }
    return signed_headers

def get_content_digest(payload):
  payload = json.dumps(payload) if payload else ''
  return base64.b64encode(hashlib.sha512(payload.encode('utf-8')).digest()).decode('utf-8')

def get_signature_input(sig_base):
  sig_base_comps = sig_base.split('\n')
  sig_param_str = [item for item in sig_base_comps if '@signature-params' in item].pop()
  start_sig_input_ind = sig_param_str.index('(')
  end_sig_input_str = len(sig_param_str)
  return sig_param_str[start_sig_input_ind:end_sig_input_str]

def get_signature(object_to_sign, signed_type):
  sig_base = get_signature_base(object_to_sign, signed_type)
  sig_base_bytes =  bytes(sig_base, 'utf-8')
  sig_input = get_signature_input(sig_base)
  hash = SHA256.new(sig_base_bytes)
  with open(Config.PRIVATE_KEY_PATH, "rb") as priv_key_file:
    private_key = RSA.import_key(priv_key_file.read())
  return base64.b64encode(pkcs1_15.new(private_key).sign(hash)).decode("utf-8"), sig_input

def convert_request_dot_body(request_body):
  payload = '' if not request_body else request_body.decode('utf-8')
  try:
    payload = json.loads(payload)
  except Exception:
    if payload:
      logger.error("{} was not a valid json string....".format(payload))
  return payload

def get_key_id():
  return 'mock_uss_priv_key'

def get_signature_base(object_to_sign, signed_type):
  covered_components = ["@method", "@path", "@query", "authorization", "content-type", "content-digest", "x-utm-jws-header"] if 'Request' in signed_type else ["@status", "content-type", "content-digest", "x-utm-jws-header"]
  headers = {key.lower(): value for key,value in object_to_sign.headers.items()}
  content_digest = get_content_digest(convert_request_dot_body(object_to_sign.body)) if 'Request' in signed_type else get_content_digest('' if not object_to_sign.json else json.dumps(object_to_sign.json))
  if 'Request' in signed_type:
    parsed_url = urllib.parse.urlparse(object_to_sign.url)
    base_value_map = {
      "@method": object_to_sign.method,
      "@path": parsed_url.path,
      "@query": "?" if not parsed_url.query else parsed_url.query,
      "authorization": headers.get('authorization', ''),
      "content-type": headers.get('content-type', ''),
      "content-digest": "sha-512=:{}:".format(content_digest),
      "x-utm-jws-header": get_x_utm_jws_header()
      }
  else:
    base_value_map = {
      "@status": object_to_sign.status_code,
      "content-type": headers['content-type'],
      "content-digest": "sha-512=:{}:".format(content_digest),
      "x-utm-jws-header": get_x_utm_jws_header()
    }
  curr_time = str(int(time.time()))
  signature_param_str = "\"{}\": ({});{}".format("@signature-params", wrap_components(covered_components), curr_time)
  sig_base = ""
  for component in covered_components:
    sig_base += "\"{}\": {}\n".format(
                component, base_value_map[component]
    )
  sig_base += signature_param_str
  return sig_base

def wrap_components(components):
  return " ".join(list(map(lambda comp: "\"{}\"".format(comp), components)))
