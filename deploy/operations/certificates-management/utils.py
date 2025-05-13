import logging
import re
import ssl
import sys
import unicodedata

l = logging.getLogger(__name__)


def slugify(text):
    text = unicodedata.normalize("NFKD", text).encode("ascii", "ignore").decode("ascii")
    text = text.lower()
    text = re.sub(r"[^a-z0-9_\.]+", "-", text)
    text = text.strip("-")
    return text


def get_cert_display_name(path):
    try:
        cert_dict = ssl._ssl._test_decode_cert(
            path
        )  # We do use an internal function, to avoid installing dependencies
    except Exception as e:
        l.error(e)
        sys.exit(1)

    serial = cert_dict.get("serialNumber", "")

    orga = ""
    cn = ""

    for kv in cert_dict.get("subject", []):
        for k, v in kv:
            if k == "organizationName":
                orga = v
            elif k == "commonName":
                cn = v

    return f"SN={serial[-8:]}, O={orga}, CN={cn}"


def get_cert_serial(path):
    try:
        cert_dict = ssl._ssl._test_decode_cert(
            path
        )  # We do use an internal function, to avoid installing dependencies
    except Exception as e:
        l.error(e)
        sys.exit(1)

    return cert_dict["serialNumber"]
