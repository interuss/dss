from monitoring.deployment_manager.infrastructure import deployment_action
from monitoring.deployment_manager.infrastructure import Context


@deployment_action('list_pods')
def list_pods(context: Context):
    """List all Kubernetes pods"""
    ret = context.clients.core.list_pod_for_all_namespaces(watch=False)
    msg = '\n'.join(['{}\t{}\t{}'.format(i.status.pod_ip, i.metadata.namespace, i.metadata.name) for i in ret.items])
    context.log.msg('Pods:\n' + msg)
