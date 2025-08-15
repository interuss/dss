# DSS Deployment

## Introduction

An operational DSS deployment requires a specific architecture to be compliant with [standards requirements](https://github.com/interuss/dss?tab=readme-ov-file#standards-and-regulations) and meet performance expectations as described in [architecture](./architecture.md).  This page describes the deployment procedures recommended by InterUSS to achieve this compliance and meet these expectations.

## Deployment layers

This repository provides three layers of abstraction to deploy and operate a DSS instance via Kubernetes.

![Deployment layers](assets/deployment_layers.png)

As described below, InterUSS provides tooling for Kubernetes deployments on Amazon Web Services (EKS) and Google Cloud (GKE).
However, you can do this on any supported [cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/) or even on your own infrastructure.
Review [InterUSS pooling requirements](./architecture.md#objective) and consult the Kubernetes documentation for your chosen provider.

The three layers are the following:

1. [Infrastructure](./infrastructure/index.md) provides instructions and tooling to easily provision a Kubernetes cluster and cloud resources (load balancers, storage...) to a cloud provider. The resulting infrastructure meets the [Pooling requirements](./architecture.md#objective).
   Terraform modules are provided for:
    - [Amazon Web Services (EKS)](infrastructure/terraform-aws-dss/index.md)
    - [Google (GKE)](infrastructure/terraform-google-dss/index.md)

1. Services provides the tooling to deploy a DSS instance to a Kubernetes cluster.
    - [Helm Charts](services/helm-charts.md)
    - [Tanka](services/tanka.md)

1. [Operations](./operations/index.md) provides instructions to operate a deployed DSS instance.
    - [Pooling procedure](./operations/index.md#pooling-procedure)
    - [Troubleshooting](./operations/troubleshooting.md)

Depending on your level of expertise and your internal organizational practices, you should be able to use each layer independently or complementary.

For local deployment approaches, see the documentation located in the [build folder](./build.md#deployment-options)

## Getting started

You can find below two guides to deploy a DSS instance from scratch:
- [Amazon Web Services (EKS)](infrastructure/terraform-aws-dss/index.md#getting-started)
- [Google (GKE)](infrastructure/terraform-google-dss/index.md#getting-started)

For a complete use case, you can look into the configurations of the [CI job](https://github.com/interuss/dss/blob/master/.github/workflows/dss-deploy.yml) in operations: [ci](operations/ci/index.md)

## Migrations and upgrades

Information related to migrations and upgrades can be found in [MIGRATION.md](migration.md).

## Development

The following diagram represents the resources in this repository per layer.
![Deploy Overview](assets/generated/deploy_overview.png)

### Formatting

Terraform files must be formatted using `terraform fmt -recursive` command to pass the CI linter check.
