# Deployment of DSS services to Minikube

## Upload or update local image

Should you want to run the local docker image that you [built](./index.md#docker-images), run the following commands to upload / update your image

1. `minikube image -p dss-local-cluster load interuss-local/dss`

In the helm charts, use `docker.io/interuss-local/dss:latest` as image and be sure to set the `imagePullPolicy` to `Never`.

## Deployment

You can now deploy the DSS services using Helm or Tanka. See the repository `/deploy/services` for more information.

=== "Helm"
    Minikube specific settings:

    * Use the `global.cloudProvider` setting with the value `minikube` and deploy the charts on the `dss-local-cluster` kubernetes context.

=== "Tanka"
    An example configuration is provided in the repository: `/deploy/services/tanka/examples/minikube`

---

To access the service, find the external IP using the `kubectl get services dss-dss-gateway` command. The port 80, without HTTPs is used.

## Next steps

[Decommission the Minikube instance](../../decommissioning/minikube.md) when it is no longer needed.
