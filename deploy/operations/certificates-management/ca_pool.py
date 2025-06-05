import base64
import hashlib
import logging
import os
import re
import shutil
import tempfile

from utils import get_cert_display_name, get_cert_serial

l = logging.getLogger(__name__)


def build_pool_hash(cluster):

    CAs = []
    for f in os.listdir(cluster.ca_pool_dir):

        if f.endswith(".crt") and f != "ca.crt":
            CAs.append(f.lower())

    CAs = sorted(CAs)

    h = hashlib.sha256()
    h.update((",".join(CAs)).encode("utf-8"))

    # Create an hash without special chars (replaced by 'Aa')
    hashed = base64.b64encode(h.digest(), b"Aa").decode("utf-8")

    return f"{hashed[:5]}-{hashed[-10:-5]}"


def add_cas(cluster, certificate):

    folder = cluster.ca_pool_dir

    l.debug("Getting new CA metadata")

    with tempfile.NamedTemporaryFile(delete_on_close=False) as tf:
        tf.write(certificate.encode("utf-8"))
        tf.close()

        serial = get_cert_serial(tf.name)
        name = get_cert_display_name(tf.name)

        filename = f"{serial}.crt"

        target_file = os.path.join(folder, filename)

        if os.path.exists(target_file):
            l.info(f"CA {name} already present in the pool")
            return

        l.info(f"Adding CA {name} in the pool")

        with open(target_file, "w") as f:
            f.write(certificate)


def regenerate_ca_files(cluster):

    l.debug("Regenerating CA files from all CA in the pool")

    CAs = []
    for filename in os.listdir(cluster.ca_pool_dir):

        if filename.endswith(".crt") and filename != "ca.crt":
            with open(os.path.join(cluster.ca_pool_dir, filename), "r") as f:
                CAs.append(f.read())

    CAs = sorted(CAs)

    with open(cluster.ca_pool_ca, "w") as f:
        f.write("\n\n".join(CAs))

    shutil.copy(cluster.ca_pool_ca, cluster.client_ca)
    shutil.copy(cluster.ca_cert_file, cluster.client_instance_ca)

    for node_type in ["master", "tserver"]:
        shutil.copy(cluster.ca_pool_ca, getattr(cluster, f"{node_type}_ca"))

    h = build_pool_hash(cluster)

    l.info(f"Regenerated CA files from the CA pool. Current pool hash: {h}")


def do_add_cas(cluster, certificates):
    pattern = re.compile(
        r"-----BEGIN CERTIFICATE-----\s*.+?\s*-----END CERTIFICATE-----", re.DOTALL
    )
    for cert in pattern.findall(certificates):
        add_cas(cluster, cert)

    regenerate_ca_files(cluster)


def do_remove_cas(cluster, certificates_or_serial):
    pattern = re.compile(
        r"-----BEGIN CERTIFICATE-----\s*.+?\s*-----END CERTIFICATE-----", re.DOTALL
    )
    for cert in pattern.findall(certificates_or_serial):
        with tempfile.NamedTemporaryFile(delete_on_close=False) as tf:
            tf.write(cert.encode("utf-8"))
            tf.close()
            serial = get_cert_serial(tf.name)
            name = get_cert_display_name(tf.name)

            filename = f"{serial}.crt"

            target = os.path.join(cluster.ca_pool_dir, filename)

            if os.path.isfile(target):
                os.unlink(target)
                l.info(f"Removed certificate {name}")
            else:
                l.info(f"Certificate {name} not present in pool")

    for filename in sorted(os.listdir(cluster.ca_pool_dir)):
        if filename.endswith(".crt") and filename != "ca.crt":

            serial = get_cert_serial(os.path.join(cluster.ca_pool_dir, filename))
            name = get_cert_display_name(os.path.join(cluster.ca_pool_dir, filename))

            if certificates_or_serial == name or certificates_or_serial == serial or f"SN={certificates_or_serial}, " in name or name.startswith(certificates_or_serial):
                os.unlink(os.path.join(cluster.ca_pool_dir, filename))
                l.info(f"Removed certificate {name}")

    regenerate_ca_files(cluster)

def do_get_ca(cluster):
    with open(cluster.ca_cert_file, "r") as f:
        print(f.read())

def do_get_pool_ca(cluster):
    with open(cluster.ca_pool_ca, "r") as f:
        print(f.read())

def do_list_pool_ca(cluster):

    h = build_pool_hash(cluster)

    print(f"Current CA pool hash: {h}")

    for filename in sorted(os.listdir(cluster.ca_pool_dir)):
        if filename.endswith(".crt") and filename != "ca.crt":
            print(get_cert_display_name(os.path.join(cluster.ca_pool_dir, filename)))
