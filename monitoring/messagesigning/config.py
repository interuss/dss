import os


class Config:
    BASE_PATH = os.getcwd()
    PRIVATE_KEY_PATH = "{}/keys/priv.pem".format(BASE_PATH)
    PUBLIC_KEY_PATH = "{}/keys/pub.der".format(BASE_PATH)