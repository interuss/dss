from typing import Optional

from kubernetes.client import AppsV1Api, V1Deployment, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import common_k8s


def get(client: AppsV1Api, log: BoundLogger, namespace: V1Namespace, dep: V1Deployment) -> Optional[V1Deployment]:
    return common_k8s.get_resource(
        lambda: client.list_namespaced_deployment(namespace=namespace.metadata.name),
        log, 'deployment', dep.metadata.name)


def upsert(client: AppsV1Api, log: BoundLogger, namespace: V1Namespace, dep: V1Deployment) -> V1Deployment:
    existing_dep = get(client, log, namespace, dep)
    return common_k8s.upsert_resource(
        existing_dep, dep, log, 'deployment',
        lambda: client.create_namespaced_deployment(
            body=dep, namespace=namespace.metadata.name),
        lambda: client.patch_namespaced_deployment(
            existing_dep.metadata.name, namespace.metadata.name, dep))
