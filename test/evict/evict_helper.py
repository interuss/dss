import sys
import subprocess
import logging
import os
import tempfile

CORE_SERVICE_CONTAINER = "dss_sandbox-local-dss-core-service-1"


class EvictHelper:
    def __init__(self):
        self.logger: logging.Logger = logging.getLogger(__name__)

    def run_evict(
        self,
        scd_oir: bool = False,
        scd_sub: bool = False,
        rid_isa: bool = False,
        rid_sub: bool = False,
        scd_ttl: str | None = None,
        rid_ttl: str | None = None,
        scd_limit: int | None = None,
        rid_limit: int | None = None,
        locality: str = "local_dev",
        delete: bool = False,
        output_file: str | None = None,
        input_file: str | None = None,
    ):
        db_hostname = os.environ.get("DB_HOSTNAME", "local-dss-crdb")
        db_port = os.environ.get("DB_PORT", "26257")
        db_username = os.environ.get("DB_USERNAME", "root")

        command = [
            "docker",
            "exec",
            CORE_SERVICE_CONTAINER,
            "db-manager",
            "evict",
            f"--scd_oir={str(scd_oir).lower()}",
            f"--scd_sub={str(scd_sub).lower()}",
            f"--rid_isa={str(rid_isa).lower()}",
            f"--rid_sub={str(rid_sub).lower()}",
            "--locality",
            locality,
            "--datastore_host",
            db_hostname,
            "--datastore_port",
            db_port,
            "--datastore_user",
            db_username,
        ]

        if delete:
            command.append("--delete")

        if scd_ttl:
            command += [
                "--scd_ttl",
                str(scd_ttl).lower(),
            ]

        if rid_ttl:
            command += [
                "--rid_ttl",
                str(rid_ttl).lower(),
            ]

        if scd_limit is not None:
            command += [
                "--scd_limit",
                str(scd_limit),
            ]

        if rid_limit is not None:
            command += [
                "--rid_limit",
                str(rid_limit),
            ]

        if output_file:
            command += ["--output", output_file]

        if input_file:
            command += ["--input", input_file]

        process = subprocess.run(
            " ".join(command), shell=True, capture_output=True, timeout=5
        )

        if process.returncode != 0:
            self.logger.error("❌ Unable to run evict command")
            self.logger.error(process.stdout.decode("utf-8"))
            self.logger.error(process.stderr.decode("utf-8"))
            sys.exit(1)

    def evict_scd_operational_intents(
        self, ttl: str, delete: bool, limit: int | None = None
    ):
        self.run_evict(scd_oir=True, delete=delete, scd_ttl=ttl, scd_limit=limit)

    def evict_scd_subscriptions(self, ttl: str, delete: bool, limit: int | None = None):
        self.run_evict(scd_sub=True, delete=delete, scd_ttl=ttl, scd_limit=limit)

    def evict_rid_ISAs(
        self,
        ttl: str,
        delete: bool,
        locality: str = "local_dev",
        limit: int | None = None,
    ):
        self.run_evict(
            rid_isa=True, delete=delete, rid_ttl=ttl, locality=locality, rid_limit=limit
        )

    def evict_rid_subscriptions(
        self,
        ttl: str,
        delete: bool,
        locality: str = "local_dev",
        limit: int | None = None,
    ):
        self.run_evict(
            rid_sub=True, delete=delete, rid_ttl=ttl, locality=locality, rid_limit=limit
        )

    def read_file(self, path: str) -> str:
        """Returns the content of a file of the core service container."""
        with tempfile.TemporaryDirectory() as tmp:
            local_path = os.path.join(tmp, "file")
            self._docker_cp(f"{CORE_SERVICE_CONTAINER}:{path}", local_path)
            with open(local_path) as f:
                return f.read()

    def write_file(self, path: str, content: str):
        """Writes content to a file of the core service container."""
        process = subprocess.run(
            ["docker", "exec", "-i", CORE_SERVICE_CONTAINER, "sh", "-c", f"cat > {path}"],
            input=content.encode("utf-8"),
            capture_output=True,
            timeout=5,
        )
        if process.returncode != 0:
            self.logger.error(f"❌ Unable to write {path}")
            self.logger.error(process.stderr.decode("utf-8"))
            sys.exit(1)

    def _docker_cp(self, source: str, destination: str):
        process = subprocess.run(
            ["docker", "cp", source, destination], capture_output=True, timeout=5
        )
        if process.returncode != 0:
            self.logger.error(f"❌ Unable to copy {source} to {destination}")
            self.logger.error(process.stderr.decode("utf-8"))
            sys.exit(1)
