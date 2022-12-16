import base64
import time
from Crypto.PublicKey import RSA
from Crypto.Signature import pkcs1_15
from Crypto.Hash import SHA256
from loguru import logger
import urllib.parse
from monitoring.monitorlib.messagesigning.hasher import get_content_digest


def get_x_utm_jws_header(cert_url):
    return '"alg"="{}", "typ"="{}", "kid"="{}", "x5u"="{}"'.format(
        "RS256",
        "JOSE",
        _get_key_id(),
        cert_url,
    )


def get_signed_headers(object_to_sign, private_key_path, cert_url):
    signed_type = str(type(object_to_sign))
    is_signing_request = "PreparedRequest" in signed_type
    sig, sig_input = get_signature(object_to_sign, signed_type, private_key_path)
    content_digest = (
        get_content_digest(object_to_sign.body)
        if is_signing_request
        else get_content_digest(object_to_sign.get_data())
    )
    signed_headers = {
        "x-utm-message-signature-input": "utm-message-signature={}".format(sig_input),
        "x-utm-message-signature": "utm-message-signature=:{}:".format(sig),
        "x-utm-jws-header": get_x_utm_jws_header(cert_url),
        "content-digest": "sha-512=:{}:".format(content_digest),
    }
    return signed_headers


def get_signature_input(sig_base):
    sig_base_comps = sig_base.split("\n")
    sig_param_str = [
        item for item in sig_base_comps if "@signature-params" in item
    ].pop()
    start_sig_input_ind = sig_param_str.index("(")
    end_sig_input_str = len(sig_param_str)
    return sig_param_str[start_sig_input_ind:end_sig_input_str]


def get_signature(object_to_sign, signed_type, private_key_path):
    sig_base = get_signature_base(object_to_sign, signed_type)
    sig_base_bytes = bytes(sig_base, "utf-8")
    sig_input = get_signature_input(sig_base)
    hash = SHA256.new(sig_base_bytes)
    with open(private_key_path, "rb") as priv_key_file:
        private_key = RSA.import_key(priv_key_file.read())
    return (
        base64.b64encode(pkcs1_15.new(private_key).sign(hash)).decode("utf-8"),
        sig_input,
    )


def _get_key_id():
    return "mock_uss_priv_key"


def get_signature_base(object_to_sign, signed_type, cert_url):
    is_signing_requests = "PreparedRequest" in signed_type
    covered_components = (
        [
            "@method",
            "@path",
            "@query",
            "authorization",
            "content-type",
            "content-digest",
            "x-utm-jws-header",
        ]
        if is_signing_requests
        else ["@status", "content-type", "content-digest", "x-utm-jws-header"]
    )
    headers = {key.lower(): value for key, value in object_to_sign.headers.items()}
    is_signing_requests = "PreparedRequest" in signed_type
    content_digest = (
        get_content_digest(object_to_sign.body)
        if is_signing_requests
        else get_content_digest(object_to_sign.get_data())
    )
    if is_signing_requests:
        parsed_url = urllib.parse.urlparse(object_to_sign.url)
        base_value_map = {
            "@method": object_to_sign.method,
            "@path": parsed_url.path,
            "@query": "?" if not parsed_url.query else parsed_url.query,
            "authorization": headers.get("authorization", ""),
            "content-type": headers.get("content-type", ""),
            "content-digest": "sha-512=:{}:".format(content_digest),
            "x-utm-jws-header": get_x_utm_jws_header(cert_url),
        }
    else:
        base_value_map = {
            "@status": object_to_sign.status_code,
            "content-type": headers["content-type"],
            "content-digest": "sha-512=:{}:".format(content_digest),
            "x-utm-jws-header": get_x_utm_jws_header(),
        }
    curr_time = str(int(time.time()))
    signature_param_str = '"{}": ({});{}'.format(
        "@signature-params", _wrap_components_in_quotes(covered_components), curr_time
    )
    sig_base = ""
    for component in covered_components:
        sig_base += '"{}": {}\n'.format(component, base_value_map[component])
    sig_base += signature_param_str
    return sig_base


def _wrap_components_in_quotes(components):
    return " ".join(list(map(lambda comp: '"{}"'.format(comp), components)))
