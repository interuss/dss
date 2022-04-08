from typing import Any, List

from kubernetes import client as k8s


DEPLOYMENT_NAME = 'webserver-deployment'
CONTAINER_PORT = 8001
APP_NAME = 'webbserver-app'

SERVICE_NAME = 'webserver-service'
SERVICE_PORT = 8002
SERVICE_PORT_NAME = 'webserver-port'

INGRESS_NAME = 'webserver-ingress'


def define_namespace(name: str) -> k8s.V1Namespace:
    return k8s.V1Namespace(metadata=k8s.V1ObjectMeta(name=name))


def _define_webserver_deployment() -> k8s.V1Deployment:
    # Define Pod container template
    container = k8s.V1Container(
        name='echo',
        image='hashicorp/http-echo',
        command=['/http-echo', '-listen=:{}'.format(CONTAINER_PORT), '-text="Echo server on port {}"'.format(CONTAINER_PORT)],
        ports=[k8s.V1ContainerPort(container_port=CONTAINER_PORT)],
        resources=k8s.V1ResourceRequirements(
            requests={'cpu': '100m', 'memory': '200Mi'},
            limits={'cpu': '500m', 'memory': '500Mi'},
        ),
    )

    # Create and configure a spec section
    template = k8s.V1PodTemplateSpec(
        metadata=k8s.V1ObjectMeta(labels={'app': APP_NAME}),
        spec=k8s.V1PodSpec(containers=[container]),
    )

    # Create the specification of deployment
    spec = k8s.V1DeploymentSpec(
        replicas=1, template=template, selector=k8s.V1LabelSelector(match_labels={'app': APP_NAME}))

    # Instantiate the deployment object
    deployment = k8s.V1Deployment(
        metadata=k8s.V1ObjectMeta(name=DEPLOYMENT_NAME),
        spec=spec,
    )

    return deployment


def _define_webserver_service() -> k8s.V1Service:
    spec = k8s.V1ServiceSpec(
        selector={'app': APP_NAME},
        type='LoadBalancer',
        ports=[
            k8s.V1ServicePort(
                name=SERVICE_PORT_NAME,
                port=SERVICE_PORT,
                target_port=CONTAINER_PORT,
            )
        ]
    )

    svc = k8s.V1Service(
        metadata=k8s.V1ObjectMeta(name=SERVICE_NAME),
        spec=spec,
    )

    return svc


def _define_webserver_ingress() -> k8s.V1Ingress:
    spec = k8s.V1IngressSpec(
        rules=[
            k8s.V1IngressRule(
                http=k8s.V1HTTPIngressRuleValue(
                    paths=[
                        k8s.V1HTTPIngressPath(
                            path='/',
                            path_type='Prefix',
                            backend=k8s.V1IngressBackend(
                                service=k8s.V1IngressServiceBackend(
                                    name=SERVICE_NAME,
                                    port=k8s.V1ServiceBackendPort(name=SERVICE_PORT_NAME))))
                    ]
                )
            )
        ]
    )

    ingress = k8s.V1Ingress(
        metadata=k8s.V1ObjectMeta(name=INGRESS_NAME),
        spec=spec,
    )

    return ingress


def define_resources() -> List[Any]:
    return [
        _define_webserver_deployment(),
        _define_webserver_service(),
        _define_webserver_ingress(),
    ]
