# Modifying a deployment

This page describes how to apply a configuration change to a deployed DSS instance, for instance to tune the
[regular cleanup](cleanup.md#regular-cleanup) or the [CPU and memory allocation](performance.md#cpu-and-memory-allocation).

Some changes require a specific procedure:

- Flags listed in [flags compatibility](upgrades.md#flags-compatibility) must be changed using the [Zero-Traffic upgrade procedure](upgrades.md#zero-traffic-upgrade-procedure).
- Kubernetes version changes are described in [Kubernetes upgrades](kubernetes-upgrades.md).
- CockroachDB version changes are described in [CockroachDB version upgrades](crdb-upgrades.md).
- The [regular cleanup](cleanup.md#regular-cleanup) applies to the whole pool and should be coordinated with the other members of the pool.
- The nodes of a pool running the [Raft store](../deployment/pooling/raft.md) read their certificates and CA bundle only at startup: updating the `dss.raft.certs` secret requires [restarting the nodes](#restarting-nodes).

## Applying a change

=== "Terraform"
    1. Update the variables in your `terraform.tfvars`. See TFVARS.gen.md ([AWS](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-aws-dss/TFVARS.gen.md), [Google](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-google-dss/TFVARS.gen.md)) for variables descriptions.
    1. From your infrastructure folder, run `terraform plan` and review the changes.
    1. Run `terraform apply` to apply the changes to the infrastructure and regenerate the Tanka and Helm configuration.
    1. Go to `/build/workspace/${cluster_context}` and run `./get-credentials.sh` to login to kubernetes.
    1. Apply the regenerated configuration to the cluster:
        * Tanka: go to `/deploy/services/tanka/workspace/${cluster_context}` and run `tk apply .`.
        * Helm: go to `/deploy/services/helm-charts/dss` and run `helm upgrade --kube-context=$KUBE_CONTEXT -f /build/workspace/${cluster_context}/helm_values.yml $RELEASE_NAME .`.

=== "Tanka"
    1. Update the configuration in your `main.jsonnet`.
    1. Run `tk diff .` and review the changes.
    1. Run `tk apply .`.

=== "Helm"
    1. Update your values file.
    1. Go to `/deploy/services/helm-charts/dss` and run `helm upgrade --kube-context=$KUBE_CONTEXT -f [values file] $RELEASE_NAME .`.

    Set values in the values file rather than with `helm upgrade --set`, to keep the values file aligned with the state of the cluster.

## Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the procedures in this section may change.

### Restarting nodes

The data of a node (its identity, its Raft log and its snapshots) is stored on
the persistent volume `raft-data` of its pod, mounted at `/raft_data`.  A
restarted node reloads it and rejoins the pool: it replays its log, or
receives a snapshot from a peer if it has been down too long.

The nodes of an instance are restarted with a single command; the stateful
set rolls them back one at a time, in pod order, waiting for a node to be
ready — back to `member` — before restarting the next one:

```shell
kubectl --context $CLUSTER_CONTEXT rollout restart statefulset/core-service --namespace $NAMESPACE
```

(With the helm chart, the statefulset is named `<release>-core-service`, e.g.
`dss-core-service`.)

Use this to restart the nodes whenever their `dss.raft.certs` secret has been
updated (see [certificates](../deployment/pooling/raft.md#certificates)).

If the data directory of a node is lost or corrupted, the node must not start
with an empty one: its identity would be regenerated while the pool still
remembers its previous one.  Remove the old identity from the pool, delete the
data directory, restart the node and add it back to the pool: see
[replacing a node](pooling.md#replacing-a-node).
