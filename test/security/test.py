# Small python script to do authentification tests. Use only standard libraries to
# run everywhere.
# Extract URLs directly from the code to be avoid the need of updated API and
# to ensure we get all patterns from the code.
# Assume all endpoint need authentification, but some can be whitelisted there.
# Expect that `start-locally` have been ran

import sys
import logging
import re
import glob
import json
from urllib import request
import http.client
import base64
import hmac
import hashlib

ENDPOINT_WITHOUT_AUTHS = [
    "/aux/v1/configuration/accepted_ca_certs",
    "/aux/v1/configuration/ca_certs",
    "/aux/v1/version",
]

ALL_SCOPES = [
    "utm.constraint_management",
    "utm.conformance_monitoring_sa",
    "utm.strategic_coordination",
    "rid.service_provider",
    "rid.display_provider",
    "dss.write.identification_service_areas",
    "dss.read.identification_service_areas",
    "interuss.pool_status.read",
    "interuss.pool_status.heartbeat.write",
    "utm.availability_arbitration",
]


logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s %(levelname)-8s %(name)-10s %(message)-50s",
    handlers=[logging.StreamHandler(sys.stdout)],
)

logger = logging.getLogger("security")


def build_urls():

    pattern = re.compile(r'Method:\s*http\.Method(\w+).*?Path:\s*"([^"]+)"')

    urls = set()

    for filepath in glob.glob("pkg/api/**/*.go", recursive=True):
        with open(filepath) as f:
            for match in pattern.finditer(f.read()):
                method = match.group(1).upper()
                path = match.group(2)
                urls.add((method, path))

    if not urls:
        logger.error("❌ No URL found.")
        sys.exit(1)

    return urls


def get_token(scope=None, audience=None, expire=None):

    if scope:
        scopes = [scope]
    else:
        scopes = ALL_SCOPES

    if not audience:
        audience = "localhost"

    r = request.urlopen(
        f"http://localhost:8085/token?grant_type=client_credentials&scope={'%20'.join(scopes)}&intended_audience={audience}&issuer=localhost&sub=test_security{'&expire=' + expire if expire else ''}"
    ).read()
    data = json.loads(r)

    if "access_token" not in data:
        logger.error(
            "❌ Unable to retrieve access token. Is the dummy auth server running?"
        )
        sys.exit(1)

    return data["access_token"]


def fill_path(path):
    return re.sub(r"\{[^}]+\}", "test", path)


def test_no_authentification(method, path):

    logger.debug("❓ Testing without authentification...")

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(method, fill_path(path))
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 without authentification header."
        )
        sys.exit(1)
    else:
        logger.info("✅ No authentification generated a 401.")


def test_wrong_signature(method, path):

    logger.debug("❓ Testing a token with a wrong signature...")

    invalid_signature = ".".join(get_token().split(".")[:-1] + ["testsig"])

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(
        method,
        fill_path(path),
        headers={"Authorization": f"Bearer {invalid_signature}"},
    )
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 with a token with a wrong signature."
        )
        sys.exit(1)
    else:
        logger.info("✅ Token with a wrong signature generated a 401.")


def test_wrong_scope(method, path):

    logger.debug("❓ Testing a token with a wrong scope...")

    invalid_scope = get_token("test_wrong_scope")

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(
        method, fill_path(path), headers={"Authorization": f"Bearer {invalid_scope}"}
    )
    resp = conn.getresponse()
    if resp.status != 403:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 403 with a token with a wrong scope."
        )
        sys.exit(1)
    else:
        logger.info("✅ Token with a wrong scope generated a 403.")


def test_wrong_audience(method, path):

    logger.debug("❓ Testing a token with a wrong audience...")

    invalid_audience = get_token(audience="test_wrong_audience")

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(
        method, fill_path(path), headers={"Authorization": f"Bearer {invalid_audience}"}
    )
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 with a token with a wrong audience."
        )
        sys.exit(1)
    else:
        logger.info("✅ Token with a wrong audience generated a 401.")


def test_expired(method, path):

    logger.debug("❓ Testing an expired token...")

    invalid_audience = get_token(expire="42")  # very old timestamp

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(
        method, fill_path(path), headers={"Authorization": f"Bearer {invalid_audience}"}
    )
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 with an expired token."
        )
        sys.exit(1)
    else:
        logger.info("✅ Expired token generated a 401.")


def test_hs256_token(method, path):

    logger.debug("❓ Testing HS256 algorithm confusion attack...")

    valid_token = get_token()
    _, payload_b64, _ = valid_token.split(".")

    # Read the public key as the HMAC secret
    with open("build/test-certs/auth2.pem", "rb") as f:
        secret = f.read()

    header = (
        base64.urlsafe_b64encode(b'{"alg":"HS256","typ":"JWT"}').rstrip(b"=").decode()
    )
    signing_input = f"{header}.{payload_b64}".encode()
    sig = hmac.new(secret, msg=signing_input, digestmod=hashlib.sha256).digest()
    sig_b64 = base64.urlsafe_b64encode(sig).rstrip(b"=").decode()
    token = f"{header}.{payload_b64}.{sig_b64}"

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(method, fill_path(path), headers={"Authorization": f"Bearer {token}"})
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 with a HS256 confusion token."
        )
        sys.exit(1)
    else:
        logger.info("✅ HS256 confusion attack generated a 401.")


def test_alg_none(method, path):

    logger.debug("❓ Testing alg:none attack")

    valid_token = get_token()
    _, payload_b64, _ = valid_token.split(".")

    header = (
        base64.urlsafe_b64encode(b'{"alg":"none","typ":"JWT"}').rstrip(b"=").decode()
    )
    token = f"{header}.{payload_b64}."

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(method, fill_path(path), headers={"Authorization": f"Bearer {token}"})
    resp = conn.getresponse()
    if resp.status != 401:
        logger.error(
            f"❌ Unexpected response {resp.status} instead of 401 with alg:none token."
        )
        sys.exit(1)
    else:
        logger.info("✅ alg:none attack generated a 401.")


def test_valid_token(method, path):

    logger.debug("❓ Testing a valid token...")

    valid_token = get_token()

    conn = http.client.HTTPConnection("localhost:8082")
    conn.request(
        method, fill_path(path), headers={"Authorization": f"Bearer {valid_token}"}
    )
    resp = conn.getresponse()
    if resp.status in [401, 403]:
        logger.error(f"❌ Unexpected response {resp.status} since the token in valid.")
        sys.exit(1)
    else:
        logger.info(f"✅ Valid token generated a {resp.status}.")


urls = build_urls()

for method, path in sorted(urls):
    logger.info(f"📋 Testing {method} {path}...")

    if path in ENDPOINT_WITHOUT_AUTHS:
        logger.info("✅ Endpoint is not protected.")
        continue

    test_no_authentification(method, path)
    test_wrong_signature(method, path)
    test_wrong_scope(method, path)
    test_wrong_audience(method, path)
    test_expired(method, path)
    test_hs256_token(method, path)
    test_alg_none(method, path)
    test_valid_token(method, path)
