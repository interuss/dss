#!/usr/bin/env python3

import argparse
import os
import shutil
import subprocess


class YugabyteCluster(object):

    def __init__(self, cluster_context, namespace, ca_cert_to_join=None):
        self._ca_cert_to_join = ca_cert_to_join
        self._cluster_context = cluster_context
        self._namespace = namespace

    @property
    def ca_cert_to_join(self):
        return self._ca_cert_to_join

    @property
    def namespace(self):
        return self._namespace

    @property
    def directory(self):
        # Replace characters breaking folder names
        def remove_special_chars(s: str):
            for c in [":", "/"]:
                s = s.replace(c, "_")
            return s

        return os.path.join(
            os.getcwd(),
            "workspace-yugabyte", remove_special_chars(self._cluster_context)
        )

    @property
    def ca_key_dir(self):
        return os.path.join(self.directory, "ca_key_dir")

    @property
    def ca_key_file(self):
        return os.path.join(self.ca_key_dir, "ca.key")

    @property
    def ca_cert_file(self):
        return os.path.join(self.ca_key_dir, "ca.crt")

    @property
    def ca_conf(self):
        return os.path.join(self.ca_key_dir, "ca.conf")

    @property
    def ca_certs_dir(self):
        return os.path.join(self.directory, "ca_certs_dir")

    @property
    def client_certs_dir(self):
        return os.path.join(self.directory, "client_certs_dir")

    @property
    def master_certs_dir(self):
        return os.path.join(self.directory, "master_certs_dir")

    @property
    def tserver_certs_dir(self):
        return os.path.join(self.directory, "tserver_certs_dir")


