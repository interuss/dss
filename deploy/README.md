# DSS Deployment

> This folder contains the increments toward the new deployment approach as described in #874.

An operational DSS requires two different services: the DSS core-service API and the Cockroach database. 
This folder contains the tools to prepare the infrastructure in multiple cloud providers, deploy the services and operate it.

The deployment tools are conceptually broken down in three phases:

- [Infrastructure](#infrastructure)
- [Services](#services)
- [Operations](#operations)

## [Infrastructure](./infrastructure)
It is responsible to prepare infrastructure on various cloud providers to accept deployment of Services below. It includes the Kubernetes cluster creation, cluster nodes, load balancer and associated fixed IPs, etc. This stage is cloud provider specific.
  
Terraform modules are provided for:
- [Amazon Web Services (EKS)](infrastructure/modules/terraform-aws-dss)
- [Google (GKE)](infrastructure/modules/terraform-google-dss)

## [Services](./services)
It is responsible for managing Kubernetes resources and **deploying** the Services required by an operational DSS. The ambition is to be cloud provider agnostic for the Services part. 

Services can be deployed using those approaches:
  - [Helm Charts](services/helm-charts/dss)
  - [Tanka](../build/deploy)

## [Operations](./operations)
It is responsible to provide diagnostic capabilities and utilities to **operate** the Services, such as certificates management may be simplified using the deployment manager CLI tools. It also contains the Infrastructure and Services configurations [used by the CI](../.github/workflows/dss-deploy.yml).

The following diagram represents the modules provided in this repository per phase and their impact on the various resources.
![Deploy Overview](../assets/generated/deploy_overview.png)

## Getting started

If you wish to deploy a DSS from scratch, "Getting Started" instructions can be found in the terraform modules and covers all steps to get a running DSS:
- [Amazon Web Services (EKS)](infrastructure/modules/terraform-aws-dss/README.md#Getting-started)
- [Google (GKE)](infrastructure/modules/terraform-google-dss/README.md#Getting-started)

For a real use case, you can look into the configurations of the [CI job](../.github/workflows/dss-deploy.yml) in operations: [ci](operations/ci)

## Migrations and upgrades

Information related to migrations and upgrades can be found in [MIGRATION.md](MIGRATION.md).

## Development

### Formatting

Terraform files must be formatted using `terraform fmt -recursive` command to pass the CI linter check.
