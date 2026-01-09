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

---

## Deployment of the DSS services

Following the successful terraform run, you should find a new [workspace directory](https://github.com/interuss/dss/tree/master/build/workspace)
for the new cluster. The new workspace name corresponds to the cluster context. The cluster context
can be retrieved by running `terraform output` in your infrastructure folder (eg /deploy/infrastructure/personal/terraform-google-dss-dev).

It contains scripts to operate the cluster and setup the services.

1. Go to the new workspace `/build/workspace/${cluster_context}`.
    2. Run `./get-credentials.sh` to login to kubernetes. You can now access the cluster with `kubectl`.

3. Prepare the datastore certificates:
=== "Yugabyte"
    1. Generate the certificates using `./dss-certs.sh init`
    1. If joining a cluster, check `dss-certs.sh`'s [help](../operations/certificates-management.md) to add others CA in your pool and share your CA with others pools members.
    1. Deploy the certificates using `./dss-certs.sh apply`.

=== "CockroachDB"
    1. Generate the certificates using `./make-certs.sh`. Follow script instructions if you are not initializing the cluster.
    1. Deploy the certificates using `./apply-certs.sh`.

---

5. Go to the tanka workspace in `/deploy/services/tanka/workspace/${cluster_context}`.
6. Run `tk apply .` to deploy the services to kubernetes. (This may take up to 30 min)
7. Wait for services to initialize:
    - On Google Cloud, the highest-latency operation is provisioning of the HTTPS certificate which generally takes 10-45 minutes. To track this progress:
        - Go to the "Services & Ingress" left-side tab from the Kubernetes Engine page.
        - Click on the https-ingress item (filter by just the cluster of interest if you have multiple clusters in your project).
        - Under the "Ingress" section for Details, click on the link corresponding with "Load balancer".
        - Under Frontend for Details, the Certificate column for HTTPS protocol will have an icon next to it which will change to a green checkmark when provisioning is complete.
        - Click on the certificate link to see provisioning progress.
        - If everything indicates OK and you still receive a cipher mismatch error message when attempting to visit /healthy, wait an additional 5 minutes before attempting to troubleshoot further.
8. Verify that basic services are functioning by navigating to https://your-gateway-domain.com/healthy.

## Clean up

To delete all resources, run `terraform destroy`. Note that this operation can't be reverted and all data will be lost.

For Google Cloud Engine, make sure to manually clean up the persistent storage: https://console.cloud.google.com/compute/disks
