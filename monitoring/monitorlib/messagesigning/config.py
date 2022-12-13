import os


class Config:
    BASE_PATH = os.path.dirname(os.path.abspath(__file__))
    CERT_PATH = '/var/test-certs'
    PRIVATE_KEY_PATH = "{}/messagesigning/mock_faa_priv.pem".format(CERT_PATH)
    PUBLIC_KEY_PATH = "{}/messagesigning/mock_faa_pub.der".format(CERT_PATH)