def parse_args():
    parser = argparse.ArgumentParser(
        description="Creates certificates for a new Cockroachdb cluster"
    )
    parser.add_argument(
        "--cluster-context",
        metavar="CLUSTER_CONTEXT",
        required=True,
        help="kubernetes cluster context name",
    )
    parser.add_argument(
        "--namespace",
        metavar="NAMESPACE",
        required=True,
        help="kubernetes cluster namespace you are deploying to.",
    )
    # TODO
    # parser.add_argument(
    #     "--node-address",
    #     metavar="ADDRESS",
    #     nargs="*",
    #     default=[],
    #     help="extra addresses to add to the node certificate",
    # )
    parser.add_argument(
        "--ca-cert-to-join",
        metavar="FILENAME",
        help="file containing an existing CA cert of a cluster to join.",
    )
    parser.add_argument(
        "--overwrite-ca-cert",
        action="store_true",
        default=False,
        help="True to generate new CA certs, false to use the existing one",
    )
    parser.add_argument(
        "--node-count",
        metavar="NODE_COUNT",
        default="3",
        help="Number of yugabyte nodes in the cluster, default to 3",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    cr = YugabyteCluster(args.cluster_context, args.namespace, args.ca_cert_to_join)

    # Create the generated directories.
    if not os.path.exists("workspace-yugabyte"):
        os.makedirs("workspace-yugabyte")

    if not os.path.exists(cr.directory):
        os.makedirs(cr.directory)

    create_ca = not os.path.exists(cr.ca_key_dir) or args.overwrite_ca_cert
    if create_ca:
        # Create a new CA.
        # Delete and recreate the ca_certs_dir.
        shutil.rmtree(cr.ca_key_dir, ignore_errors=True)
        os.mkdir(cr.ca_key_dir)

        # Build master, tserver and client certs.
        os.mkdir(cr.master_certs_dir)
        os.mkdir(cr.tserver_certs_dir)
        os.mkdir(cr.client_certs_dir)

    if create_ca:

        with open(cr.ca_conf, "w") as f:
            f.write(
                f"""
        [ ca ]
        default_ca = my_ca

[ my_ca ]
default_days = 3650

serial = {cr.ca_key_dir}/serial.txt
database = {cr.ca_key_dir}/index.txt
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
organizationName = Yugabyte
commonName = CA for YugabyteDB

[ my_extensions ]
keyUsage = critical,digitalSignature,nonRepudiation,keyEncipherment,keyCertSign
basicConstraints = critical,CA:true,pathlen:1

"""
            )

        with open(f"{cr.ca_key_dir}/serial.txt", "w") as f:
            f.write("01")

        with open(f"{cr.ca_key_dir}/index.txt", "w") as f:
            f.write("")

        subprocess.check_call(["openssl", "genrsa", "-out", cr.ca_key_file])

        subprocess.check_call(
            [
                "openssl",
                "req",
                "-new",
                "-x509",
                "-days",
                "3650",
                "-config",
                cr.ca_conf,
                "-key",
                cr.ca_key_file,
                "-out",
                cr.ca_cert_file,
            ]
        )

    ### CLIENT CERTIFICATE

    # Copy CA
    shutil.copy(cr.ca_cert_file, f"{cr.client_certs_dir}/root.crt")

    for client in ["yugabytedb"]:

        crt_name = f"{cr.client_certs_dir}/{client}.crt"
        key_name = f"{cr.client_certs_dir}/{client}.key"

        conf_name = f"{cr.ca_key_dir}/conf.client.{client}"
        csr_name = f"{cr.ca_key_dir}/csr.client.{client}"

        if os.path.exists(crt_name) and os.path.exists(key_name):  # No need to regenerate it
            continue

        with open(conf_name, "w") as f:
            f.write(
                f"""[ req ]
prompt=no
distinguished_name = my_distinguished_name

[ my_distinguished_name ]
organizationName = Yugabyte
commonName = {client}
"""
            )

        subprocess.check_call(["openssl", "genrsa", "-out", key_name])

        subprocess.check_call(
            [
                "openssl",
                "req",
                "-new",
                "-config",
                conf_name,
                "-key",
                key_name,
                "-out",
                csr_name,
            ]
        )

        subprocess.check_call(
            [
                "openssl",
                "ca",
                "-config",
                cr.ca_conf,
                "-keyfile",
                cr.ca_key_file,
                "-cert",
                cr.ca_cert_file,
                "-policy",
                "my_policy",
                "-out",
                crt_name,
                "-outdir",
                cr.client_certs_dir,
                "-in",
                csr_name,
                "-days",
                "3650",
                "-batch",
                "-extfile",
                conf_name,
            ]
        )

    ### SERVERS

    for server_type in ["master", "tserver"]:

        folder = getattr(cr, f"{server_type}_certs_dir")

        # Copy CA
        shutil.copy(cr.ca_cert_file, folder)

        for server_id in range(0, int(args.node_count)):
            short_name = f"yb-{server_type}-{server_id}"
            short_name_group = f"{short_name}.yb-{server_type}s"
            full_name_group = f"{short_name}.{cr.namespace}.svc.cluster.local"
            full_name = f"{short_name_group}.{cr.namespace}.svc.cluster.local"

            crt_name = f"{folder}/node.{full_name}.crt"
            key_name = f"{folder}/node.{full_name}.key"

            conf_name = f"{cr.ca_key_dir}/conf.{full_name}"
            csr_name = f"{cr.ca_key_dir}/csr.{full_name}"

            if os.path.exists(crt_name) and os.path.exists(key_name):  # No need to regenerate it
                continue

            with open(conf_name, "w") as f:
                f.write(
                    f"""[ req ]
prompt=no
distinguished_name = my_distinguished_name

[ my_distinguished_name ]
organizationName = Yugabyte
commonName = {full_name}

# Multiple subject alternative names (SANs) such as IP Address,
# DNS Name, Email, URI, and so on, can be specified under this section
[ req_ext]
subjectAltName = @alt_names
[alt_names]
DNS.1 = {short_name}
DNS.2 = {full_name}
DNS.3 = {short_name_group}
DNS.4 = {full_name_group}
DNS.5 = yb-{server_type}s
DNS.6 = yb-{server_type}s.{cr.namespace}
DNS.7 = yb-{server_type}s.{cr.namespace}.svc.cluster.local
"""
                )

            subprocess.check_call(["openssl", "genrsa", "-out", key_name])

            subprocess.check_call(
                [
                    "openssl",
                    "req",
                    "-new",
                    "-config",
                    conf_name,
                    "-key",
                    key_name,
                    "-out",
                    csr_name,
                ]
            )

            subprocess.check_call(
                [
                    "openssl",
                    "ca",
                    "-config",
                    cr.ca_conf,
                    "-keyfile",
                    cr.ca_key_file,
                    "-cert",
                    cr.ca_cert_file,
                    "-policy",
                    "my_policy",
                    "-out",
                    crt_name,
                    "-outdir",
                    folder,
                    "-in",
                    csr_name,
                    "-days",
                    "3650",
                    "-batch",
                    "-extfile",
                    conf_name,
                    "-extensions",
                    "req_ext",
                ]
            )


if __name__ == "__main__":
    main()
