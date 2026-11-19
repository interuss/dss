# Modifying a deployment

This page describes how to apply a configuration change to a deployed DSS instance, for instance to tune the
[regular cleanup](cleanup.md#regular-cleanup) or the [CPU and memory allocation](performance.md#cpu-and-memory-allocation).

Some changes require a specific procedure:

- Flags listed in [flags compatibility](upgrades.md#flags-compatibility) must be changed using the [Zero-Traffic upgrade procedure](upgrades.md#zero-traffic-upgrade-procedure).
- Kubernetes version changes are described in [Kubernetes upgrades](kubernetes-upgrades.md).
- CockroachDB version changes are described in [CockroachDB version upgrades](crdb-upgrades.md).
- The [regular cleanup](cleanup.md#regular-cleanup) applies to the whole pool and should be coordinated with the other members of the pool.

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
