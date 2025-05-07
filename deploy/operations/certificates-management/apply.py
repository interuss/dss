import subprocess
import os

import logging
l = logging.getLogger(__name__)

def do_apply(cluster):

    l.debug("Applying kubernetes configuration")

    l.debug(f"Creating namespace {cluster.namespace}")

    try:
        subprocess.check_call(
            ["kubectl", "create", "namespace", cluster.namespace, "--context", cluster.cluster_context],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

        l.info(f"Created namespace {cluster.namespace}")

    except subprocess.CalledProcessError:  # We do assume everything else works
        l.debug(f"Namespace {cluster.namespace} already exists")

    for secret in ["yb-master-yugabyte-tls-cert", "yb-tserver-yugabyte-tls-cert", "yugabyte-tls-client-cert", "dss.public.certs"]:

        try:
            subprocess.check_call(
                ["kubectl", "delete", "secret", secret, "--namespace", cluster.namespace, "--context", cluster.cluster_context],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )

            l.info(f"Deleted old secret '{secret}'")

        except subprocess.CalledProcessError:  # We do assume everything else works
            l.debug(f"Secret '{secret}' not present on the cluster")

    for secret, folder in [
        ("yb-master-yugabyte-tls-cert", cluster.master_certs_dir),
        ("yb-tserver-yugabyte-tls-cert", cluster.tserver_certs_dir),
        ("yugabyte-tls-client-cert", cluster.client_certs_dir),
        ("dss.public.certs", os.path.join("..", "..", "..", "build", "jwt-public-certs")),
    ]:

        subprocess.check_call(
            ["kubectl", "create", "secret", "generic", secret, "--namespace", cluster.namespace, "--context", cluster.cluster_context, "--from-file", folder],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

        l.info(f"Created secret '{secret}'")
