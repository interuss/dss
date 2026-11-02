import subprocess

from cm_helper import CMHelper, Workspace

RAFT_1 = "workspace/_tests-workspace-1/raft"
RAFT_2 = "workspace/_tests-workspace-2/raft"


def verifies(ca_file: str, cert_file: str) -> bool:
    return (
        subprocess.run(
            ["openssl", "verify", "-CAfile", ca_file, cert_file], capture_output=True
        ).returncode
        == 0
    )


def test_generate_raft_nodes(cm: CMHelper):
    assert not cm.run_command(Workspace.WORKSPACE_1, ["generate-raft-nodes"])[0]
    assert cm.run_command(Workspace.WORKSPACE_1, ["init"])[0]
    assert cm.run_command(Workspace.WORKSPACE_2, ["init"])[0]
    assert not cm.run_command(
        Workspace.WORKSPACE_1, ["--nodes-public-address", '""', "generate-raft-nodes"]
    )[0]

    assert cm.run_command(Workspace.WORKSPACE_1, ["generate-raft-nodes"])[0]
    assert cm.run_command(Workspace.WORKSPACE_2, ["generate-raft-nodes"])[0]

    # Each certificate is valid for the public address of its node
    for node_id in range(3):
        san = subprocess.run(
            [
                "openssl",
                "x509",
                "-in",
                f"{RAFT_1}/node.raft-{node_id}.crt",
                "-noout",
                "-ext",
                "subjectAltName",
            ],
            capture_output=True,
            check=True,
        ).stdout.decode("utf-8")
        assert f"DNS:{node_id}.raft.w-1" in san

    assert not verifies(f"{RAFT_1}/ca.crt", f"{RAFT_2}/node.raft-0.crt")

    success, process = cm.run_command(Workspace.WORKSPACE_2, ["get-ca"])
    assert success
    ca_2 = process.stdout.decode("utf-8")

    assert cm.run_command(Workspace.WORKSPACE_1, ["add-pool-ca"], ca_2)[0]
    assert verifies(f"{RAFT_1}/ca.crt", f"{RAFT_2}/node.raft-0.crt")
    assert not verifies(f"{RAFT_2}/ca.crt", f"{RAFT_1}/node.raft-0.crt")

    assert cm.run_command(Workspace.WORKSPACE_1, ["remove-pool-ca"], ca_2)[0]
    assert not verifies(f"{RAFT_1}/ca.crt", f"{RAFT_2}/node.raft-0.crt")
