import hashlib
import base64
from loguru import logger


def get_content_digest(payload):
    payload = payload if payload else bytes()
    type_of_payload = str(type(payload))
    if type_of_payload == "<class 'str'>":
        payload = payload.encode("utf-8")
    return base64.b64encode(hashlib.sha512(payload).digest()).decode("utf-8")
