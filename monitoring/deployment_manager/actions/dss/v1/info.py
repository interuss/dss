import base64

from kubernetes.client import V1Secret

from monitoring.deployment_manager.infrastructure import deployment_action, Context

import cryptography.exceptions
import cryptography.hazmat.backends
import cryptography.hazmat.primitives.hashes
import cryptography.hazmat.primitives.serialization
import cryptography.x509
import pem


def _public_key_bytes(public_key) -> bytes:
    return public_key.public_bytes(
        cryptography.hazmat.primitives.serialization.Encoding.PEM,
        cryptography.hazmat.primitives.serialization.PublicFormat.SubjectPublicKeyInfo)


@deployment_action('dss/info/print_ca_public_certs')
def print_ca_public_certs(context: Context):
    if 'dss' not in context.spec:
        raise ValueError('DSS system is not defined in deployment configuration')
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
