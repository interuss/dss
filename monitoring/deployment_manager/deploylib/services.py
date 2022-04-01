from typing import Optional

from kubernetes.client import CoreV1Api, V1Service, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import common_k8s


def get(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace, svc: V1Service) -> Optional[V1Service]:
    return common_k8s.get_resource(
        lambda: client.list_namespaced_service(namespace=namespace.metadata.name),
        log, 'service', svc.metadata.name)


def upsert(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace, svc: V1Service) -> V1Service:
    existing_svc = get(client, log, namespace, svc)
    return common_k8s.upsert_resource(
        existing_svc, svc, log, 'service',
        lambda: client.create_namespaced_service(
            body=svc, namespace=namespace.metadata.name),
        lambda: client.patch_namespaced_service(
            existing_svc.metadata.name, namespace.metadata.name, svc))
