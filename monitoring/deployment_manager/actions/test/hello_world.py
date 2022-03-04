import time

import kubernetes
from kubernetes import client as k8s

from monitoring.deployment_manager import deploylib
import monitoring.deployment_manager.deploylib.deployments
import monitoring.deployment_manager.deploylib.namespaces
from monitoring.deployment_manager.infrastructure import deployment_action, Context


DEPLOYMENT_NAME = 'nginx-deployment'


def _define_namespace(name: str) -> k8s.V1Namespace:
    return k8s.V1Namespace(metadata=k8s.V1ObjectMeta(name=name))


def _define_nginx_deployment() -> kubernetes.client.V1Deployment:
    # Define Pod container template
    container = k8s.V1Container(
        name="nginx",
        image="nginx:1.15.4",
        ports=[k8s.V1ContainerPort(container_port=80)],
        resources=k8s.V1ResourceRequirements(
            requests={"cpu": "100m", "memory": "200Mi"},
            limits={"cpu": "500m", "memory": "500Mi"},
        ),
    )

    # Create and configure a spec section
    template = k8s.V1PodTemplateSpec(
        metadata=k8s.V1ObjectMeta(labels={"app": "nginx"}),
        spec=k8s.V1PodSpec(containers=[container]),
    )

    # Create the specification of deployment
    spec = k8s.V1DeploymentSpec(
        replicas=1, template=template, selector={
            "matchLabels":
                {"app": "nginx"}})

    # Instantiate the deployment object
    deployment = k8s.V1Deployment(
        api_version="apps/v1",
        kind="Deployment",
        metadata=k8s.V1ObjectMeta(name=DEPLOYMENT_NAME),
        spec=spec,
    )

    return deployment


def _define_nginx_service(name: str) -> kubernetes.client.V1Service:
    spec = k8s.V1ServiceSpec()
    svc = kubernetes.client.V1Service(
        api_version="core/v1",
        kind="Service",
        metadata=k8s.V1ObjectMeta(name=name),
        spec=spec,
    )

    return svc


@deployment_action('test/hello_world/deploy')
def deploy(context: Context) -> None:
    """Bring up the hello_world system"""
    namespace = _define_namespace(context.spec.test.v1.namespace)
    dep = _define_nginx_deployment()

    active_namespace = deploylib.namespaces.upsert(context.clients.core, context.log, namespace)
    deploylib.deployments.upsert(context.clients.apps, context.log, active_namespace, dep)


@deployment_action('test/hello_world/destroy')
def destroy(context: Context) -> None:
    """Tear down the hello_world system"""
    namespace = deploylib.namespaces.get(context.clients.core, context.log, _define_namespace(context.spec.test.v1.namespace))
    if namespace is None:
        context.log.warn('Namespace `{}` does not exist in `{}` cluster'.format(context.spec.test.v1.namespace, context.spec.cluster.name))
        return

    dep = _define_nginx_deployment()

    existing_dep = deploylib.deployments.get(context.clients.apps, context.log, namespace, dep)
    if existing_dep is None:
        context.log.warn('No existing deployment found in `{}` namespace of `{}` cluster'.format(namespace.metadata.name, context.spec.cluster.name))

    context.log.warn('Destroying hello_world system using `{}` namespace of `{}` cluster in 15 seconds...'.format(namespace.metadata.name, context.spec.cluster.name))
    if existing_dep is not None:
        time.sleep(15)
        context.log.msg('Deleting deployment')
        status = context.clients.apps.delete_namespaced_deployment(name=existing_dep.metadata.name, namespace=namespace.metadata.name)
        context.log.msg('Deployment deleted', message=status.message)

    context.log.msg('Deleting namespace')
    status = context.clients.core.delete_namespace(name=namespace.metadata.name)
    context.log.msg('Namespace deleted', name=status.message)
