from typing import Optional

from implicitdict import ImplicitDict

from monitoring.deployment_manager.systems.dss.configuration import DSS
from monitoring.deployment_manager.systems.test.configuration import Test


class KubernetesCluster(ImplicitDict):
    name: str
    """Name of the Kubernetes cluster containing this deployment.

    Contained in the NAME column of the response to
    `kubectl config get-contexts`.
    """


class DeploymentSpec(ImplicitDict):
    cluster: Optional[KubernetesCluster]
    """Definition of Kubernetes cluster containing this deployment."""

    test: Optional[Test]
    """Test systems in this deployment."""

    dss: Optional[DSS]
    """DSS instance in this deployment."""
