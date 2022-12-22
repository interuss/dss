def get_x_utm_jws_header(cert_url):
    return 'alg="{}", typ="{}", kid="{}", x5u="{}"'.format(
        "RS256",
        "JOSE",
        "mock_uss_priv_key",
        cert_url,
    )
