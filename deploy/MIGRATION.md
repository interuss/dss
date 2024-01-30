# Kubernetes version migration

This page provides information on how to upgrade your Kubernetes cluster deployed using the
tools from this repository.

**Important notes:**

- The migration plan below has been tested with the deployment of services using [Helm](services/helm-charts) and [Tanka](../build/deploy) without Istio enabled. Note that this configuration flag has been decommissioned since [#995](https://github.com/interuss/dss/pull/995).
- Further work is required to test and evaluate the availability of the DSS during migrations.
- It is highly recommended to rehearse such operation on a test cluster before applying them to a production environment.

## Google - Google Kubernetes Engine

Migrations of GKE clusters are managed using terraform.

### 1.27 to 1.28

1. Change your `terraform.tfvars` to use `1.28` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.28
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

### 1.26 to 1.27

1. Change your `terraform.tfvars` to use `1.27` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.27
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

### 1.25 to 1.26

1. Change your `terraform.tfvars` to use `1.26` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.26
   ```
2. Run `terraform apply`
3. Monitor the upgrade of the nodes in the Google Cloud console.

### 1.24 to 1.25

1. Change your `terraform.tfvars` to use `1.25` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.25
   ```
2. Run `terraform apply`. This operation may take more than 30min.
3. Monitor the upgrade of the nodes in the Google Cloud console.

## AWS - Elastic Kubernetes Service

Currently, upgrades of EKS can't be achieved reliably with terraform directly. The recommended workaround is to
use the web console of AWS Elastic Kubernetes Service (EKS) to upgrade the cluster.
Before proceeding, always check on the cluster page the *Upgrade Insights* tab which provides a report of the
availability of Kubernetes resources in each version. The following sections omit this check if no resource is
expected to be reported in the context of a standard deployment performed with the tools in this repository.

### 1.27 to 1.28

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.28` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.28
   ```
   
### 1.26 to 1.27

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.27` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.27
   ```
   
### 1.25 to 1.26

1. Upgrade the cluster (control plane) using the AWS console. It should take ~15 minutes.
2. Update the *Node Group* in the *Compute* tab with *Rolling Update* strategy to upgrade the nodes using the AWS console.
3. Change your `terraform.tfvars` to use `1.26` by adding or updating the `kubernetes_version` variable:
   ```terraform
   kubernetes_version = 1.26
   ```
   
### 1.24 to 1.25

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
