from typing import Optional

from monitoring.monitorlib.typing import ImplicitDict

from monitoring.deployment_manager.actions.test.configuration import Test


class KubernetesCluster(ImplicitDict):
    name: str


class DeploymentSpec(ImplicitDict):
    cluster: Optional[KubernetesCluster]
    test: Optional[Test] = None
