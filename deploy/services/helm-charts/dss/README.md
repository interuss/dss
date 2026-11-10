# Helm Chart
This directory provides an [Helm Chart](https://helm.sh/) to deploy the DSS and datastore (CockroachDB, Yugabyte, or the Raft-based store) kubernetes resources.

## Requirements
1. A Kubernetes cluster should be running and you should be properly authenticated. Requirements and instructions to create a new Kubernetes cluster can be found here:
    * [AWS](../../../../docs/deployment/infrastructure/aws.md)
    * [Google](../../../../docs/deployment/infrastructure/google.md)
    * [Minikube](../../../../docs/deployment/infrastructure/minikube.md)

2. Create the certificates and apply them to the cluster using the instructions [here](../../../../docs/deployment/pooling/index.md)
3. Install [Helm](https://helm.sh/) version 3.11.3 or higher

## Usage
1. Copy `values.example.yaml` to `values.dev.yaml` and edit it. In particular, the key `dss.image` must be set manually. See `values.schema.json` for schema definition. The root key `cockroachdb` supports all values supported by the [`cockroachdb` Chart](https://github.com/cockroachdb/helm-charts/tree/master/cockroachdb#configuration)). Note that values.yaml contains the default values and are always passed to helm.
2. Validate the configuration: `helm lint -f values.dev.yaml .`
3. Set a RELEASE_NAME to `dss`: `export RELEASE_NAME=dss`
It is temporarily the only release name possible.
4. Set the kube client context of your system, example: `export KUBE_CONTEXT=gke_interuss-deploy-example_europe-west6-a_dss-dev-w6`
5. Run `helm dep update --kube-context=$KUBE_CONTEXT`
6. Install the chart: `helm install --kube-context=$KUBE_CONTEXT -f values.dev.yaml $RELEASE_NAME .`

### Update the chart
When changing the values in values.dev.yaml, values.yaml, the templates or upgrading the helm chart dependencies, changes can be applied to the cluster using the following command:

1. Run `helm upgrade --kube-context=$KUBE_CONTEXT -f values.dev.yaml $RELEASE_NAME .`

### Form the Raft groups

With `raft.enabled`, each node of `raft.nodes` starts waiting and is not ready until it is a member of the Raft groups (aux, rid, scd) of the pool. The groups are formed once for the whole pool with `dssctl`, which is available in the core-service image and talks to the node on its admin socket:

1. On the very first deployment of the pool only, bootstrap the groups on one node:
   `kubectl --context=$KUBE_CONTEXT exec $RELEASE_NAME-core-service-0 -- dssctl raft bootstrap`
2. Add every other node of the pool, from this release or from other participants, one at a time, from any member, with the URL of its endpoint:
   `kubectl --context=$KUBE_CONTEXT exec $RELEASE_NAME-core-service-0 -- dssctl raft add --url https://<endpoint>`
3. Check the members: `kubectl --context=$KUBE_CONTEXT exec $RELEASE_NAME-core-service-0 -- dssctl raft list`

A node restarted with its persistent volume rejoins its groups on its own. A node whose volume was lost starts waiting again: remove its old ID with `dssctl raft remove` and add it again.

### Disable CockroachDB probes

The `cockroachdb` Chart always renders a liveness and a readiness probe. To disable one of them (see https://github.com/interuss/dss/issues/1732 for potential reasons to do so), replace it by a probe that always succeeds in `values.dev.yaml`:

```yaml
cockroachdb:
  statefulset:
    customLivenessProbe:
      exec:
        command: ["/bin/bash", "-c", "true"]
    customReadinessProbe:
      exec:
        command: ["/bin/bash", "-c", "true"]
```
