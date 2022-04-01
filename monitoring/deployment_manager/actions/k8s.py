from kubernetes.client import V1IngressClass

from monitoring.deployment_manager.infrastructure import deployment_action
from monitoring.deployment_manager.infrastructure import Context


@deployment_action('list_pods')
def list_pods(context: Context):
    """List all Kubernetes pods"""
    ret = context.clients.core.list_pod_for_all_namespaces(watch=False)
    msg = '\n'.join(['{}\t{}\t{}'.format(i.status.pod_ip, i.metadata.namespace, i.metadata.name) for i in ret.items])
    context.log.msg('Pods:\n' + msg)


@deployment_action('list_ingress_controllers')
def list_ingress_controllers(context: Context):
    """List all available ingress controllers"""
    class_list = context.clients.networking.list_ingress_class()
    msg = '\n'.join(['{}\t{}\t{}'.format(c.metadata.name, c.spec.controller, c.spec.parameters) for c in class_list.items])
    context.log.msg('Ingress controllers:\n' + msg)
