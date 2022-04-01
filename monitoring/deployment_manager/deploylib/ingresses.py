from typing import Optional

from kubernetes.client import NetworkingV1Api, V1Ingress, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import common_k8s


def get(client: NetworkingV1Api, log: BoundLogger, namespace: V1Namespace, ingress: V1Ingress) -> Optional[V1Ingress]:
    return common_k8s.get_resource(
        lambda: client.list_namespaced_ingress(namespace=namespace.metadata.name),
        log, 'ingress', ingress.metadata.name)


def upsert(client: NetworkingV1Api, log: BoundLogger, namespace: V1Namespace, ingress: V1Ingress) -> V1Ingress:
    existing_ingress = get(client, log, namespace, ingress)
    return common_k8s.upsert_resource(
        existing_ingress, ingress, log, 'ingress',
        lambda: client.create_namespaced_ingress(
            body=ingress, namespace=namespace.metadata.name),
        lambda: client.patch_namespaced_ingress(
            existing_ingress.metadata.name, namespace.metadata.name, ingress))
