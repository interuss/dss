import os

from utils import slugify


class Cluster(object):
    """Represent an instance of a cluster, expose paths"""

    def __init__(self, name, cluster_context, namespace, organization, nodes_count):
        self._name = name
        self.cluster_context = cluster_context
        self.namespace = namespace
        self.organization = organization
        self.nodes_count = nodes_count

    @property
    def name(self):
        return slugify(self._name)

    @property
    def directory(self):
        # Replace characters breaking folder names
        def remove_special_chars(s: str):
            for c in [":", "/"]:
                s = s.replace(c, "_")
            return s

        return os.path.join(os.getcwd(), "workspace", remove_special_chars(self._name))

    @property
    def ca_key_dir(self):
        return os.path.join(self.directory, "ca")

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
    def client_certs_dir(self):
        return os.path.join(self.directory, "clients")

    @property
    def client_ca(self):
        return os.path.join(self.client_certs_dir, "root.crt")

    @property
    def master_certs_dir(self):
        return os.path.join(self.directory, "masters")

    @property
    def master_ca(self):
        return os.path.join(self.master_certs_dir, "ca.crt")

    @property
    def tserver_certs_dir(self):
        return os.path.join(self.directory, "tservers")

    @property
    def tserver_ca(self):
        return os.path.join(self.tserver_certs_dir, "ca.crt")

    @property
    def ca_pool_dir(self):
        return os.path.join(self.directory, "ca_pool")

    @property
    def ca_pool_ca(self):
        return os.path.join(self.ca_pool_dir, "ca.crt")

    @property
    def is_ready(self):
        return os.path.exists(self.ca_key_file)

    @property
    def clients(self):
        return ["yugabytedb"]  # TODO: Do we need more, like a specifc one for the DSS?

    def get_client_cert_file(self, client):
        return f"{self.client_certs_dir}/{client}.crt"

    def get_client_key_file(self, client):
        return f"{self.client_certs_dir}/{client}.key"

    def get_client_csr_file(self, client):
        return f"{self.ca_key_dir}/client.{client}.csr"

    def get_client_conf_file(self, client):
        return f"{self.ca_key_dir}/client.{client}.conf"

    def is_client_ready(self, client):
        return os.path.exists(self.get_client_cert_file(client))

    def get_node_short_name(self, node_type, node_id):
        return f"yb-{node_type}-{node_id}"

    def get_node_short_name_group(self, node_type, node_id):
        short_name = self.get_node_short_name(node_type, node_id)
        return f"{short_name}.yb-{node_type}s"

    def get_node_full_name(self, node_type, node_id):
        short_name_group = self.get_node_short_name_group(node_type, node_id)
        return f"{short_name_group}.{self.namespace}.svc.cluster.local"

    def get_node_full_name_without_group(self, node_type, node_id):
        short_name = self.get_node_short_name(node_type, node_id)
        return f"{short_name}.{self.namespace}.svc.cluster.local"

    def get_node_cert_file(self, node_type, node_id):
        folder = getattr(self, f"{node_type}_certs_dir")
        full_name = self.get_node_full_name(node_type, node_id)
        return f"{folder}/node.{full_name}.crt"

    def get_node_key_file(self, node_type, node_id):
        folder = getattr(self, f"{node_type}_certs_dir")
        full_name = self.get_node_full_name(node_type, node_id)
        return f"{folder}/node.{full_name}.key"

    def get_node_csr_file(self, node_type, node_id):
        full_name = self.get_node_full_name(node_type, node_id)
        return f"{self.ca_key_dir}/node.{full_name}.csr"

    def get_node_conf_file(self, node_type, node_id):
        full_name = self.get_node_full_name(node_type, node_id)
        return f"{self.ca_key_dir}/node.{full_name}.conf"

    def is_node_ready(self, node_type, node_id):
        return os.path.exists(self.get_node_cert_file(node_type, node_id))
