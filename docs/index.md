# DSS Deployment User Documentation

## Introduction

This website provides instructions to deploy the InterUSS USS to USS Discovery and Synchronization service.

An operational DSS deployment requires a specific architecture to be compliant with [standards requirements](https://github.com/interuss/dss?tab=readme-ov-file#standards-and-regulations) and meet performance expectations as described in [architecture](./architecture.md).  This page describes the deployment procedures recommended by InterUSS to achieve this compliance and meet these expectations.

## Getting started

- Deploy a DSS instance on [Amazon Web Services (EKS)](infrastructure/aws.md) using terraform
- Deploy a DSS instance on [Google (GKE)](infrastructure/google.md) using terraform
- Deploy a DSS instance on [Google (GKE)](infrastructure/google.md) step by step
- Deploy a DSS instance on [Minikube](infrastructure/minikube.md)

## Deployment layers

The deployment of a DSS instance involves 3 stages:

1. Provisioning the required cloud resources, in particular a Kubernetes cluster: **The Infrastructure**.

1. Deploying the DSS applications under the form of kubernetes resources: **The Services**.

1. Recommending procedures and guidelines on how to operate the DSS: **The Operations**.

![Deployment layers](assets/deployment_layers.png)

As described below, InterUSS provides tooling for Kubernetes deployments on Amazon Web Services (EKS) and Google Cloud (GKE).
However, you can do this on any supported [cloud provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/) or even on your own infrastructure.
Review [InterUSS pooling requirements](./architecture.md#objective) and consult the Kubernetes documentation for your chosen provider.

Depending on your level of expertise and your internal organizational practices, you should be able to use each layer independently or complementary.

## Migrations and upgrades

Information related to migrations and upgrades can be found in [the migration section](operations/migrations.md).

## Development

The following diagram represents the resources in this repository per layer.
![Deploy Overview](assets/generated/deploy_overview.png)

### Formatting

Terraform files must be formatted using `terraform fmt -recursive` command to pass the CI linter check.
