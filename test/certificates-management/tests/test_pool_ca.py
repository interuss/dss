from cm_helper import CMHelper, Workspace


TEST_CA = """-----BEGIN CERTIFICATE-----
MIIDSTCCAjGgAwIBAgIUWFdoPeBGkAIHBpitaxnTQ/1l6lowDQYJKoZIhvcNAQEL
BQAwNDELMAkGA1UEBhMCQ0gxDTALBgNVBAgMBFRlc3QxFjAUBgNVBAoMDVVuaXQg
VGVzdHMgU0EwHhcNMjYwMjE5MDg1MDE4WhcNMjgxMjA5MDg1MDE4WjA0MQswCQYD
VQQGEwJDSDENMAsGA1UECAwEVGVzdDEWMBQGA1UECgwNVW5pdCBUZXN0cyBTQTCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALFwwv+tqU9TyAbMz8FxghpB
EDDYjGqFMair3FTlsj/ZgGPZABhhUCcD1Kbo7h948i45GL6IrTtgK+eie3sce4z/
lXCaQXU/v2RQkh/bO//3laDJ4NrK2LfIcUN7xQaWZWiNT8G9sjznqRAAXbomIF/P
r2oMtOeKjIcjhNNTqmeZjCYM008tgdWDH6S+UcxlDTFu14z5MmK8LoOujZERoz1U
co/ZFSzcUF9tvG8uVygO+u88VghU7K0cxX9pk+BJ736KJ56AV7A9/jgEuzCsKBNv
ersYCGTx53625TZp9Kt3gtCU/AIwRqsE5uEL5nwJmBKr4ZM0aqxyuSfvHZAMz98C
AwEAAaNTMFEwHQYDVR0OBBYEFDfWctmPow1RuTEb+FBD1fn9GZT3MB8GA1UdIwQY
MBaAFDfWctmPow1RuTEb+FBD1fn9GZT3MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggEBAJ2r29qnsUQEzHif3UOJOZa9Ga0kBpc/jCOuoh5Sfdjkb0/+
uYvKBsKQNbn31b3bmuthijGlWMuULrQch/iM2DzK+EcvMN23jJgAH7whQpjsBXeY
Xslhp8pDVBeureSrk2yMPvcZryAVbYp090x1uOwiCJNmf5xDpgVzuG3onPaU65z0
8p9C3JRtAg+4KbdNPPeJSeDwHQkaF5zKPUYfgqurbpmKkJq198kBpVg8kc0QIpJ8
29qb1qqyFHEvFIlL3m/IpkmCHZB+jhWwOsJunyluUQ90WI1Ghcn3WRTH5G1m7QYs
PI+APTPVUUTqvkWllAC+jlil5ID9ooaqY+16SIM=
-----END CERTIFICATE-----"""

TEST_CA_SN = "FD65EA5A"


def test_pool_ca(cm: CMHelper):

    assert cm.run_command(Workspace.WORKSPACE_1, ["init"])[0]

    # Test CA must not be present in the pool initially
    success, process = cm.run_command(Workspace.WORKSPACE_1, ["get-pool-ca"])
    assert success
    assert TEST_CA not in process.stdout.decode("utf-8")

    # We add the test CA
    assert cm.run_command(Workspace.WORKSPACE_1, ["add-pool-ca"], TEST_CA)[0]

    # Now it should be in the pool CA
    success, process = cm.run_command(Workspace.WORKSPACE_1, ["get-pool-ca"])
    assert success
    assert TEST_CA in process.stdout.decode("utf-8")

    success, process = cm.run_command(Workspace.WORKSPACE_1, ["list-pool-ca"])
    assert success
    assert TEST_CA_SN in process.stdout.decode("utf-8")

    # We remove the test CA
    assert cm.run_command(Workspace.WORKSPACE_1, ["remove-pool-ca"], TEST_CA)[0]

    # It shouldn't be in the pool CA anymore
    success, process = cm.run_command(Workspace.WORKSPACE_1, ["get-pool-ca"])
    assert success
    assert TEST_CA not in process.stdout.decode("utf-8")

    # We add it back to test removal via serial number
    assert cm.run_command(Workspace.WORKSPACE_1, ["add-pool-ca"], TEST_CA)[0]

    success, process = cm.run_command(Workspace.WORKSPACE_1, ["get-pool-ca"])
    assert success
    assert TEST_CA in process.stdout.decode("utf-8")

    assert cm.run_command(
        Workspace.WORKSPACE_1, ["remove-pool-ca", "--ca-serial", TEST_CA_SN]
    )[0]

    success, process = cm.run_command(Workspace.WORKSPACE_1, ["get-pool-ca"])
    assert success
    assert TEST_CA not in process.stdout.decode("utf-8")
