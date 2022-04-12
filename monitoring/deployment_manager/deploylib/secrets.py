from typing import Optional

from kubernetes.client import CoreV1Api, V1Secret, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import common_k8s


def get(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace, name: str) -> Optional[V1Secret]:
    return common_k8s.get_resource(
        lambda: client.list_namespaced_secret(namespace=namespace.metadata.name),
        log, 'secret', name)


def upsert(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace, secret: V1Secret) -> V1Secret:
    existing_secret = get(client, log, namespace, secret.metadata.name)
    return common_k8s.upsert_resource(
        existing_secret, secret, log, 'secret',
        lambda: client.create_namespaced_secret(
            body=secret, namespace=namespace.metadata.name),
        lambda: client.patch_namespaced_secret(
            existing_secret.metadata.name, namespace.metadata.name, secret))
