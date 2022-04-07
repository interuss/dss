import time

from monitoring.deployment_manager import deploylib
import monitoring.deployment_manager.deploylib.namespaces
import monitoring.deployment_manager.deploylib.systems
from monitoring.deployment_manager.infrastructure import deployment_action, Context
from monitoring.deployment_manager.systems.test import hello_world


@deployment_action('test/hello_world/deploy')
def deploy(context: Context) -> None:
    """Bring up the hello_world system"""
    namespace = hello_world.define_namespace(context.spec.test.v1.namespace)
    resources = hello_world.define_resources()

    active_namespace = deploylib.namespaces.upsert(context.clients.core, context.log, namespace)
    deploylib.systems.upsert_resources(resources, active_namespace, context.clients, context.log)

    context.log.msg('hello_world system deployment complete.  Run `minikube tunnel` and then navigate to http://localhost in a browser to see the web content.')


@deployment_action('test/hello_world/destroy')
def destroy(context: Context) -> None:
    """Tear down the hello_world system"""
    namespace = deploylib.namespaces.get(context.clients.core, context.log, hello_world.define_namespace(context.spec.test.v1.namespace))
    if namespace is None:
        context.log.warn('Namespace `{}` does not exist in `{}` cluster'.format(context.spec.test.v1.namespace, context.spec.cluster.name))
        return

    resources = hello_world.define_resources()
    resources.reverse()
    existing_resources = deploylib.systems.get_resources(resources, namespace, context.clients, context.log, context.spec.cluster.name)

    context.log.warn('Destroying hello_world system in `{}` namespace of `{}` cluster in 15 seconds...'.format(namespace.metadata.name, context.spec.cluster.name))
    time.sleep(15)

    deploylib.systems.delete_resources(existing_resources, namespace, context.clients, context.log)

    context.log.msg('Deleting namespace')
    status = context.clients.core.delete_namespace(name=namespace.metadata.name)
    context.log.msg('Namespace deleted', name=status.message)

    context.log.msg('hello_world system removal complete')
