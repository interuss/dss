# CockroachDB and Kubernetes version migration

This page provides information on how to upgrade your CockroachDB and Kubernetes cluster deployed using the
tools from this repository.

## CockroachDB upgrades

Cockroach DB must be upgraded on all DSS instances of the pool at the same time. Therefore, all DSS instances
connected to a pool must coordinate the upgrade. The rollout of the upgrades on the whole CRDB cluster
must be carefully performed in a sequence in order to keep the majority of nodes healthy during that period
and prevent downtime.
For a Pooled deployment, one of the DSS Instance must take the role of the upgrade "Leader" and coordinate the
upgrade with other "Followers" DSS instances.
In general a Cockroach DB upgrade consists of:
1. Upgrade preparation: Verify that the cluster is in a nominal state ready for upgrade.
1. Decide how the upgrade will be finalized (for major upgrades only): Like CockroachDB, we recommend disabling auto-finalization.
1. Perform the rolling upgrade: This step should be performed first by the Leader and as quickly as possible by the Followers **one after the other**. Note that during this period, the performance of the cluster may be impacted since, as documented by CockroachDB, "a query that is sent to an upgraded node can be distributed only among other upgraded nodes. Data accesses that would otherwise be local may become remote, and the performance of these queries can suffer."
1. Roll back the upgrade (optional): Like the rolling upgrade, this step should be carefully coordinated with all DSS instances to guarantee the minimum number of healthy nodes to keep the cluster available.
1. Finish the upgrade: This step should be accomplished by the Leader.

The following sections provide links to the CockroachDB migration documentation.

**Important notes:**

- Further work is required to test and evaluate the availability of the DSS during migrations.
- We recommend to review carefully the instructions provided by CockroachDB and to rehearse all migrations on a test
  environment before applying them to production.

### Terraform deployment notes

