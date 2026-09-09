# Deploying DSS infrastructure

This section describes how to deploy the infrastructure for a DSS instance.

## Prerequisites

Before beginning infrastructure deployment, download & install the following tools to your workstation:

- [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) to
  interact with kubernetes
    - Confirm successful installation with `kubectl version --client` (should
      succeed from any working directory).
    - Note that kubectl can alternatively be installed via the Google Cloud SDK
     `gcloud` shell if using Google Cloud.

## Deployment Options

The DSS can be deployed on various platforms. Choose the method that best suits your needs:

| Platform | Tools | Description |
| :--- | :--- | :--- |
| **Amazon Web Services** | Terraform | [Deploy on AWS using Terraform](aws.md) to provision EKS and required resources. |
| **Google Cloud Platform** | Terraform | [Deploy on GCP using Terraform](google.md) to provision GKE and required resources. |
| **Locally** | Minikube | [Deploy locally using Minikube](minikube.md) for development and testing. |
