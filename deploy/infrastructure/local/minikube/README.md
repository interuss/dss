# minikube

This module provide instructions to prepare a local minikube cluster.

Minikube is going to take care of most of the work by spawning a local kubernetes cluster.

## Getting started

### Prerequisites

Download & install the following tools to your workstation:

1. Install [minikube](https://minikube.sigs.k8s.io/docs/start/) (First step only).
2. Install tools from [Prerequisites](../../../../build/README.md)

### Create a new minikube cluster

1. Run `minikube start -p dss-local-cluster` to create a new cluster.
2. Run `minikube tunnel -p dss-local-cluster` and keep it running to expose LoadBalancer services.

If needed, you can change the name of the cluster (`dss-local-cluster` in this documentation) as needed. You may also deploy multiple cluster at the same time, using different names.

### Access to the cluster

Minikube provide a UI, should you want to keep track of deployment and/or inspect the cluster. To start it, use the following command:

1. `minikube dashboard -p dss-local-cluster`

You can also use any other tool as needed. You can switch to the cluster's context by using the following command:

1. `kubectl config use-context dss-local-cluster`

### Upload or update local image

Should you want to run the local docker image that you [built](../../../../build/README.md), run the following commands to upload / update your image

1. `minikube image -p dss-local-cluster push interuss-local/dss`

In the helm charts, use `docker.io/interuss-local/dss:latest` as image and be sure to set the `imagePullPolicy` to `Never`.

## Deployment of the DSS services

You can now deploy the DSS services using [helm charts](../../../services/helm-charts/dss/README.md).

Follow the instructions in the [README](../../../services/helm-charts/dss/README.md), especially the ones related to certificate generation and publication to the cluster. However, there are some minikube specific things to do:

* Use the `global.cloudProvider` setting with the value `minikube` and deploy the charts on the `dss-local-cluster` kubernetes context.
* To access the service, find the external IP using the `kubectl get services dss-dss-gateway` command. The port 80, without HTTPs is used.

You may also use the tanka files to deploy the service. An example configuration is provided [there](../../../services/tanka/examples/minikube/).

## Clean up

To delete all resources, run `minikube delete -p dss-local-cluster`.  Note that this operation can't be reverted and all data will be lost.
