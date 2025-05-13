import logging
import os
import subprocess
import sys

from ca_pool import do_add_cas
from nodes import do_generate_nodes
from utils import get_cert_display_name

l = logging.getLogger(__name__)


def generate_ca_config(cluster):
    l.debug("Creating CA configuration files")

    with open(cluster.ca_conf, "w") as f:
        f.write(
            f"""
        [ ca ]
        default_ca = my_ca

[ my_ca ]
default_days = 3650

serial = {cluster.ca_key_dir}/serial.txt
database = {cluster.ca_key_dir}/index.txt
default_md = sha256
policy = my_policy

[ my_policy ]

organizationName = supplied
commonName = supplied

[req]
prompt=no
distinguished_name = my_distinguished_name
x509_extensions = my_extensions

[ my_distinguished_name ]
organizationName = {cluster.organization}
commonName = CA.{cluster.name}

[ my_extensions ]
keyUsage = critical,digitalSignature,nonRepudiation,keyEncipherment,keyCertSign
basicConstraints = critical,CA:true,pathlen:1

"""
        )

    with open(f"{cluster.ca_key_dir}/serial.txt", "w") as f:
        f.write("0001")

    with open(f"{cluster.ca_key_dir}/index.txt", "w") as f:
        f.write("")

    l.info("Created CA configuration files")


def generate_ca_key(cluster):
    l.debug("Generating CA private key")
    subprocess.check_call(
        ["openssl", "genrsa", "-out", cluster.ca_key_file, "4096"],
        stdout=subprocess.DEVNULL,
    )
    l.info("Generated CA private key")


def generate_ca_cert(cluster):
    l.debug("Generating CA certificate")
    subprocess.check_call(
        [
            "openssl",
            "req",
            "-new",
            "-x509",
            "-days",
            "3650",
            "-config",
            cluster.ca_conf,
            "-key",
            cluster.ca_key_file,
            "-out",
            cluster.ca_cert_file,
        ],
        stdout=subprocess.DEVNULL,
    )

    name = get_cert_display_name(cluster.ca_cert_file)

    l.info(f"Generated CA certificate '{name}'")


def generate_ca(cluster):
    generate_ca_config(cluster)
    generate_ca_key(cluster)
    generate_ca_cert(cluster)


def make_directories(cluster):

    l.debug("Creating directories")

    if not os.path.exists("workspace"):
        os.makedirs("workspace")

    os.mkdir(cluster.directory)
    os.mkdir(cluster.ca_key_dir)
    os.mkdir(cluster.master_certs_dir)
    os.mkdir(cluster.tserver_certs_dir)
    os.mkdir(cluster.client_certs_dir)
    os.mkdir(cluster.ca_pool_dir)

    l.info("Created directories")


def generate_clients(cluster):

    for client in cluster.clients:
        if cluster.is_client_ready(client):
            l.debug(f"Client '{client}' certificates already generated")
            continue
        generate_client_config(cluster, client)
        generate_client_key(cluster, client)
        generate_client_csr(cluster, client)
        generate_client_cert(cluster, client)


def generate_client_config(cluster, client):

    l.debug(f"Creating client '{client}' configuration file")

    with open(cluster.get_client_conf_file(client), "w") as f:
        f.write(
            f"""[ req ]
prompt=no
distinguished_name = my_distinguished_name

[ my_distinguished_name ]
organizationName = {cluster.organization}
commonName = client.{client}
"""
        )

    l.info(f"Created client '{client}' configuration file")


def generate_client_key(cluster, client):

    l.debug(f"Generating client '{client}' private key")

    subprocess.check_call(
        ["openssl", "genrsa", "-out", cluster.get_client_key_file(client), "4096"]
    )

    l.info(f"Generated client '{client}' private key")


def generate_client_csr(cluster, client):

    l.debug(f"Generating client '{client}' certificate request")

    subprocess.check_call(
        [
            "openssl",
            "req",
            "-new",
            "-config",
            cluster.get_client_conf_file(client),
            "-key",
            cluster.get_client_key_file(client),
            "-out",
            cluster.get_client_csr_file(client),
        ],
        stdout=subprocess.DEVNULL,
    )

    l.info(f"Generated client '{client}' certificate request")


def generate_client_cert(cluster, client):

    l.debug(f"Generating client '{client}' certificate")

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
            cluster.get_client_cert_file(client),
            "-outdir",
            cluster.client_certs_dir,
            "-in",
            cluster.get_client_csr_file(client),
            "-days",
            "3650",
            "-batch",
            "-extfile",
            cluster.get_client_conf_file(client),
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    name = get_cert_display_name(cluster.get_client_cert_file(client))

    l.info(f"Generated client '{client}' certificate '{name}'")


def do_init(cluster):
    """Initialize a new cluster"""

    l.info("Initialization of a new cluster")

    if cluster.is_ready:
        l.error("Cluster is already initialized, unable to continue")
        sys.exit(1)
    else:
        l.debug("Cluster is not already initialized, continuing")

    make_directories(cluster)
    generate_ca(cluster)
    generate_clients(cluster)

    do_generate_nodes(cluster)

    with open(cluster.ca_cert_file, "r") as f:
        do_add_cas(cluster, f.read())

    l.info(
        "The new cluster certificates are ready! Don't forget to 'apply' the configuration."
    )
