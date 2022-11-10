import os


class Config:
    BASE_PATH = os.path.dirname(__file__)
    PRIVATE_KEY_PATH = "{}/keys/mock_priv.pem".format(BASE_PATH)
    PUBLIC_KEY_PATH = "{}/keys/mock_pub.der".format(BASE_PATH)
