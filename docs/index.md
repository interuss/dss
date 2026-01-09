# DSS Deployment User Documentation

## Introduction

This website provides instructions to deploy the InterUSS USS to USS Discovery and Synchronization service.

An operational DSS deployment requires a specific architecture to be compliant with [standards requirements](https://github.com/interuss/dss?tab=readme-ov-file#standards-and-regulations) and meet performance expectations as described in [architecture](./architecture.md).  
This page describes the deployment procedures recommended by InterUSS to achieve this compliance and meet these expectations.


## Getting started

- Review [architecture requirements](architecture.md)
- Deploy a DSS instance to [Amazon Web Services (EKS)](infrastructure/aws.md) using terraform
- Deploy a DSS instance to [Google (GKE)](infrastructure/google.md) using terraform
- Deploy a DSS instance to [Google (GKE)](infrastructure/google-manual.md) manually step by step
- Deploy a DSS instance to [Minikube](infrastructure/minikube.md)

## Tooling

The deployment of a DSS instance involves 3 stages:

1. Provisioning the required cloud resources, in particular a Kubernetes cluster: **The Infrastructure**.

1. Deploying the DSS applications under the form of kubernetes resources: **The Services**.

1. Recommending procedures and guidelines on how to operate the DSS: **The Operations**.

![Deployment layers](assets/generated/deployment_layers.png)

Depending on your level of expertise and your internal organizational practices, you should be able to use each layer independently or complementary.

InterUSS offers two terraform modules to deploy the **Infrastructure**:

- [Amazon Web Services](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-aws-dss/)
- [Google Cloud Platform](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-google-dss/)

The **Services** are deployed using the following tools:

- [Tanka](https://github.com/interuss/dss/blob/master/deploy/services/tanka/)
- [Helm Chart](https://github.com/interuss/dss/blob/master/deploy/services/helm-charts/dss)

See [Operate a DSS instance](operations/index.md) for more information on tools to perform the **Operations**.
