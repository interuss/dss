import logging
import os
import shutil
import subprocess
import sys

from utils import get_cert_display_name

logger = logging.getLogger(__name__)


def generate_node_config(cluster, node_type, node_id):
    logger.debug(f"Creating {node_type} #{node_id} configuration file")

    short_name = cluster.get_node_short_name(node_type, node_id)
    short_name_group = cluster.get_node_short_name_group(node_type, node_id)
    full_name = cluster.get_node_full_name(node_type, node_id)
    full_name_without_group = cluster.get_node_full_name_without_group(
        node_type, node_id
    )
    public_address = cluster.get_node_public_address(node_type, node_id)

    # Raft nodes are reached by their peers at their public address only.
    if node_type == "raft":
        alt_names = [public_address]
    else:
        alt_names = [
            short_name,
            full_name,
            short_name_group,
            full_name_without_group,
            f"yb-{node_type}s",
            f"yb-{node_type}s.{cluster.namespace}",
            f"yb-{node_type}s.{cluster.namespace}.svc.cluster.local",
        ]
        if public_address:
            alt_names.append(public_address)

    with open(cluster.get_node_conf_file(node_type, node_id), "w") as f:
        f.write(
            f"""[ req ]
prompt=no
distinguished_name = my_distinguished_name

[ my_distinguished_name ]
organizationName = {cluster.organization}
commonName = {short_name}

# Multiple subject alternative names (SANs) such as IP Address,
# DNS Name, Email, URI, and so on, can be specified under this section
[ req_ext ]
subjectAltName = @alt_names
[ alt_names ]
"""
        )
        for i, name in enumerate(alt_names, 1):
            f.write(f"DNS.{i} = {name}\n")

    logger.info(f"Created {node_type} #{node_id} configuration file")


def generate_node_key(cluster, node_type, node_id):
    logger.debug(f"Generating {node_type} #{node_id} private key")

    file = cluster.get_node_key_file(node_type, node_id)

    subprocess.check_call(
        [
            "openssl",
            "genrsa",
            "-out",
            file,
            "4096",
        ]
    )

    second_file = cluster.get_node_key_second_file(node_type, node_id)

    if second_file:
        shutil.copy(file, second_file)

    logger.info(f"Generated {node_type} #{node_id} private key")


def generate_node_csr(cluster, node_type, node_id):
    logger.debug(f"Generating {node_type} #{node_id} certificate request")

    subprocess.check_call(
        [
            "openssl",
            "req",
            "-new",
            "-config",
            cluster.get_node_conf_file(node_type, node_id),
            "-key",
            cluster.get_node_key_file(node_type, node_id),
            "-out",
            cluster.get_node_csr_file(node_type, node_id),
        ],
        stdout=subprocess.DEVNULL,
    )

    logger.info(f"Generated {node_type} #{node_id} certificate request")


def generate_node_cert(cluster, node_type, node_id):
    logger.debug(f"Generating {node_type} #{node_id} certificate")

    file = cluster.get_node_cert_file(node_type, node_id)

    subprocess.check_call(
        [
            "openssl",
            "ca",
            "-config",
            cluster.ca_conf,
            "-keyfile",
            cluster.ca_key_file,
            "-cert",
            cluster.ca_cert_file,
            "-policy",
            "my_policy",
            "-out",
            file,
            "-outdir",
            getattr(cluster, f"{node_type}_certs_dir"),
            "-in",
            cluster.get_node_csr_file(node_type, node_id),
            "-days",
            "3650",
            "-batch",
            "-extfile",
            cluster.get_node_conf_file(node_type, node_id),
            "-extensions",
            "req_ext",
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    second_file = cluster.get_node_cert_second_file(node_type, node_id)

    if second_file:
        shutil.copy(file, second_file)

    name = get_cert_display_name(cluster.get_node_cert_file(node_type, node_id))

    logger.info(f"Generated {node_type} #{node_id} certificate '{name}'")


def generate_node(cluster, node_type, node_id):
    if cluster.is_node_ready(node_type, node_id):
        logger.debug(f"{node_type} #{node_id} certificates already generated")
        return

    generate_node_config(cluster, node_type, node_id)
    generate_node_key(cluster, node_type, node_id)
    generate_node_csr(cluster, node_type, node_id)
    generate_node_cert(cluster, node_type, node_id)

    shutil.copy(cluster.ca_pool_ca, getattr(cluster, f"{node_type}_ca"))


def do_generate_nodes(cluster):
    """Generate certificates for all nodes (master and tserver)"""

    logger.info("Generation of nodes certificates")

    if not cluster.is_ready:
        logger.error("Cluster is not already initialized, unable to continue")
        sys.exit(1)
    else:
        logger.debug("Cluster is initialized, continuing")

    for node_type in ["master", "tserver"]:
        for node_id in range(0, int(cluster.nodes_count)):
            generate_node(cluster, node_type, node_id)

    generate_node(cluster, "prometheus", "")

    logger.info("All nodes certificates are ready")


def do_generate_raft_nodes(cluster):
    """Generate certificates for all raft nodes"""

    logger.info("Generation of raft nodes certificates")

    if not cluster.is_ready:
        logger.error("Cluster is not already initialized, unable to continue")
        sys.exit(1)

    if not cluster.nodes_public_address:
        logger.error("--nodes-public-address is required for raft nodes")
        sys.exit(1)

    os.makedirs(cluster.raft_certs_dir, exist_ok=True)

    for node_id in range(0, int(cluster.nodes_count)):
        generate_node(cluster, "raft", node_id)

    logger.info("All raft nodes certificates are ready")
