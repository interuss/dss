from dataclasses import dataclass
from typing import Callable, Dict, Optional

import kubernetes
import structlog

from monitoring.deployment_manager.deployment_spec import DeploymentSpec


@dataclass
class Clients(object):
    core: kubernetes.client.CoreV1Api
    apps: kubernetes.client.AppsV1Api
    networking: kubernetes.client.NetworkingV1Api


@dataclass
class Context(object):
    spec: DeploymentSpec
    log: structlog.BoundLogger
    clients: Optional[Clients]


def make_context(spec: DeploymentSpec):
    if spec.cluster is not None:
        contexts, active_context = kubernetes.config.list_kube_config_contexts()
        if not contexts:
            raise ValueError('Cannot find any context in kube-config file.')
        matching_contexts = [c['name'] for c in contexts if c['name'] == spec.cluster.name]
        if not matching_contexts:
            raise ValueError('Cannot find definition for context `{}` in kube-config file'.format(spec.cluster.name))
        if len(matching_contexts) > 1:
            raise ValueError('Found multiple context definitions with the name `{}` in kube-config file'.format(spec.cluster.name))

        api_client = kubernetes.config.new_client_from_config(context=matching_contexts[0])
        clients = Clients(
            core=kubernetes.client.CoreV1Api(api_client=api_client),
            apps=kubernetes.client.AppsV1Api(api_client=api_client),
            networking=kubernetes.client.NetworkingV1Api(api_client=api_client))
    else:
        clients = None

    log = structlog.get_logger()

    return Context(spec=spec, log=log, clients=clients)


actions: Dict[str, Callable[[Context], None]] = {}


def deployment_action(name: str):
    def decorator_declare_action(action: Callable[[Context], None]):
        global actions
        actions[name] = action
        return action
    return decorator_declare_action
