from cm_helper import CMHelper, Workspace


def test_regenerate(cm: CMHelper):

    assert cm.run_command(Workspace.WORKSPACE_1, ["init"])[0]
    assert cm.run_command(Workspace.WORKSPACE_1, ["regenerate-nodes"])[0]
