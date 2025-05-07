import logging
import subprocess
import sys

from utils import get_cert_display_name


def generate_node_config(cluster, node_type, node_id):

    l.debug(f"Creating {node_type} #{node_id} configuration file")

    short_name = cluster.get_node_short_name(node_type, node_id)
    short_name_group = cluster.get_node_short_name_group(node_type, node_id)
    full_name = cluster.get_node_full_name(node_type, node_id)
    full_name_without_group = cluster.get_node_full_name_without_group(
        node_type, node_id
    )

    with open(cluster.get_node_conf_file(node_type, node_id), "w") as f:
        f.write(
            f"""[ req ]
prompt=no
distinguished_name = my_distinguished_name

[ my_distinguished_name ]
organizationName = {cluster.organization}
commonName = {full_name}

# Multiple subject alternative names (SANs) such as IP Address,
# DNS Name, Email, URI, and so on, can be specified under this section
[ req_ext]
subjectAltName = @alt_names
[alt_names]
DNS.1 = {short_name}
DNS.2 = {full_name}
DNS.3 = {short_name_group}
DNS.4 = {full_name_without_group}
DNS.5 = yb-{node_type}s
DNS.6 = yb-{node_type}s.{cluster.namespace}
DNS.7 = yb-{node_type}s.{cluster.namespace}.svc.cluster.local
"""
        )

    l.info(f"Created {node_type} #{node_id} configuration file")


def generate_node_key(cluster, node_type, node_id):

    l.debug(f"Generating {node_type} #{node_id} private key")

    subprocess.check_call(
        [
            "openssl",
            "genrsa",
            "-out",
            cluster.get_node_key_file(node_type, node_id),
            "4096",
        ]
    )

    l.info(f"Generated {node_type} #{node_id} private key")


def generate_node_csr(cluster, node_type, node_id):

    l.debug(f"Generating {node_type} #{node_id} certificate request")

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

    l.info(f"Generated {node_type} #{node_id} certificate request")


def generate_node_cert(cluster, node_type, node_id):

    l.debug(f"Generating {node_type} #{node_id} certificate")

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
            cluster.get_node_cert_file(node_type, node_id),
            "-outdir",
            getattr(cluster, f"{node_type}_certs_dir"),
            "-in",
            cluster.get_node_csr_file(node_type, node_id),
            "-days",
            "3650",
            "-batch",
            "-extfile",
            cluster.get_node_conf_file(node_type, node_id),
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    name = get_cert_display_name(cluster.get_node_cert_file(node_type, node_id))

    l.info(f"Generated {node_type} #{node_id} certificate '{name}'")


def generate_node(cluster, node_type, node_id):
    if cluster.is_node_ready(node_type, node_id):
        l.debug(f"{node_type} #{node_id} certificiates already generated")
        return

    generate_node_config(cluster, node_type, node_id)
    generate_node_key(cluster, node_type, node_id)
    generate_node_csr(cluster, node_type, node_id)
    generate_node_cert(cluster, node_type, node_id)


l = logging.getLogger(__name__)


def do_generate_nodes(cluster):
    """Generate certificates for all nodes (master and tserver)"""

    l.info("Generation of nodes certificates")

    if not cluster.is_ready:
        l.error("Cluster is not already initialized, unable to continue")
        sys.exit(1)
    else:
        l.debug("Cluster is initialized, continuing")

    for node_type in ["master", "tserver"]:
        for node_id in range(0, int(cluster.nodes_count)):
            generate_node(cluster, node_type, node_id)

    l.info("All nodes certificates are ready")
