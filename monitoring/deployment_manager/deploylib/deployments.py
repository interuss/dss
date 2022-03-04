from typing import Optional

from kubernetes.client import AppsV1Api, V1Deployment, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import comparisons


def get(client: AppsV1Api, log: BoundLogger, namespace: V1Namespace, dep: V1Deployment) -> Optional[V1Deployment]:
    log.msg('Checking for existing deployment', name=dep.metadata.name)
    deployment_list = client.list_namespaced_deployment(namespace=namespace.metadata.name)
    matching_deployments = [d for d in deployment_list.items
                            if d.metadata.name == dep.metadata.name]
    if len(matching_deployments) > 2:
        raise ValueError('Found {} deployments matching `{}`'.format(len(matching_deployments), dep.metadata.name))
    if not matching_deployments:
        return None
    return matching_deployments[0]


def upsert(client: AppsV1Api, log: BoundLogger, namespace: V1Namespace, dep: V1Deployment) -> V1Deployment:
    existing_dep = get(client, log, namespace, dep)
    if existing_dep is not None:
        if comparisons.specs_are_the_same(existing_dep, dep):
            log.msg('Existing deployment does not need to be updated', name=existing_dep.metadata.name)
            new_dep = existing_dep
        else:
            log.msg('Updating existing deployment')
            new_dep = client.patch_namespaced_deployment(
                existing_dep.metadata.name, namespace.metadata.name, dep)
            log.msg('Deployment updated', name=new_dep.metadata.name)
    else:
        log.msg('Creating new deployment')
        new_dep = client.create_namespaced_deployment(
            body=dep, namespace=namespace.metadata.name)
        log.msg('Deployment created', name=new_dep.metadata.name)
    return new_dep
