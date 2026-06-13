# Deploy a DSS instance to Google Cloud Platform with terraform

This guide will help you deploy a DSS instance to Google Cloud Platform with terraform and tanka.

## Getting started

### Prerequisites

#### Terraform

1. Install [terraform](https://developer.hashicorp.com/terraform/downloads) to your workstation.
    1. Verify installation with `terraform version`

#### Kubernetes tools

1. Install kubectl from [Prerequisites](index.md#prerequisites)
    1. Verify kubectl installation with `kubectl version`

#### Google Cloud Platform

1. Install and initialize [Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk).
    1. Confirm successful initialization with `gcloud info`; check "Account".
2. Ensure a GCP project is available (create one in web UI if needed)
    1. Consider `$GOOGLE_PROJECT_NAME` to refer to this project
3. Check that the GCP DSS project is correctly selected: `gcloud config list project`
    1. Set another one if needed using: `gcloud config set project $GOOGLE_PROJECT_NAME`
4. Enable the following API using [Google Cloud CLI](https://cloud.google.com/endpoints/docs/openapi/enable-api#gcloud):
    1. `gcloud services enable compute.googleapis.com`
    2. `gcloud services enable container.googleapis.com`
    3. If you want to manage DNS entries with terraform: `gcloud services enable dns.googleapis.com`
5. Install the auth plugin to connect to kubernetes: `gcloud components install gke-gcloud-auth-plugin`
6. Run `gcloud auth application-default login` to generate credentials to call Google Cloud Platform APIs.
    1. If the result of performing the authorization indicates 404 in the browser, check whether a local dummy-oauth instance is running (using port 8085).  Stop the dummy-oauth instance if it is running.

### Deployment of the Kubernetes cluster

!!! tip "Paths in the documentation"
    In the documentation, we often refer to path starting from the root (prefixed with `/`). This is to indicate that the path is relative to the root of the repository.

1. Create a new folder in `/deploy/infrastructure/personal/` for the deployment named for example: `terraform-google-dss-dev`
2. From `/deploy/infrastructure/modules/terraform-google-dss`, copy `main.tf`, `output.tf`, `variables.gen.tf` and `terraform.dev.example.tfvars` to the infrastructure personal folder.
3. In the infrastructure personal folder:
    1. Rename `terraform.dev.example.tfvars` to `terraform.tfvars`.
    2. Check that the directory contains the following files:
        1. main.tf
        2. output.tf
        3. terraform.tfvars
        4. variables.gen.tf
    3. Set the variables in `terraform.tfvars` according to your environment. See [TFVARS.gen.md](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-google-dss/TFVARS.gen.md) for variables descriptions.
    4. Initialize terraform: `terraform init`.
    5. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
    6. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)
4. Configure the DNS resolution to the public ip addresses. DNS entries can be either managed with terraform or managed manually:

=== "Terraform managed DNS entries"
    If your DNS zone is managed on the same account, it is possible to instruct terraform to create and manage
    it with the rest of the infrastructure.

    - **For Google Cloud Platform**, configure the zone in your google account and set the `google_dns_managed_zone_name` variable the zone name. Zones can be listed by running `gcloud dns managed-zones list`. Entries will be automatically created by terraform.

=== "Manual DNS entries setup"
    If DNS entries are managed manually, set them up manually using the following steps:

    1. Retrieve IP addresses and expected hostnames: `terraform output`
    Example of expected output:
    ```
        crdb_addresses = [
            {
                "address" = "34.65.15.23"
                "expected_dns" = "0.interuss.example.com"
            },
            {
                "address" = "34.65.146.56"
                "expected_dns" = "1.interuss.example.com"
            },
            {
                "address" = "34.65.191.145"
                "expected_dns" = "2.interuss.example.com"
            },
        ]
        gateway_address = {
            "address" = "35.186.236.146"
            "expected_dns" = "dss.interuss.example.com"
        }
    ```
    2. Create the related DNS A entries to point to the static ips indicated in the output:
      - `crdb_addresses[*].expected_dns`
      - `gateway_address.expected_dns`

## Next steps

Proceed to [pooling configuration](../pooling/index.md).
