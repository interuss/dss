# Deploy a DSS instance locally on Minikube

This section provide instructions to prepare a local minikube cluster.

Minikube is going to take care of most of the work by spawning a local kubernetes cluster.

## Getting started

### Prerequisites

In addition to the [general infrastructure prerequisites](./index.md#prerequisites), download & install the following tools to your workstation:

1. Install [minikube](https://minikube.sigs.k8s.io/docs/start/) (First step only).

### Create a new minikube cluster

1. Run `minikube start -p dss-local-cluster` to create a new cluster.
2. Run `minikube tunnel -p dss-local-cluster` and keep it running to expose LoadBalancer services.

If needed, you can change the name of the cluster (`dss-local-cluster` in this documentation) as needed. You may also deploy multiple cluster at the same time, using different names.

### Access to the cluster

Minikube provide a UI, should you want to keep track of deployment and/or inspect the cluster. To start it, use the following command:

1. `minikube dashboard -p dss-local-cluster`

You can also use any other tool as needed. You can switch to the cluster's context by using the following command:

1. `kubectl config use-context dss-local-cluster`

## Next steps

[Deploy services](../services/to-minikube.md)
