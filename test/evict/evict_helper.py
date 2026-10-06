import sys
import subprocess
import logging
import os


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
        delete: bool = False,
    ):
        db_hostname = os.environ.get("DB_HOSTNAME", "local-dss-crdb")
        db_port = os.environ.get("DB_PORT", "26257")
        db_username = os.environ.get("DB_USERNAME", "root")

        command = [
            "docker",
            "exec",
            "dss_sandbox-local-dss-core-service-1",
            "db-manager",
            "evict",
            f"--scd_oir={str(scd_oir).lower()}",
            f"--scd_sub={str(scd_sub).lower()}",
            f"--rid_isa={str(rid_isa).lower()}",
            f"--rid_sub={str(rid_sub).lower()}",
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
        limit: int | None = None,
    ):
        self.run_evict(rid_isa=True, delete=delete, rid_ttl=ttl, rid_limit=limit)

    def evict_rid_subscriptions(
        self,
        ttl: str,
        delete: bool,
        limit: int | None = None,
    ):
        self.run_evict(rid_sub=True, delete=delete, rid_ttl=ttl, rid_limit=limit)
