import subprocess
import logging
from typing import Optional

from enum import Enum


class Workspace(Enum):
    WORKSPACE_1 = "1"
    WORKSPACE_2 = "2"


class CMHelper:
    def __init__(self):
        self.logger: logging.Logger = logging.getLogger(__name__)

    def prepare(self):
        self.run_command(Workspace.WORKSPACE_1, ["destroy"], "y")
        self.run_command(Workspace.WORKSPACE_2, ["destroy"], "y")

    def _build_workspace_args(self, workspace: Workspace) -> list[str]:
        return [
            "--name",
            f"_tests-workspace-{workspace.value}",
            "--cluster-context",
            f"_cluster-tests-workspace-{workspace.value}",
            "--namespace",
            f"ns-{workspace.value}",
            "--nodes-public-address",
            f'"<ID>.<TYPE>.w-{workspace.value}"',
        ]

    def run_command(
        self, workspace: Workspace, args: list[str], stdin: Optional[str] = None
    ) -> tuple[bool, subprocess.CompletedProcess]:
        command = (
            [
                "python",
                "../../deploy/operations/certificates-management/dss-certs.py",
            ]
            + self._build_workspace_args(workspace)
            + args
        )

        process = subprocess.run(
            " ".join(command),
            shell=True,
            capture_output=True,
            input=stdin.encode("utf-8") if stdin else None,
        )

        return process.returncode == 0, process
