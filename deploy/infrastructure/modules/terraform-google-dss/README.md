# terraform-google-dss

This terraform module creates a Kubernetes cluster in Google Cloud Engine and generates 
the tanka files to deploy a DSS instance.

## Getting started

### Prerequisites
Download & install the following tools to your workstation:

1. Install [terraform](https://developer.hashicorp.com/terraform/downloads).
2. Install tools from [Prerequisites](../../../../build/README.md)
3. Install provider specific tools:
    1. [Google Cloud Engine](./README.md#google-cloud-engine)
    
#### Google Cloud Engine

1. Install and initialize [Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk).
    1. Confirm successful installation with `gcloud version`.
2. Check that the DSS project is correctly selected: gcloud config list project
    1. Set another one if needed using: `gcloud config set project $GOOGLE_PROJECT_NAME`
3. Enable the following API using [Google Cloud CLI](https://cloud.google.com/endpoints/docs/openapi/enable-api#gcloud):
    1. `compute.googleapis.com`
    2. `container.googleapis.com`
    3. If you want to manage DNS entries with terraform: `dns.googleapis.com`
4. Install the auth plugin to connect to kubernetes: `gcloud components install gke-gcloud-auth-plugin`
5. Run `gcloud auth application-default login` to generate credentials to call Google Cloud Platform APIs.
    1. If the result of performing the authorization indicates 404 in the browser, check whether a local dummy-oauth instance is running (using port 8085).  Stop the dummy-oauth instance if it is running.

### Deployment of the Kubernetes cluster

1. Create a new folder in `/deploy/infrastructure/personal/` named for instance `terraform-google-dss-dev`.
2. Copy main.tf, output.tf and variables.gen.tf to the new folder. (Note that the modules can be added to existing projects)
3. Copy `terraform.dev.example.tfvars` and rename to `terraform.tfvars`
4. Check that your new directory contains the following files:
   - main.tf
   - output.tf
   - terraform.tfvars
   - variables.gen.tf
5. Set the variables in `terraform.tfvars` according to your environment. See [TFVARS.gen.md](TFVARS.gen.md) for variables descriptions.
6. In the new directory (ie /deploy/infrastructure/personal/terraform-google-dss-dev), initialize terraform: `terraform init`.
7. Run `terraform plan` to check that the configuration is valid. It will display the resources which will be provisioned.
8. Run `terraform apply` to deploy the cluster. (This operation may take up to 15 min.)
9. Configure the DNS resolution to the public ip addresses. DNS entries can be either managed manually or 
handled by terraform depending on the cloud provider. See [DNS](DNS.md) for details.

## Deployment of the DSS services

During the successful run, the terraform job has created a new [workspace](../../../../build/workspace/)
for the new cluster. The new workspace name corresponds to the cluster context. The cluster context
can be retrieved by running `terraform output` in your infrastructure folder (ie /deploy/infrastructure/personal/terraform-google-dss-dev).

It contains scripts to operate the cluster and setup the services.

1. Go to the new workspace `/build/workspace/${cluster_context}`.
2. Run `./get-credentials.sh` to login to kubernetes. You can now access the cluster with `kubectl`.
3. Generate the certificates using `./make-certs.sh`. Follow script instructions if you are not initializing the cluster.
4. Deploy the certificates using `./apply-certs.sh`.
5. Run `tk apply .` to deploy the services to kubernetes. (This may take up to 30 min)
6. Wait for services to initialize:
    - On Google Cloud, the highest-latency operation is provisioning of the HTTPS certificate which generally takes 10-45 minutes. To track this progress:
        - Go to the "Services & Ingress" left-side tab from the Kubernetes Engine page.
        - Click on the https-ingress item (filter by just the cluster of interest if you have multiple clusters in your project).
        - Under the "Ingress" section for Details, click on the link corresponding with "Load balancer".
        - Under Frontend for Details, the Certificate column for HTTPS protocol will have an icon next to it which will change to a green checkmark when provisioning is complete.
        - Click on the certificate link to see provisioning progress.
        - If everything indicates OK and you still receive a cipher mismatch error message when attempting to visit /healthy, wait an additional 5 minutes before attempting to troubleshoot further.
7. Verify that basic services are functioning by navigating to https://your-gateway-domain.com/healthy.

## Clean up

To delete all resources, run `terraform destroy`. Note that this operation can't be reverted and all data will be lost.

For Google Cloud Engine, make sure to manually clean up the persistent storage: https://console.cloud.google.com/compute/disks 
