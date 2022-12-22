from monitoring.monitorlib.messagesigning.hasher import get_content_digest
import monitoring.monitorlib.messagesigning.message_signer as signer
import pytest
from requests import PreparedRequest


@pytest.fixture()
def prepared_request():
    prepped_req = PreparedRequest()
    prepped_req.headers = {
        "Content-type": "application/json",
        "Authorization": "Bearer some token",
    }
    prepped_req.method = "GET"
    prepped_req.url = "https://some-host/somePath"

    return prepped_req


def test_content_digest():
    expected_digest = "z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg=="
    empty_string_result = get_content_digest("")
    empty_bytes_result = get_content_digest(b"")
    null_result = get_content_digest(None)
    assert expected_digest == empty_string_result
    assert expected_digest == empty_bytes_result
    assert expected_digest == null_result


def test_get_x_utm_jws_header():
    exptected_jws_header = '"alg"="RS256", "typ"="JOSE", "kid"="mock_uss_priv_key", "x5u"="https://test-host.com/publickey"'
    assert exptected_jws_header == signer.get_x_utm_jws_header(
        "https://test-host.com/publickey"
    )


def test_get_signature_base_with_prepped_request(prepared_request):
    exptected_sig_base = """
    "@method": GET
    "@path": /somePath
    "@query": ?
    "authorization": Bearer some token
    "content-type": application/json
    "content-digest": sha-512=:z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg==:
    "x-utm-jws-header": "alg"="RS256", "typ"="JOSE", "kid"="mock_uss_priv_key", "x5u"="http://test-host/publickeyendpoint",
    @signature-params": ("@method" "@path" "@query" "authorization" "content-type" "content-digest" "x-utm-jws-header")
    """
    sig_base = signer.get_signature_base(
        prepared_request, "PreparedRequest", "http://test-host/publickeyendpoint"
    )

    # Strip timestamp, we just want to check sig base equality.
    token_ind = sig_base.index(";")
    sig_base = sig_base[0:token_ind]

    assert exptected_sig_base == sig_base