If a DSS instance has been deployed with terraform, first upgrade the cluster using [Helm](MIGRATION.md#helm-deployment-notes) or [Tanka](MIGRATION.md#tanka-deployment-notes).
Then, update the variable `crdb_image_tag` in your `terraform.tfvars` to align your configuration with the new state of
the cluster and avoid a conflict on the next terraform update.

### Helm deployment notes

If you deployed the DSS using the Helm chart and the instructions provided in this repository, follow the instructions
provided by CockroachDB `Cluster Upgrade with Helm` (See specific links below). Note that the CockroachDB documentation
suggest to edit the values using `helm upgrade ... --set` commands. However, you can alternatively update `helm_values.yml`
in your deployment. With both approaches, you will need to use the root key `cockroachdb` since the cockroachdb helm chart is
a dependency of the dss chart.
For instance, setting the image tag and partition using the command line would look like this:
```
helm upgrade [RELEASE_NAME] [PATH_TO_DSS_HELM] --set cockroachdb.image.tag=v24.1.3 --reuse-values
```
```
helm upgrade [RELEASE_NAME] [PATH_TO_DSS_HELM] --set cockroachdb.statefulset.updateStrategy.rollingUpdate.partition=0 --reuse-values
```
and if editing the image tag and rollout partition in your `values.yaml`, it would look like this:
```yaml
cockroachdb:
  image:
    # ...
    tag: # version
  statefulset:
    updateStrategy:
      rollingUpdate:
        partition: 0
```
New values can then be applied using `helm upgrade [RELEASE_NAME] [PATH_TO_DSS_HELM] -f [values.yaml]`

#### 21.2.7 to 24.1.3

CockroachDB requires to upgrade one minor version at a time:

1. 21.2.7 to 22.1: see [CockroachDB Cluster upgrade for Helm](https://www.cockroachlabs.com/docs/v22.1/upgrade-cockroachdb-kubernetes?filters=helm).
2. 22.1 to 22.2: see [CockroachDB Cluster upgrade for Helm](https://www.cockroachlabs.com/docs/v22.2/upgrade-cockroachdb-kubernetes?filters=helm).
3. 22.2 to 23.1: see [CockroachDB Cluster upgrade for Helm](https://www.cockroachlabs.com/docs/v23.1/upgrade-cockroachdb-kubernetes?filters=helm).
4. 23.1 to 23.2: see [CockroachDB Cluster upgrade for Helm](https://www.cockroachlabs.com/docs/v23.2/upgrade-cockroachdb-kubernetes?filters=helm).
5. 23.2 to 24.1.3: see [CockroachDB Cluster upgrade for Helm](https://www.cockroachlabs.com/docs/v24.1/upgrade-cockroachdb-kubernetes?filters=helm).

### Tanka deployment notes

For deployments using Tanka configuration, following the `Cluster Upgrade with Manual configs` approach is recommended. (See specific links below)
To apply the changes to your cluster, you have two options:
1. Follow the manual steps and reflect the new values in the *Leader* and *Followers* Tanka configuration, especially the new image version 
(see [`VAR_CRDB_DOCKER_IMAGE_NAME`](../build/README.md)) to ensure the new configuration is aligned with the cluster state.

#### 21.2.7 to 24.1.3

CockroachDB requires to upgrade one minor version at a time. As mentioned above, no instruction is provided for tanka specifically,
therefore we recommend to follow the manual steps:

1. 21.2.7 to 22.1: see [CockroachDB Cluster upgrade with Manual configs](https://www.cockroachlabs.com/docs/v22.1/upgrade-cockroachdb-kubernetes?filters=manual).
2. 22.1 to 22.2: see [CockroachDB Cluster upgrade with Manual configs](https://www.cockroachlabs.com/docs/v22.2/upgrade-cockroachdb-kubernetes?filters=manual).
3. 22.2 to 23.1: see [CockroachDB Cluster upgrade with Manual configs](https://www.cockroachlabs.com/docs/v23.1/upgrade-cockroachdb-kubernetes?filters=manual).
4. 23.1 to 23.2: see [CockroachDB Cluster upgrade with Manual configs](https://www.cockroachlabs.com/docs/v23.2/upgrade-cockroachdb-kubernetes?filters=manual).
5. 23.2 to 24.1.3: see [CockroachDB Cluster upgrade with Manual configs](https://www.cockroachlabs.com/docs/v24.1/upgrade-cockroachdb-kubernetes?filters=manual).

## Kubernetes upgrades

**Important notes:**

- The migration plan below has been tested with the deployment of services using [Helm](services/helm-charts) and [Tanka](../build/deploy) without Istio enabled. Note that this configuration flag has been decommissioned since [#995](https://github.com/interuss/dss/pull/995).
- Further work is required to test and evaluate the availability of the DSS during migrations.
- It is highly recommended to rehearse such operation on a test cluster before applying them to a production environment.

### Google - Google Kubernetes Engine

Migrations of GKE clusters are managed using terraform.

#### 1.27 to 1.28

1. Change your `terraform.tfvars` to use `1.28` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.28
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

#### 1.26 to 1.27

1. Change your `terraform.tfvars` to use `1.27` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.27
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

#### 1.25 to 1.26

1. Change your `terraform.tfvars` to use `1.26` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.26
   ```
2. Run `terraform apply`
3. Monitor the upgrade of the nodes in the Google Cloud console.

#### 1.24 to 1.25

1. Change your `terraform.tfvars` to use `1.25` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.25
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

### AWS - Elastic Kubernetes Service

Currently, upgrades of EKS can't be achieved reliably with terraform directly. The recommended workaround is to
use the web console of AWS Elastic Kubernetes Service (EKS) to upgrade the cluster.
Before proceeding, always check on the cluster page the *Upgrade Insights* tab which provides a report of the
availability of Kubernetes resources in each version. The following sections omit this check if no resource is
expected to be reported in the context of a standard deployment performed with the tools in this repository.

#### 1.27 to 1.28

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.28` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.28
   ```
   
#### 1.26 to 1.27

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.27` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.27
   ```
   
#### 1.25 to 1.26

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.26` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.26
   ```
   
#### 1.24 to 1.25

1. Check for deprecated resources:
    - Click on the Upgrade Insights tab to see deprecation warnings on the cluster page.
    - Evaluate errors in Deprecated APIs removed in Kubernetes v1.25. Using `kubectl get podsecuritypolicies`,
      check if there is only one *Pod Security Policy* named `eks.privileged`. If it is the case,
      according to the [AWS documentation](https://docs.aws.amazon.com/eks/latest/userguide/pod-security-policy-removal-faq.html), you can proceed.
2. Upgrade the cluster using the AWS console. It should take ~15 minutes.
3. Change your `terraform.tfvars` to use `1.25` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.25
   ```
