from typing import Any, List

from kubernetes import client as k8s


DEPLOYMENT_NAME = 'nginx-deployment'
WEBSERVER_PORT = 'webserver-port'
APP_NAME = 'webbserver-app'


def define_namespace(name: str) -> k8s.V1Namespace:
    return k8s.V1Namespace(metadata=k8s.V1ObjectMeta(name=name))


def _define_webserver_deployment() -> k8s.V1Deployment:
    # Define Pod container template
    container = k8s.V1Container(
        name='nginx',
        image='nginx:1.15.4',
        ports=[k8s.V1ContainerPort(container_port=80)],
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
        metadata=k8s.V1ObjectMeta(name=DEPLOYMENT_NAME, labels={'app': APP_NAME}),
        spec=spec,
    )

    return deployment


def _define_webserver_service() -> k8s.V1Service:
    spec = k8s.V1ServiceSpec(
        selector={'app': APP_NAME},
        type='NodePort',
        ports=[
            k8s.V1ServicePort(
                name=WEBSERVER_PORT,
                port=8080,
                target_port=80
            )
        ]
    )

    svc = k8s.V1Service(
        metadata=k8s.V1ObjectMeta(name='webserver-service'),
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
                                    name='webserver-service',
                                    port=k8s.V1ServiceBackendPort(name=WEBSERVER_PORT))))
                    ]
                )
            )
        ]
    )

    ingress = k8s.V1Ingress(
        metadata=k8s.V1ObjectMeta(name='webserver-ingress'),
        spec=spec,
    )

    return ingress


def define_resources() -> List[Any]:
    return [
        _define_webserver_deployment(),
        _define_webserver_service(),
        #_define_webserver_ingress(),
    ]
