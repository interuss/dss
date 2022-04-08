import base64
import os

from kubernetes.client import V1Secret
import pem
import yaml

from monitoring.deployment_manager.deploylib.crdb_cluster_api import ClusterAPI
from monitoring.deployment_manager.deploylib.port_forwarding import get_requests_session_for_pod
from monitoring.deployment_manager.infrastructure import deployment_action, Context
from monitoring.deployment_manager.actions.dss.v1.common import requires_v1_dss

import cryptography.exceptions
import cryptography.hazmat.backends
import cryptography.hazmat.primitives.hashes
import cryptography.hazmat.primitives.serialization
import cryptography.x509


def _public_key_bytes(public_key) -> bytes:
    return public_key.public_bytes(
        cryptography.hazmat.primitives.serialization.Encoding.PEM,
        cryptography.hazmat.primitives.serialization.PublicFormat.SubjectPublicKeyInfo)


@deployment_action('dss/info/print_ca_public_certs')
@requires_v1_dss
def print_ca_public_certs(context: Context):
    """Print the accepted CA certificates accepted by the cluster."""
    backend = cryptography.hazmat.backends.default_backend()

    # Read the list of accepted CA certificates
    ca_crt: V1Secret = context.clients.core.read_namespaced_secret(name='cockroachdb.ca.crt', namespace=context.spec.dss.v1.namespace)
    ca_crt_content = base64.b64decode(ca_crt.data['ca.crt'])
    certs = [cryptography.x509.load_pem_x509_certificate(str(c).encode('utf-8'), backend)
             for c in pem.parse(ca_crt_content)]
    context.log.msg('ca.crt:\n' + ca_crt_content.decode('utf-8'))

    # Read the private key to determine which certificate belongs to this instance
    ca_key: V1Secret = context.clients.core.read_namespaced_secret(name='cockroachdb.ca.key', namespace=context.spec.dss.v1.namespace)
    ca_key_content = base64.b64decode(ca_key.data['ca.key']).decode('utf-8')
    private_key = cryptography.hazmat.primitives.serialization.load_pem_private_key(
        ca_key_content.encode('utf-8'), password=None, backend=backend)
    public_key_bytes = _public_key_bytes(private_key.public_key())

    matches = [i for i in range(len(certs))
               if public_key_bytes == _public_key_bytes(certs[i].public_key())]
    match_words = ['first', 'second', 'third'] + ['{}th'.format(i) for i in range(4, 50)]
    if not matches:
        context.log.warn('This DSS instance\'s public key does not appear in any of the listed certificates')
    else:
        context.log.msg('This DSS instance\'s public key matches the {} certificate{} in those above'.format(', '.join(match_words[m] for m in matches), 's' if len(matches) > 1 else ''))


@deployment_action('dss/crdb/status')
@requires_v1_dss
def crdb_status(context: Context):
    """Retrieve and print information about the CockroachDB cluster.

    The following environment variables must be set to describe a CRDB user with
    admin priviledges: CRDB_WEBVIEWER_USERNAME,  CRDB_WEBVIEWER_PASSWORD
    """
    username = os.environ.get('CRDB_WEBVIEWER_USERNAME', None)
    password = os.environ.get('CRDB_WEBVIEWER_PASSWORD', None)
    if username is None or password is None:
        raise ValueError('Environment variables CRDB_WEBVIEWER_USERNAME and CRDB_WEBVIEWER_PASSWORD must be set; see usage help')
    pod_name = 'cockroachdb-0'
    pod_session, host_port = get_requests_session_for_pod(pod_name, context.spec.dss.v1.namespace, 8080, context.clients.core)
    cluster = ClusterAPI(pod_session, base_url='https://{}/api/v2'.format(host_port), username=username, password=password)
    up = cluster.is_up()
    ready = cluster.is_ready()
    source = '{} ({}, {})'.format(pod_name, 'up' if up else 'DOWN', 'ready' if ready else 'NOT READY')
    if up and ready:
        nodes = cluster.get_nodes()
        summary = dict()
        for n in nodes:
            k, v = n.summarize()
            summary[k] = v
        context.log.msg('{} reports:\n'.format(source) + yaml.dump(summary))
    else:
        context.log.msg('{} not ready to query nodes'.format(source))
