# Infrastructure

As a phase in [DSS deployment](..), this folder contains the terraform modules required to prepare the infrastructure to host a DSS deployment.  To deploy infrastructure manually (rather than terraform, as described here), see ["Deploying a DSS instance via Kubernetes"](../../build/README.md#deploying-a-dss-instance-via-kubernetes).

See [Services](../README.md#services) to deploy the DSS once the infrastructure is ready.

## Modules
The [modules](modules) directory contains the terraform public modules required to prepare the infrastructure on various cloud providers.

- [terraform-aws-dss](./modules/terraform-aws-dss/README.md): Amazon Web Services deployment
- [terraform-google-dss](./modules/terraform-google-dss/README.md): Google Cloud Engine deployment

This terraform module creates a Kubernetes cluster in Amazon Web Services using the Elastic Kubernetes Service (EKS)
and generates the tanka files to deploy a DSS instance.
This terraform module creates a Kubernetes cluster in Google Cloud Engine and generates
the tanka files to deploy a DSS instance.


## Dependencies
The [dependencies](dependencies) directory contains submodules used by the public modules described above. They are not expected to be
used directly by users. Those submodules are the combination of the cloud specific dependencies `terraform-*-kubernetes`
and `terraform-common-dss`. `terraform-common-dss` module aggregates and outputs the infrastructure configuration
which can be used as input to the `Services` deployment as shown in the diagram below.

![Infrastructure Modules](../../assets/generated/deploy_infrastructure_modules.png)

## Local

The [local](local) directory contains various documentation that can be used to spawn a cluster locally.

- [minikuke](./local/minikube/README.md): Minikube local deployment

## Utils
This [utils folder](utils) contains scripts to help manage the terraform modules and dependencies. See the README in that folder for details.
