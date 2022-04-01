from typing import Optional

from kubernetes.client import CoreV1Api, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import common_k8s


def get(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace) -> Optional[V1Namespace]:
    return common_k8s.get_resource(
        lambda: client.list_namespace(), log, 'namespace', namespace.metadata.name)


def upsert(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace) -> V1Namespace:
    return common_k8s.upsert_resource(
        get(client, log, namespace), namespace, log, 'namespace',
        lambda: client.create_namespace(body=namespace),
        lambda: client.patch_namespace(namespace.metadata.name, body=namespace))
