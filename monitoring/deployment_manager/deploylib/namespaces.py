from typing import Optional

from kubernetes.client import CoreV1Api, V1Namespace
from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import comparisons


def get(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace) -> Optional[V1Namespace]:
    log.msg('Checking for existing namespace', name=namespace.metadata.name)
    namespace_list = client.list_namespace()
    matching_namespaces = [n for n in namespace_list.items
                           if n.metadata.name == namespace.metadata.name]
    if len(matching_namespaces) > 1:
        raise ValueError('Found {} namespaces matching `{}`'.format(len(matching_namespaces), namespace.metadata.name))
    if not matching_namespaces:
        return None
    return matching_namespaces[0]


def upsert(client: CoreV1Api, log: BoundLogger, namespace: V1Namespace) -> V1Namespace:
    old_namespace = get(client, log, namespace)
    if old_namespace is not None:
        if comparisons.specs_are_the_same(old_namespace, namespace):
            log.msg('Existing namespace does not need to be updated', name=old_namespace.metadata.name)
            new_namespace = old_namespace
        else:
            log.msg('Updating existing namespace')
            new_namespace = client.patch_namespace(namespace.metadata.name, body=namespace)
            log.msg('Namespace updated', name=new_namespace.metadata.name)
    else:
        log.msg('Creating namespace `{}`'.format(namespace.metadata.name))
        new_namespace = client.create_namespace(body=namespace)
        log.msg('Created namespace', name=new_namespace.metadata.name)
    return new_namespace
