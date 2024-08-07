# DSS Helm Chart
This [Helm Chart](https://helm.sh/) deploys the DSS and cockroachdb kubernetes resources.

## Requirements
1. A Kubernetes cluster should be running and you should be properly authenticated.
Requirements and instructions to create a new Kubernetes cluster can be found [here](../../../infrastructure/README.md).
2. Create the certificates and apply them to the cluster using the instructions of [section 6 and 7](../../../../build/README.md)
3. Install [Helm](https://helm.sh/) version 3.11.3 or higher

## Usage
1. Copy `values.example.yaml` to `values.dev.yaml` and edit it. In particular, the key `dss.image` must be set manually. See `values.schema.json` for schema definition. The root key `cockroachdb` supports all values supported by the [`cockroachdb` Chart](https://github.com/cockroachdb/helm-charts/tree/master/cockroachdb#configuration)). Note that values.yaml contains the default values and are always passed to helm.
2. Validate the configuration: `helm lint -f values.dev.yaml .`
3. Set a RELEASE_NAME to `dss`: `export RELEASE_NAME=dss`
It is temporarily the only release name possible.
4. Set the kube client context of your , example: `export KUBE_CONTEXT=gke_interuss-deploy-example_europe-west6-a_dss-dev-w6`
5. Run `helm dep update --kube-context=$KUBE_CONTEXT`
6. Install the chart: `helm install --kube-context=$KUBE_CONTEXT -f values.dev.yaml $RELEASE_NAME .`

### Update the chart
When changing the values in values.dev.yaml, values.yaml, the templates or upgrading the helm chart dependencies, changes can be applied to the cluster using the following command:

1. Run `helm upgrade --kube-context=$KUBE_CONTEXT -f values.dev.yaml $RELEASE_NAME .`
