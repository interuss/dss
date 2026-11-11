# Tanka library

This folder contains a set of configuration to be used with [tanka](https://tanka.dev/) to deploy a single DSS instance in a Kubernetes cluster.

## Requirements

1. A Kubernetes cluster should be running and you should be properly authenticated. Requirements and instructions to create a new Kubernetes cluster can be found here:
    * [AWS](../../../docs/deployment/infrastructure/aws.md)
    * [Google](../../../docs/deployment/infrastructure/google.md)
    * [Minikube](../../../docs/deployment/infrastructure/minikube.md)

2. Create the certificates and apply them to the cluster using the instructions [here](../../../docs/deployment/pooling/index.md)
3. Install [Tanka](https://tanka.dev/install)

## Usage

There is a base file in `examples/minikube`. You can directly use it to deploy the service.

Check first that you cluster is using `192.168.49.2` as IP with `minikube ip -p dss-local-cluster`. If that not the case, edit `examples/minikube/spec.json` accordingly.

To apply changes, use the following command:

* `cd examples/minikube`
* `tk apply .`

You may edit `main.jsonnet` as needed should you want to change configuration, or docker images used.

## Form the Raft groups

With the Raft datastore, each node of `raft.nodes` starts waiting and is not ready until it is a member of the Raft groups (aux, rid, scd) of the pool. The groups are formed once for the whole pool with `dssctl`, which is available in the core-service image and talks to the node on its admin socket. The pods are named `core-service-<i>`, in the namespace of the environment's `spec.json`:

1. On the very first deployment of the pool only, bootstrap the groups on one node:
   `kubectl --context=$KUBE_CONTEXT --namespace=$NAMESPACE exec core-service-0 -- dssctl raft bootstrap`
2. Add every other node of the pool, from this instance or from other participants, one at a time, from any member, with the URL of its endpoint:
   `kubectl --context=$KUBE_CONTEXT --namespace=$NAMESPACE exec core-service-0 -- dssctl raft add --url https://<endpoint>`
3. Check the members: `kubectl --context=$KUBE_CONTEXT --namespace=$NAMESPACE exec core-service-0 -- dssctl raft list`

A node restarted with its persistent volume rejoins its groups on its own. A node whose volume was lost starts waiting again: remove its old ID with `dssctl raft remove` and add it again.

## Job cleanup

Jobs are not automatically removed. It is possible to use tanka to delete unmanaged resources (eg previous jobs) by enabling the [garbage collection](https://tanka.dev/garbage-collection/) feature in your `spec.json` file. Use `tk prune` to cleanup resources not present anymore in the jsonnet.
