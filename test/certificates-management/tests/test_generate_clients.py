from cm_helper import CMHelper, Workspace


def test_generate_clients(cm: CMHelper):

    assert cm.run_command(Workspace.WORKSPACE_1, ["init"])[0]
    assert cm.run_command(Workspace.WORKSPACE_1, ["generate-clients"])[0]
